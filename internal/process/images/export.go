package images

import (
	"foxy/internal/utils"
	"github.com/davidbyttow/govips/v2/vips"
	"log"
	"strconv"
)

type ExportOptions struct {
	Format          *string `json:"format"`
	Quality         *int    `json:"quality"`
	ReductionEffort *int    `json:"reductionEffort"`
	Lossless        *bool   `json:"lossless"`
	NearLossless    *bool   `json:"nearLossless"`
}

func (*ExportOptions) Params() []string {
	return []string{"fmt", "q", "nloss", "lossless", "reduction"}
}

func (opt *ExportOptions) ParseParams(param string, options []string) (needsVision bool) {
	needsVision = false

	switch param {
	case "fmt":
		if len(options) == 1 {
			opt.Format = utils.Ptr(options[0])
		}
	case "q":
		if len(options) == 1 {
			q, err := strconv.Atoi(options[0])
			if err == nil {
				opt.Quality = utils.Ptr(q)
			}
		}
	case "nloss":
		if len(options) == 1 {
			nloss := options[0] == "1"

			opt.NearLossless = utils.Ptr(nloss)
		}
	case "lossless":
		if len(options) == 1 {
			loss := options[0] == "1"
			opt.Lossless = utils.Ptr(loss)
		}
	case "reduction":
		if len(options) == 1 {
			e, err := strconv.Atoi(options[0])
			if err == nil {
				opt.ReductionEffort = utils.Ptr(e)
			}
		}
	}

	return
}

func (opt *ExportOptions) ExportPNG(sourceImage *vips.ImageRef) (*[]byte, error) {
	png := vips.NewPngExportParams()
	png.Quality = utils.IfNil(opt.Quality, 85)
	png.StripMetadata = true

	buffer, _, err := sourceImage.ExportPng(png)
	if err != nil {
		log.Println("Export PNG Error:", err)
		return nil, err
	}

	return &buffer, nil
}

func (opt *ExportOptions) ExportJPEG(sourceImage *vips.ImageRef) (*[]byte, error) {
	jpg := vips.NewJpegExportParams()
	jpg.Quality = utils.IfNil(opt.Quality, 85)
	jpg.StripMetadata = true

	buffer, _, err := sourceImage.ExportJpeg(jpg)
	if err != nil {
		log.Println("Export JPEG error:", err)
		return nil, err
	}

	return &buffer, nil
}

func (opt *ExportOptions) ExportAVIF(sourceImage *vips.ImageRef) (*[]byte, error) {
	avif := vips.NewAvifExportParams()
	avif.Quality = utils.IfNil(opt.Quality, 85)
	avif.StripMetadata = true

	buffer, _, err := sourceImage.ExportAvif(avif)
	if err != nil {
		log.Println("Export AVIF error:", err)
		return nil, err
	}

	return &buffer, nil
}

func (opt *ExportOptions) ExportWEBP(sourceImage *vips.ImageRef) (*[]byte, error) {
	webp := vips.NewWebpExportParams()
	webp.Quality = utils.IfNil(opt.Quality, 85)
	webp.StripMetadata = true

	if opt.Lossless != nil {
		webp.Lossless = *opt.Lossless
	}

	if opt.NearLossless != nil {
		webp.NearLossless = *opt.NearLossless
	}

	if opt.ReductionEffort != nil {
		webp.ReductionEffort = *opt.ReductionEffort
	}

	buffer, _, err := sourceImage.ExportWebp(webp)
	if err != nil {
		log.Println("Export WebP Error:", err)
		return nil, err
	}

	return &buffer, nil
}

func (opt *ExportOptions) Export(sourceImage *vips.ImageRef) (*[]byte, error) {
	if opt.Format == nil {
		return opt.ExportJPEG(sourceImage)
	}

	if *opt.Format == "png" {
		return opt.ExportPNG(sourceImage)
	} else if *opt.Format == "webp" {
		return opt.ExportWEBP(sourceImage)
	} else if *opt.Format == "avif" {
		return opt.ExportAVIF(sourceImage)
	} else {
		return opt.ExportJPEG(sourceImage)
	}
}
