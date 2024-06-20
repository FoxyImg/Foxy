package params

import (
	"foxy/internal/config"
	"foxy/internal/utils"
	"foxy/internal/vision"
	"github.com/davidbyttow/govips/v2/vips"
	"math"
	"strconv"
	"time"
)

type SourceCropParams struct {
	X      *int `json:"x,omitempty"`
	Y      *int `json:"y,omitempty"`
	Width  *int `json:"width,omitempty"`
	Height *int `json:"height,omitempty"`
}

func (*SourceCropParams) Params() []string {
	return []string{"src"}
}

func (opts *SourceCropParams) ParseParams(param string, options []string) (needsVision bool) {
	needsVision = false

	if len(options) != 4 {
		return
	}

	x, err := strconv.ParseFloat(options[0], 64)
	if err != nil {
		return
	}

	y, err := strconv.ParseFloat(options[1], 64)
	if err != nil {
		return
	}

	w, err := strconv.ParseFloat(options[2], 64)
	if err != nil {
		return
	}

	h, err := strconv.ParseFloat(options[3], 64)
	if err != nil {
		return
	}

	opts.X = utils.Ptr(int(math.Round(x)))
	opts.Y = utils.Ptr(int(math.Round(y)))
	opts.Width = utils.Ptr(int(math.Round(w)))
	opts.Height = utils.Ptr(int(math.Round(h)))

	return
}

func (opts *SourceCropParams) Process(sourceKey string, sourceId string, config *config.Config, sourceImage *vips.ImageRef, params *ImageParams, imageMeta *vision.Metadata) (*vips.ImageRef, error) {
	if opts.X == nil || opts.Y == nil || opts.Width == nil || opts.Height == nil {
		return sourceImage, nil
	}

	defer utils.TrackTime(time.Now(), "Source Crop")

	sw := float64(sourceImage.Width())
	sh := float64(sourceImage.Height())

	err := sourceImage.Crop(*opts.X, *opts.Y, *opts.Width, *opts.Height)
	if err != nil {
		return nil, err
	}

	if imageMeta != nil && sw > 0 && sh > 0 {
		_ = imageMeta.SourceCrop(sw, sh, float64(*opts.X)/sw, float64(*opts.Y)/sh, float64(*opts.Width)/sw, float64(*opts.Height)/sh)
	}

	if params.Redact != nil && params.Redact.Regions != nil && len(*params.Redact.Regions) > 0 && sw > 0 && sh > 0 {
		_ = params.Redact.SourceCrop(sw, sh, float64(*opts.X)/sw, float64(*opts.Y)/sh, float64(*opts.Width)/sw, float64(*opts.Height)/sh)
	}

	return sourceImage, nil
}
