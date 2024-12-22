package ffmpeg

import (
	"encoding/json"
	"errors"
	"fmt"
	"foxy/internal/aws"
	"foxy/internal/config"
	"foxy/internal/env"
	"foxy/internal/utils"
	securejoin "github.com/cyphar/filepath-securejoin"
	"github.com/davidbyttow/govips/v2/vips"
	"log"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type Ffmpeg struct {
	FfmpegPath  *string
	FfprobePath *string
}

type rawKeyframe struct {
	PtsTime string `json:"pts_time"`
	Flags   string `json:"flags"`
}

type rawPackets struct {
	Packets []rawKeyframe `json:"packets"`
}

func NewFfmpeg(ffmpegPath *string, ffprobePath *string) *Ffmpeg {
	if ffmpegPath == nil {
		ffmpegPath = env.FoxyEnvironment.FfmpegPath
	}

	if ffprobePath == nil {
		ffprobePath = env.FoxyEnvironment.FfprobePath
	}

	return &Ffmpeg{
		FfmpegPath:  ffmpegPath,
		FfprobePath: ffprobePath,
	}
}

func (ffmpeg *Ffmpeg) IsVideo(sourceKey string) bool {
	if ffmpeg.FfprobePath == nil || ffmpeg.FfmpegPath == nil {
		return false
	}

	if !strings.Contains(sourceKey, ".") {
		return false
	}

	ext := strings.ToLower(sourceKey[strings.LastIndex(sourceKey, "."):])
	return ext == ".mp4" || ext == ".mov" || ext == ".m2v" || ext == ".mkv" || ext == ".m4v"
}

func (ffmpeg *Ffmpeg) Probe(config *config.Config, sourceId string, key string) (*Meta, error) {
	if ffmpeg.FfprobePath == nil || ffmpeg.FfmpegPath == nil {
		return nil, errors.New("ffmpeg not set")
	}

	var metaFilePath *string = nil
	var metaFileName *string = nil

	if env.FoxyEnvironment.CacheDir != nil {
		mfp := "/" + sourceId + "/" + strings.TrimLeft(key, "/") + ".probe.json"
		mfn, err := securejoin.SecureJoin(strings.TrimRight(*env.FoxyEnvironment.CacheDir, "/"), mfp)
		if err != nil {
			return nil, err
		}

		metaFilePath = &mfp
		metaFileName = &mfn
	}

	if env.FoxyEnvironment.CacheDir != nil && metaFilePath != nil && metaFileName != nil {
		_, err := os.Stat(*metaFileName)
		if err == nil {
			jsonData, err := os.ReadFile(*metaFileName)
			if err == nil {
				meta := Meta{}
				jsonErr := json.Unmarshal(jsonData, &meta)
				if jsonErr == nil {
					log.Println("Metadata cache hit")
					return &meta, nil
				} else {
					log.Println("Metadata json parse error", jsonErr)
				}
			} else {
				log.Println("Error reading metadata json", err)
			}
		}
	} else {
		log.Println("Skipping meta cache")
	}

	var pathOrUrl string
	var err error
	if config.Source.Type == "s3" {
		pathOrUrl, err = aws.GetSignedUrl(*config, key, time.Hour*1)
		if err != nil {
			return nil, err
		}
	} else if config.Source.Type == "web" {
		pathOrUrl = *config.Source.WebConfig.Url + "/" + key
	} else if config.Source.Type == "local" {
		pathOrUrl, err = securejoin.SecureJoin(*config.Source.LocalConfig.Path, "/"+key)
		if err != nil {
			return nil, err
		}
	}

	probeCmd := exec.Command(*ffmpeg.FfprobePath, "-v", "quiet", "-print_format", "json", "-show_format", "-show_streams", pathOrUrl)
	log.Println("Probe Command:", *ffmpeg.FfprobePath, "-v", "quiet", "-print_format", "json", "-show_format", "-show_streams", pathOrUrl)
	probeStdOut, err := probeCmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err = probeCmd.Start(); err != nil {
		return nil, err
	}

	var meta Meta
	if err = json.NewDecoder(probeStdOut).Decode(&meta); err != nil {
		return nil, err
	}

	keyframesCmd := exec.Command(*ffmpeg.FfprobePath, "-loglevel", "error", "-select_streams", "v:0", "-show_entries", "packet=pts_time,flags", "-of", "json", pathOrUrl)
	log.Println("Keyframes Command:", *ffmpeg.FfprobePath, "-loglevel", "error", "-select_streams", "v:0", "-show_entries", "packet=pts_time,flags", "-of", "json", pathOrUrl)
	keyframesStdOut, err := keyframesCmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err = keyframesCmd.Start(); err != nil {
		return nil, err
	}

	var packets rawPackets
	if err = json.NewDecoder(keyframesStdOut).Decode(&packets); err != nil {
		return nil, err
	}

	for _, packet := range packets.Packets {
		if strings.Index(packet.Flags, "K") == 0 {
			meta.Keyframes = append(meta.Keyframes, packet.PtsTime)
		}
	}

	if env.FoxyEnvironment.CacheDir != nil && metaFilePath != nil && metaFileName != nil {
		metaFilePath := filepath.Dir(*metaFileName)
		err := os.MkdirAll(metaFilePath, os.ModePerm)
		if err != nil {
			log.Println("MkdirAll Error:", err)
			return &meta, nil
		}

		jsonData, err := json.Marshal(meta)
		if err != nil {
			log.Println("Marshal Error:", err)
			return &meta, nil
		}

		_ = os.WriteFile(*metaFileName, jsonData, 0644)
	}

	return &meta, nil
}

func (ffmpeg *Ffmpeg) PrepareFrameExtraction(config *config.Config, sourceId string, key string) (*Meta, *string, error) {
	if ffmpeg.FfprobePath == nil || ffmpeg.FfmpegPath == nil {
		return nil, nil, errors.New("ffmpeg not set")
	}
	if env.FoxyEnvironment.CacheDir == nil {
		return nil, nil, errors.New("no cache dir set")
	}

	meta, err := ffmpeg.Probe(config, sourceId, key)
	if err != nil {
		return nil, nil, err
	}

	var pathOrUrl string
	if config.Source.Type == "s3" {
		pathOrUrl, err = aws.GetSignedUrl(*config, key, time.Hour*1)
		if err != nil {
			return nil, nil, err
		}
	} else if config.Source.Type == "web" {
		pathOrUrl = *config.Source.WebConfig.Url + "/" + key
	} else if config.Source.Type == "local" {
		pathOrUrl, err = securejoin.SecureJoin(*config.Source.LocalConfig.Path, "/"+key)
		if err != nil {
			return nil, nil, err
		}
	}

	return meta, &pathOrUrl, nil
}

func (ffmpeg *Ffmpeg) ExtractFrameAtRelativeTime(config *config.Config, sourceId string, key string, framePercent float64) (*Meta, string, *vips.ImageRef, error) {
	meta, pathOrUrl, err := ffmpeg.PrepareFrameExtraction(config, sourceId, key)
	if err != nil {
		log.Println("ExtractFrameAtRelativeTime", err)
		return nil, key, nil, err
	}

	newKey := key + fmt.Sprintf(".frame-%d-rel", int(math.Round(framePercent*100))) + ".png"
	cachedSourceFilePath := "/" + sourceId + "/" + strings.TrimLeft(newKey, "/")
	cachedSourceFileName, err := securejoin.SecureJoin(strings.TrimRight(*env.FoxyEnvironment.CacheDir, "/"), cachedSourceFilePath)
	if err != nil {
		log.Println("ExtractFrameAtRelativeTime Err", err)
		return nil, key, nil, err
	}

	_, err = os.Stat(cachedSourceFileName)
	if err == nil {
		log.Println("Found cached source file", cachedSourceFileName)
		img, err := vips.NewImageFromFile(cachedSourceFileName)
		if err == nil {
			return meta, newKey, img, nil
		}

		log.Println("Error reading cached source file", err)
	}

	if len(meta.Keyframes) == 0 {
		return nil, key, nil, errors.New("no keyframes found")
	}

	frame := int(math.Round(float64(len(meta.Keyframes)) * framePercent))

	extractFrame := exec.Command(*ffmpeg.FfmpegPath, "-ss", meta.Keyframes[frame], "-i", *pathOrUrl, "-frames", "1", cachedSourceFileName)
	log.Println("Extract Command:", *ffmpeg.FfmpegPath, "-ss", meta.Keyframes[frame], "-i", *pathOrUrl, "-frames", "1", cachedSourceFileName)
	if err = extractFrame.Run(); err != nil {
		return nil, key, nil, err

	}
	_, err = os.Stat(cachedSourceFileName)
	if err == nil {
		img, err := vips.NewImageFromFile(cachedSourceFileName)
		if err == nil {
			return meta, newKey, img, nil
		}
	}

	return nil, key, nil, errors.New("ffmpeg failed to extract frame")
}

func (ffmpeg *Ffmpeg) ExtractFrameAtKeyframe(config *config.Config, sourceId string, key string, keyFrame int) (*Meta, string, *vips.ImageRef, error) {
	meta, pathOrUrl, err := ffmpeg.PrepareFrameExtraction(config, sourceId, key)
	if err != nil {
		return nil, key, nil, err
	}

	newKey := key + fmt.Sprintf(".frame-%d-keyframe", keyFrame) + ".png"
	cachedSourceFilePath := "/" + sourceId + "/" + strings.TrimLeft(newKey, "/")
	cachedSourceFileName, err := securejoin.SecureJoin(strings.TrimRight(*env.FoxyEnvironment.CacheDir, "/"), cachedSourceFilePath)
	if err != nil {
		return nil, key, nil, err
	}

	_, err = os.Stat(cachedSourceFileName)
	if err == nil {
		img, err := vips.NewImageFromFile(cachedSourceFileName)
		if err == nil {
			return meta, newKey, img, nil
		}

		log.Println("Error reading cached source file", err)
	}

	if len(meta.Keyframes) == 0 {
		return nil, key, nil, errors.New("no keyframes found")
	}

	frame := utils.Max(0, utils.Min(len(meta.Keyframes)-1, keyFrame))

	extractFrame := exec.Command(*ffmpeg.FfmpegPath, "-ss", meta.Keyframes[frame], "-i", *pathOrUrl, "-frames", "1", cachedSourceFileName)
	if err = extractFrame.Run(); err != nil {
		return nil, key, nil, err

	}
	_, err = os.Stat(cachedSourceFileName)
	if err == nil {
		img, err := vips.NewImageFromFile(cachedSourceFileName)
		if err == nil {
			return meta, newKey, img, nil
		}
	}

	return nil, key, nil, errors.New("ffmpeg failed to extract frame")
}

func (ffmpeg *Ffmpeg) ExtractFrameAtFrame(config *config.Config, sourceId string, key string, frame int) (*Meta, string, *vips.ImageRef, error) {
	meta, pathOrUrl, err := ffmpeg.PrepareFrameExtraction(config, sourceId, key)
	if err != nil {
		return nil, key, nil, err
	}

	newKey := key + fmt.Sprintf(".frame-%d-frame", frame) + ".png"
	cachedSourceFilePath := "/" + sourceId + "/" + strings.TrimLeft(newKey, "/")
	cachedSourceFileName, err := securejoin.SecureJoin(strings.TrimRight(*env.FoxyEnvironment.CacheDir, "/"), cachedSourceFilePath)
	if err != nil {
		return nil, key, nil, err
	}

	_, err = os.Stat(cachedSourceFileName)
	if err == nil {
		img, err := vips.NewImageFromFile(cachedSourceFileName)
		if err == nil {
			return meta, newKey, img, nil
		}

		log.Println("Error reading cached source file", err)
	}

	videoMeta := meta.GetVideoMetadata()
	if videoMeta == nil {
		return nil, key, nil, errors.New("no video metadata found")
	}

	timeOffset := videoMeta.Duration * (float64(frame) / float64(videoMeta.FrameCount))

	extractFrame := exec.Command(*ffmpeg.FfmpegPath, "-ss", fmt.Sprintf("%f", timeOffset), "-i", *pathOrUrl, "-frames", "1", cachedSourceFileName)
	if err = extractFrame.Run(); err != nil {
		return nil, key, nil, err

	}
	_, err = os.Stat(cachedSourceFileName)
	if err == nil {
		img, err := vips.NewImageFromFile(cachedSourceFileName)
		if err == nil {
			return meta, newKey, img, nil
		}
	}

	return nil, key, nil, errors.New("ffmpeg failed to extract frame")
}

func (ffmpeg *Ffmpeg) ExtractFrameAtTime(config *config.Config, sourceId string, key string, time float64) (*Meta, string, *vips.ImageRef, error) {
	meta, pathOrUrl, err := ffmpeg.PrepareFrameExtraction(config, sourceId, key)
	if err != nil {
		return nil, key, nil, err
	}

	newKey := key + fmt.Sprintf(".frame-%f-time", time) + ".png"
	cachedSourceFilePath := "/" + sourceId + "/" + strings.TrimLeft(newKey, "/")
	cachedSourceFileName, err := securejoin.SecureJoin(strings.TrimRight(*env.FoxyEnvironment.CacheDir, "/"), cachedSourceFilePath)
	if err != nil {
		return nil, key, nil, err
	}

	_, err = os.Stat(cachedSourceFileName)
	if err == nil {
		img, err := vips.NewImageFromFile(cachedSourceFileName)
		if err == nil {
			return meta, newKey, img, nil
		}

		log.Println("Error reading cached source file", err)
	}

	videoMeta := meta.GetVideoMetadata()
	if videoMeta == nil {
		return nil, key, nil, errors.New("no video metadata found")
	}

	time = math.Max(0, math.Min(videoMeta.Duration, time))

	extractFrame := exec.Command(*ffmpeg.FfmpegPath, "-ss", fmt.Sprintf("%f", time), "-i", *pathOrUrl, "-frames", "1", cachedSourceFileName)
	if err = extractFrame.Run(); err != nil {
		return nil, key, nil, err

	}
	_, err = os.Stat(cachedSourceFileName)
	if err == nil {
		img, err := vips.NewImageFromFile(cachedSourceFileName)
		if err == nil {
			return meta, newKey, img, nil
		}
	}

	return nil, key, nil, errors.New("ffmpeg failed to extract frame")
}

func (ffmpeg *Ffmpeg) GenerateStoryboard(config *config.Config, sourceId string, key string, cols int, rows int, largestDimension int) (*Meta, string, *vips.ImageRef, error) {
	meta, pathOrUrl, err := ffmpeg.PrepareFrameExtraction(config, sourceId, key)
	if err != nil {
		return nil, key, nil, err
	}

	newKey := key + fmt.Sprintf(".frame-%dx%d-%d-storboard", cols, rows, largestDimension) + ".png"
	cachedSourceFilePath := "/" + sourceId + "/" + strings.TrimLeft(newKey, "/")
	cachedSourceFileName, err := securejoin.SecureJoin(strings.TrimRight(*env.FoxyEnvironment.CacheDir, "/"), cachedSourceFilePath)
	if err != nil {
		return nil, key, nil, err
	}

	_, err = os.Stat(cachedSourceFileName)
	if err == nil {
		img, err := vips.NewImageFromFile(cachedSourceFileName)
		if err == nil {
			return meta, newKey, img, nil
		}

		log.Println("Error reading cached source file", err)
	}

	videoMeta := meta.GetVideoMetadata()
	if videoMeta == nil {
		return nil, key, nil, errors.New("no video metadata found")
	}

	targetWidth := 0
	targetHeight := 0

	if videoMeta.Width > videoMeta.Height {
		targetWidth = largestDimension
		targetHeight = int(float64(videoMeta.Height) * float64(largestDimension) / float64(videoMeta.Width))
	} else if videoMeta.Width < videoMeta.Height {
		targetWidth = int(float64(videoMeta.Width) * float64(largestDimension) / float64(videoMeta.Height))
		targetHeight = largestDimension
	} else {
		targetWidth = largestDimension
		targetHeight = largestDimension
	}

	frameSkip := videoMeta.FrameCount / ((rows * cols) - 1)

	storyboard := exec.Command(*ffmpeg.FfmpegPath, "-y", "-i", *pathOrUrl, "-vf", fmt.Sprintf("select='not(mod(n\\,%d))',scale=%d:%d,tile=%dx%d", frameSkip, targetWidth, targetHeight, cols, rows), "-frames:v", "1", cachedSourceFileName)
	if err = storyboard.Run(); err != nil {
		return nil, key, nil, err

	}
	_, err = os.Stat(cachedSourceFileName)
	if err == nil {
		img, err := vips.NewImageFromFile(cachedSourceFileName)
		if err == nil {
			return meta, newKey, img, nil
		}
	}

	return nil, key, nil, errors.New("ffmpeg failed to generate storyboard")
}
