package params

import (
	"fmt"
	"foxy/internal/config"
	"foxy/internal/vision"
	"github.com/davidbyttow/govips/v2/vips"
)

type ChannelLevelsParams struct {
	Shadows    *int `json:"shadows,omitempty"`
	MidTones   *int `json:"midtones,omitempty"`
	Highlights *int `json:"highlights,omitempty"`
}

type LevelsParams struct {
	All   *ChannelLevelsParams `json:"all,omitempty"`
	Red   *ChannelLevelsParams `json:"red,omitempty"`
	Green *ChannelLevelsParams `json:"green,omitempty"`
	Blue  *ChannelLevelsParams `json:"blue,omitempty"`
}

func (opts *ChannelLevelsParams) GenerateLevelsMap() (*vips.ImageRef, error) {
	stepsSvg := fmt.Sprintf("<stop offset='%d%%' stop-color='%s'/>", int((float64(*opts.Shadows)/255.0)*100.0), "rgb(0, 0, 0)")
	stepsSvg += fmt.Sprintf("<stop offset='%d%%' stop-color='%s'/>", int((float64(*opts.MidTones)/255.0)*100.0), "rgb(127, 127, 127)")
	stepsSvg += fmt.Sprintf("<stop offset='%d%%' stop-color='%s'/>", int((float64(*opts.Highlights)/255.0)*100.0), "rgb(255,255,255)")

	mapSvg := fmt.Sprintf("<svg width='256' height='1' xmlns='http://www.w3.org/2000/svg'><defs><linearGradient id='grad' x1='0%%' y1='0%%' x2='100%%' y2='0%%'>%s</linearGradient></defs><rect width='256' height='1' fill='url(#grad)'/></svg>", stepsSvg)
	mapImg, mapImgErr := vips.LoadImageFromBuffer([]byte(mapSvg), vips.NewImportParams())
	if mapImgErr != nil {
		return nil, mapImgErr
	}

	return mapImg, nil
}

func (*LevelsParams) Params() []string {
	return []string{"l"}
}

func (opts *LevelsParams) ParseParams(param string, options []string) (needsVision bool) {
	needsVision = false

	if len(options) != 4 {
		return
	}

	switch options[0] {
	case "all":
		opts.All = &ChannelLevelsParams{
			Shadows:    IntVal(options[1:]),
			MidTones:   IntVal(options[2:]),
			Highlights: IntVal(options[3:]),
		}
	case "r":
		opts.Red = &ChannelLevelsParams{
			Shadows:    IntVal(options[1:]),
			MidTones:   IntVal(options[2:]),
			Highlights: IntVal(options[3:]),
		}
	case "g":
		opts.Green = &ChannelLevelsParams{
			Shadows:    IntVal(options[1:]),
			MidTones:   IntVal(options[2:]),
			Highlights: IntVal(options[3:]),
		}
	case "b":
		opts.Blue = &ChannelLevelsParams{
			Shadows:    IntVal(options[1:]),
			MidTones:   IntVal(options[2:]),
			Highlights: IntVal(options[3:]),
		}
	}

	return
}

func (opts *LevelsParams) Process(sourceKey string, sourceId string, config *config.Config, sourceImage *vips.ImageRef, params *ImageParams, imageMeta *vision.Metadata) (*vips.ImageRef, error) {
	if opts.All == nil && opts.Red == nil && opts.Green == nil && opts.Blue == nil {
		return sourceImage, nil
	}

	redChannel, err := sourceImage.ExtractBandToImage(0, 1)
	if err != nil {
		return nil, err
	}

	var greenChannel *vips.ImageRef
	if sourceImage.Bands() > 1 {
		greenChannel, err = sourceImage.ExtractBandToImage(1, 1)
		if err != nil {
			return nil, err
		}
	}

	var blueChannel *vips.ImageRef
	if sourceImage.Bands() > 2 {
		blueChannel, err = sourceImage.ExtractBandToImage(2, 1)
		if err != nil {
			return nil, err
		}
	}

	var alphaChannel *vips.ImageRef
	if sourceImage.Bands() > 3 {
		alphaChannel, err = sourceImage.ExtractBandToImage(3, 1)
		if err != nil {
			return nil, err
		}
	}

	var allGradientMap *vips.ImageRef
	if opts.All != nil {
		allGradientMap, err = opts.All.GenerateLevelsMap()
		if err != nil {
			return nil, err
		}
	}

	var redGradientMap *vips.ImageRef
	if opts.Red != nil {
		redGradientMap, err = opts.Red.GenerateLevelsMap()
		if err != nil {
			return nil, err
		}

		_ = redChannel.Maplut(redGradientMap)
		redChannel, _ = redChannel.ExtractBandToImage(0, 1)
	}

	if allGradientMap != nil {
		_ = redChannel.Maplut(allGradientMap)
		redChannel, _ = redChannel.ExtractBandToImage(0, 1)
	}

	if greenChannel != nil {
		var greenGradientMap *vips.ImageRef
		if opts.Green != nil {
			greenGradientMap, err = opts.Green.GenerateLevelsMap()
			if err != nil {
				return nil, err
			}

			_ = greenChannel.Maplut(greenGradientMap)
			greenChannel, _ = greenChannel.ExtractBandToImage(0, 1)
		}

		if allGradientMap != nil {
			_ = greenChannel.Maplut(allGradientMap)
			greenChannel, _ = greenChannel.ExtractBandToImage(0, 1)
		}
	}

	if blueChannel != nil {
		var blueGradientMap *vips.ImageRef
		if opts.Blue != nil {
			blueGradientMap, err = opts.Blue.GenerateLevelsMap()
			if err != nil {
				return nil, err
			}

			_ = blueChannel.Maplut(blueGradientMap)
			blueChannel, _ = blueChannel.ExtractBandToImage(0, 1)
		}

		if allGradientMap != nil {
			_ = blueChannel.Maplut(allGradientMap)
			blueChannel, _ = blueChannel.ExtractBandToImage(0, 1)
		}
	}

	_ = redChannel.BandJoin(greenChannel, blueChannel)

	if alphaChannel != nil {
		_ = redChannel.BandJoin(alphaChannel)
	}

	return redChannel, nil
}
