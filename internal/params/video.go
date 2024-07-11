package params

import (
	"foxy/internal/config"
	"foxy/internal/ffmpeg"
	"foxy/internal/utils"
	"foxy/internal/vision"
	"github.com/davidbyttow/govips/v2/vips"
	"log"
	"strconv"
	"time"
)

type VideoParams struct {
	Type             *string  `json:"type,omitempty"`
	FrameType        *string  `json:"frameType,omitempty"`
	Frame            *int     `json:"frame,omitempty"`
	Keyframe         *int     `json:"keyframe,omitempty"`
	Time             *float64 `json:"time,omitempty"`
	RelativeTime     *float64 `json:"relativeTime,omitempty"`
	Cols             *int     `json:"cols,omitempty"`
	Rows             *int     `json:"rows,omitempty"`
	LargestDimension *int     `json:"largestDimension,omitempty"`
}

func (*VideoParams) Params() []string {
	return []string{"video"}
}

// video:frame:time:123.223
// video:frame:frame:123
// video:frame:rel:50
// video:sb:10:10:256
func (opts *VideoParams) ParseParams(param string, options []string) (needsVision bool) {
	needsVision = false

	switch options[0] {
	case "frame":
		if len(options) < 3 {
			return
		}

		opts.Type = utils.Ptr("frame")
		opts.FrameType = utils.Ptr(options[1])
		if *opts.FrameType == "time" {
			b, err := strconv.ParseFloat(options[2], 64)
			if err == nil {
				opts.Time = utils.Ptr(b)
			}
		} else if *opts.FrameType == "frame" {
			b, err := strconv.Atoi(options[2])
			if err == nil {
				opts.Frame = utils.Ptr(b)
			}
		} else if *opts.FrameType == "rel" {
			b, err := strconv.Atoi(options[2])
			if err == nil {
				opts.RelativeTime = utils.Ptr(float64(b) / 100.0)
			}
		} else if *opts.FrameType == "key" {
			b, err := strconv.Atoi(options[2])
			if err == nil {
				opts.Keyframe = utils.Ptr(b)
			}
		}
	case "sb":
		if len(options) < 4 {
			return
		}

		cols, err := strconv.Atoi(options[1])
		if err != nil {
			return
		}

		rows, err := strconv.Atoi(options[2])
		if err != nil {
			return
		}

		largestDimension, err := strconv.Atoi(options[3])
		if err != nil {
			return
		}

		if cols < 1 || rows < 1 || largestDimension < 48 {
			return
		}

		opts.Type = utils.Ptr("storyboard")
		opts.Cols = utils.Ptr(cols)
		opts.Rows = utils.Ptr(rows)
		opts.LargestDimension = utils.Ptr(largestDimension)
	}

	return needsVision
}

func (opts *VideoParams) Process(sourceKey string, sourceId string, config *config.Config, sourceImage *vips.ImageRef, params *ImageParams, imageMeta *vision.Metadata) (*vips.ImageRef, error) {
	return sourceImage, nil
}

func (opts *VideoParams) ProcessFrame(sourceKey string, sourceId string, config *config.Config, sourceImage *vips.ImageRef, params *ImageParams, imageMeta *vision.Metadata) (*ffmpeg.Meta, string, *vips.ImageRef, error) {
	defer utils.TrackTime(time.Now(), "Process Video Frame")

	if opts.Type == nil {
		log.Print("Video type is nil")
		return nil, sourceKey, sourceImage, nil
	}

	ffmpegUtility := ffmpeg.NewFfmpeg(nil, nil)
	var err error
	var ffmeta *ffmpeg.Meta

	if *opts.Type == "storyboard" {
		if opts.Cols != nil && *opts.Cols > 0 && opts.Rows != nil && *opts.Rows > 0 && opts.LargestDimension != nil && *opts.LargestDimension >= 48 {
			ffmeta, sourceKey, sourceImage, err = ffmpegUtility.GenerateStoryboard(config, sourceId, sourceKey, *opts.Cols, *opts.Rows, *opts.LargestDimension)
		}
	} else {
		if *opts.FrameType == "rel" && opts.RelativeTime != nil {
			log.Print("ExtractFrameAtRelativeTime")
			ffmeta, sourceKey, sourceImage, err = ffmpegUtility.ExtractFrameAtRelativeTime(config, sourceId, sourceKey, *opts.RelativeTime)
		} else if *opts.FrameType == "key" && opts.Keyframe != nil {
			ffmeta, sourceKey, sourceImage, err = ffmpegUtility.ExtractFrameAtKeyframe(config, sourceId, sourceKey, *opts.Keyframe)
		} else if *opts.FrameType == "time" && opts.Time != nil {
			ffmeta, sourceKey, sourceImage, err = ffmpegUtility.ExtractFrameAtTime(config, sourceId, sourceKey, *opts.Time)
		} else if *opts.FrameType == "frame" && opts.Frame != nil {
			ffmeta, sourceKey, sourceImage, err = ffmpegUtility.ExtractFrameAtFrame(config, sourceId, sourceKey, *opts.Frame)
		}
	}

	return ffmeta, sourceKey, sourceImage, err
}
