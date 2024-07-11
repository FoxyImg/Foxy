package cli

import (
	"encoding/json"
	"foxy/internal/config"
	"foxy/internal/ffmpeg"
	"foxy/internal/params"
	"foxy/internal/storage"
	"foxy/internal/utils"
	"foxy/internal/vision"
	"github.com/davidbyttow/govips/v2/vips"
	"log"
	"os"
)

type Context struct {
}

type ProcessCmd struct {
	SourceId       string `arg:"" arg:"" help:"The source id"`
	Source         string `arg:"" arg:"" help:"The source key"`
	Key            string `arg:"" arg:"" help:"The key"`
	ParamsJSON     string `arg:"" arg:"" help:"The path to the params json"`
	OutputFile     string `arg:"" arg:"" help:"The path to the output file"`
	MetaOutputFile string `arg:"" arg:"" help:"The path to the meta output file"`
}

var CLI struct {
	Process ProcessCmd `cmd:"" help:"Process an image"`
	Server  ServerCmd  `cmd:"" help:"Start the server"`
}

func (cmd *ProcessCmd) Run(ctx *Context) error {
	_, err := os.Stat(cmd.ParamsJSON)
	if err != nil {
		return err
	}

	paramsJSON, err := os.ReadFile(cmd.ParamsJSON)
	if err != nil {
		return err
	}

	imageParams := params.NewImageParams()
	err = json.Unmarshal(paramsJSON, imageParams)

	sourceConfig, err := config.GetSourceConfigFromCache(cmd.SourceId)
	if err != nil {
		return err
	}

	if imageParams.Debug != nil && !imageParams.Debug.DisableRenderCache {
		cached, _ := storage.GetCachedResult(sourceConfig, cmd.SourceId, cmd.Source, imageParams, utils.IfNil(imageParams.Export.Format, "jpg"))
		if cached != nil {
			_, err = os.Stat(cmd.OutputFile)
			if err == nil {
				_ = os.Remove(cmd.OutputFile)
			}

			data, readErr := os.ReadFile(cmd.OutputFile)
			if readErr == nil {
				return readErr
			}

			_ = os.WriteFile(cmd.OutputFile, data, 0644)
			return nil
		}
	}

	ffmpegUtility := ffmpeg.NewFfmpeg(nil, nil)
	isVideo := ffmpegUtility.IsVideo(cmd.Key)

	key := cmd.Key

	var img *vips.ImageRef
	var videoMeta *vision.VideoMetadata
	if isVideo {
		var ffmeta *ffmpeg.Meta
		ffmeta, key, img, err = imageParams.Video.ProcessFrame(key, cmd.SourceId, sourceConfig, nil, imageParams, nil)
		if err != nil {
			log.Println("ProcessFrame Error:", err)
			return err
		}
		if ffmeta != nil {
			videoMeta = ffmeta.GetVideoMetadata()
		}
	} else {
		log.Println("Get Source Image", key)
		img, err = storage.GetSourceImage(sourceConfig, cmd.SourceId, key, imageParams.Debug != nil && imageParams.Debug.DisableSourceCache)
		if err != nil {
			return err
		}
	}

	buffer, meta, err := params.ProcessImage(cmd.SourceId, sourceConfig, cmd.SourceId, key, imageParams, img)
	if err != nil {
		return err
	}

	if meta != nil {
		if videoMeta != nil {
			meta.Video = videoMeta
		}

		_, err = os.Stat(cmd.MetaOutputFile)
		if err == nil {
			_ = os.Remove(cmd.MetaOutputFile)
		}

		metaJSON, metaJSONErr := json.MarshalIndent(meta, "", "  ")
		if metaJSONErr == nil {
			err = os.WriteFile(cmd.MetaOutputFile, metaJSON, 0644)
			if err != nil {
				log.Println("Write File Error:", err)
			}
		} else {
			log.Println("Marshal JSON Error: ", metaJSONErr)
		}
	}

	if imageParams.MetaOnly {
		return nil
	}

	log.Println("Set Cached Result", key)
	_ = storage.SetCachedResult(sourceConfig, cmd.SourceId, cmd.Key, imageParams, utils.IfNil(imageParams.Export.Format, "jpg"), buffer)

	_, err = os.Stat(cmd.OutputFile)
	if err == nil {
		_ = os.Remove(cmd.OutputFile)
	}

	return os.WriteFile(cmd.OutputFile, *buffer, 0644)
}
