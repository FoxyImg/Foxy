package params

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"foxy/internal/config"
	"foxy/internal/env"
	"foxy/internal/geometry"
	"foxy/internal/storage"
	"foxy/internal/utils"
	"foxy/internal/vision"
	"github.com/davidbyttow/govips/v2/vips"
	"math"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type OverlaySize struct {
	RelativeSize *bool `json:"relativeSize,omitempty"`
	Width        *int  `json:"width,omitempty"`
	Height       *int  `json:"height,omitempty"`
	MinWidth     *int  `json:"minWidth,omitempty"`
	MinHeight    *int  `json:"minHeight,omitempty"`
	MaxWidth     *int  `json:"maxWidth,omitempty"`
	MaxHeight    *int  `json:"maxHeight,omitempty"`
}

type OverlayDropShadowParams struct {
	Opacity *float64 `json:"opacity,omitempty"`
	Blur    *float64 `json:"blur,omitempty"`
	Color   *string  `json:"color,omitempty"`
	OffsetX *int     `json:"offsetX,omitempty"`
	OffsetY *int     `json:"offsetY,omitempty"`
}

type OverlayBackgroundParams struct {
	OverlaySize
	BackgroundColor      *string  `json:"textColor,omitempty"`
	BackgroundColorType  *string  `json:"colorType,omitempty"`
	DominantColorOpacity *float64 `json:"dominantColorOpacity,omitempty"`
	Blur                 *int     `json:"blur,omitempty"`
	Saturation           *float64 `json:"saturation,omitempty"`
	Contrast             *float64 `json:"contrast,omitempty"`
	Brightness           *float64 `json:"brightness,omitempty"`
	CornerRadius         *float64 `json:"radius,omitempty"`
	RelativePadding      *bool    `json:"relativeCoords,omitempty"`
	HPadding             *int     `json:"hPadding,omitempty"`
	VPadding             *int     `json:"vPadding,omitempty"`
	HAlign               *string  `json:"hAlign,omitempty"`
	VAlign               *string  `json:"vAlign,omitempty"`
}

type OverlayParams struct {
	OverlaySize
	Type           string                   `json:"type"`
	Text           *string                  `json:"text,omitempty"`
	Font           *string                  `json:"font,omitempty"`
	ImageKey       *string                  `json:"url,omitempty"`
	Trim           *bool                    `json:"trim,omitempty"`
	Opacity        *float64                 `json:"opacity,omitempty"`
	Rotate         *int                     `json:"rotate,omitempty"`
	RelativeCoords *bool                    `json:"relativeCoords,omitempty"`
	HPadding       *int                     `json:"hPadding,omitempty"`
	VPadding       *int                     `json:"vPadding,omitempty"`
	X              *int                     `json:"x,omitempty"`
	Y              *int                     `json:"y,omitempty"`
	Fit            *string                  `json:"fit,omitempty"`
	HAnchor        *string                  `json:"hAnchor,omitempty"`
	VAnchor        *string                  `json:"vAnchor,omitempty"`
	TextColor      *string                  `json:"textColor,omitempty"`
	Substitutes    map[string]string        `json:"substitutes,omitempty"`
	DropShadow     *OverlayDropShadowParams `json:"dropShadow,omitempty"`
	Background     *OverlayBackgroundParams `json:"background,omitempty"`
}

type Overlays map[int]*OverlayParams

func (Overlays) Params() []string {
	return []string{"ov", "ovs"}
}

func (o *OverlayDropShadowParams) ParseDropShadowParams(param string, options []string) {
	switch param {
	case "o":
		o.Opacity = NormalizedFloatVal(options)
	case "c":
		o.Color = &options[0]
	case "bl":
		o.Blur = FloatVal(options)
	case "xy":
		o.OffsetX, o.OffsetY = Int2DVectorVal(options)
	}
}

func (o *OverlayBackgroundParams) ParseBackgroundParams(param string, options []string) {
	switch param {
	case "sz":
		o.RelativeSize, o.Width, o.Height = FlexibleInt2DVectorVal(options)
	case "minsz":
		o.MinWidth, o.MinHeight = Int2DVectorVal(options)
	case "maxsz":
		o.MaxWidth, o.MaxHeight = Int2DVectorVal(options)
	case "c":
		if len(options) == 1 {
			o.BackgroundColor = &options[0]
		} else if len(options) == 3 {
			o.BackgroundColorType = &options[0]
			o.DominantColorOpacity = NormalizedFloatVal(options[1:])
			o.BackgroundColor = &options[2]
		}
	case "bl":
		o.Blur = IntVal(options)
	case "sat":
		o.Saturation = NormalizedFloatVal(options)
	case "con":
		o.Contrast = FloatVal(options)
	case "bri":
		o.Brightness = NormalizedFloatVal(options)
	case "pad":
		o.RelativePadding, o.HPadding, o.VPadding = FlexibleInt2DVectorVal(options)
	case "align":
		if len(options) == 2 {
			o.HAlign = &options[0]
			o.VAlign = &options[1]
		}
	case "br":
		o.CornerRadius = FloatVal(options)
	}
}

func (o *OverlayParams) ParseParams(param string, options []string) (needsVision bool) {
	needsVision = false

	if len(options) == 0 {
		return
	}

	switch options[0] {
	case "url":
		o.Type = "image"
		o.ImageKey = DecodeBase64StringVal(options[1:])
	case "text":
		o.Type = "text"
		o.Text = DecodeBase64StringVal(options[1:])
	case "font":
		o.Font = DecodeBase64StringVal(options[1:])
	case "xy":
		o.RelativeCoords, o.X, o.Y = FlexibleInt2DVectorVal(options[1:])
	case "trim":
		o.Trim = utils.Ptr(true)
	case "a":
		if len(options) == 3 {
			o.HAnchor = &options[1]
			o.VAnchor = &options[2]
		}
	case "sub":
		if len(options) == 3 {
			if o.Substitutes == nil {
				o.Substitutes = make(map[string]string)
			}

			o.Substitutes[*DecodeBase64StringVal(options[1:])] = *DecodeBase64StringVal(options[2:])
		}
	case "pad":
		o.HPadding, o.VPadding = Int2DVectorVal(options[1:])
	case "sz":
		o.RelativeSize, o.Width, o.Height = FlexibleInt2DVectorVal(options[1:])
	case "minsz":
		o.MinWidth, o.MinHeight = Int2DVectorVal(options[1:])
	case "maxsz":
		o.MaxWidth, o.MaxHeight = Int2DVectorVal(options[1:])
	case "rot":
		o.Rotate = IntVal(options[1:])
	case "o":
		o.Opacity = NormalizedFloatVal(options[1:])
	case "fit":
		if len(options) == 2 {
			o.Fit = &options[1]
		}
	case "tc":
		o.TextColor = &options[1]
	case "ds":
		if len(options) > 2 {
			if o.DropShadow == nil {
				o.DropShadow = &OverlayDropShadowParams{}
			}

			o.DropShadow.ParseDropShadowParams(options[1], options[2:])
		}
	case "bg":
		if len(options) > 2 {
			if o.Background == nil {
				o.Background = &OverlayBackgroundParams{}
			}

			o.Background.ParseBackgroundParams(options[1], options[2:])
		}
	}

	return
}

func (overlays Overlays) ParseParams(param string, options []string) (needsVision bool) {
	needsVision = false
	if param == "ovs" {
		var tempOverlays []*OverlayParams
		overlayJSON, err := base64.URLEncoding.DecodeString(options[1])
		if err != nil {
			return
		}

		err = json.Unmarshal(overlayJSON, &tempOverlays)
		if err != nil {
			return
		}

		for idx, overlay := range tempOverlays {
			overlays[idx] = overlay
		}
	} else {
		paramIndex, err := strconv.Atoi(options[0])
		if err != nil {
			return
		}

		overlay := overlays[paramIndex]
		if overlay == nil {
			overlay = &OverlayParams{}
			overlays[paramIndex] = overlay
		}

		needsVision = overlay.ParseParams(param, options[1:])
	}

	return
}

func (bg *OverlayBackgroundParams) Process(ox int, oy int, sourceImage *vips.ImageRef, overlayImg *vips.ImageRef, overlay *OverlayParams, imageMeta *vision.Metadata) (int, int, *vips.ImageRef, error) {
	if bg.BackgroundColor == nil && bg.Blur == nil && bg.Saturation == nil && bg.Contrast == nil && bg.Brightness == nil {
		return ox, oy, overlayImg, nil
	}

	_ = DumpDebugImage("overlayImg", overlayImg)

	hAnchor := utils.IfNil(overlay.HAnchor, "right")
	vAnchor := utils.IfNil(overlay.VAnchor, "bottom")
	sourceHPadding := utils.IfNil(overlay.HPadding, 0)
	sourceVPadding := utils.IfNil(overlay.VPadding, 0)

	radius := utils.IfNil(bg.CornerRadius, 0)
	relativePadding := utils.IfNil(bg.RelativePadding, false)
	hPadding := utils.IfNil(bg.HPadding, 0)
	vPadding := utils.IfNil(bg.VPadding, 0)
	if relativePadding {
		hPadding = int(math.Round((float64(hPadding) / 100.0) * float64(sourceImage.Width())))
		vPadding = int(math.Round((float64(vPadding) / 100.0) * float64(sourceImage.Height())))
	}

	bgWidth := overlayImg.Width() + (hPadding * 2)
	bgHeight := overlayImg.Height() + (vPadding * 2)

	relativeSize := utils.IfNil(bg.RelativeSize, true)
	width := utils.IfNil(bg.Width, 0)
	height := utils.IfNil(bg.Height, 0)
	minWidth := utils.IfNil(bg.MinWidth, 0)
	minHeight := utils.IfNil(bg.MinHeight, 0)
	maxWidth := utils.IfNil(bg.MaxWidth, 0)
	maxHeight := utils.IfNil(bg.MaxHeight, 0)

	if maxWidth == 0 {
		maxWidth = math.MaxInt
	}

	if maxHeight == 0 {
		maxHeight = math.MaxInt
	}

	if width > 0 {
		if relativeSize {
			width = int(float64(sourceImage.Width()-(sourceHPadding*2)) * (float64(width) / 100.0))
		}

		bgWidth = utils.Min(sourceImage.Width()-(sourceHPadding*2), utils.Min(maxWidth, utils.Max(minWidth, utils.Max(width, bgWidth))))
	}

	if height > 0 {
		if relativeSize {
			height = int(float64(sourceImage.Height()-(sourceVPadding*2)) * (float64(height) / 100.0))
		}

		bgHeight = utils.Min(sourceImage.Height()-(sourceVPadding*2), utils.Min(maxHeight, utils.Max(minHeight, utils.Max(height, bgHeight))))
	}

	dw := bgWidth - overlayImg.Width()
	dh := bgHeight - overlayImg.Height()

	var dx, dy int
	if hAnchor == "left" {
		dx = ox
	} else if hAnchor == "center" {
		dx = ox - (dw / 2)
	} else if hAnchor == "right" {
		dx = ox - dw
	}

	if vAnchor == "top" {
		dy = oy
	} else if vAnchor == "center" {
		dy = oy - (dh / 2)
	} else if vAnchor == "bottom" {
		dy = oy - dh
	}

	bgColor := utils.IfNil(bg.BackgroundColor, "#ffffff00")
	if bgColor[0] != '#' {
		bgColor = "#" + bgColor
	}

	if bg.BackgroundColorType != nil && *bg.BackgroundColorType != "color" {
		if *bg.BackgroundColorType == "dom" && len(imageMeta.DominantColors.Colors) > 0 {
			dominantColor := imageMeta.DominantColors.Colors[0]
			bgColor = fmt.Sprintf("#%02x%02x%02x%02x", dominantColor.R, dominantColor.G, dominantColor.B, int(utils.IfNil(bg.DominantColorOpacity, 1.0)*255))
		} else if *bg.BackgroundColorType == "light" && imageMeta.DominantColors.Lightest != nil {
			bgColor = fmt.Sprintf("#%02x%02x%02x%02x", imageMeta.DominantColors.Lightest.R, imageMeta.DominantColors.Lightest.G, imageMeta.DominantColors.Lightest.B, int(utils.IfNil(bg.DominantColorOpacity, 1.0)*255))
		} else if *bg.BackgroundColorType == "dark" && imageMeta.DominantColors.Darkest != nil {
			bgColor = fmt.Sprintf("#%02x%02x%02x%02x", imageMeta.DominantColors.Darkest.R, imageMeta.DominantColors.Darkest.G, imageMeta.DominantColors.Darkest.B, int(utils.IfNil(bg.DominantColorOpacity, 1.0)*255))
		}
	}

	dx = utils.Max(0, dx)
	dy = utils.Max(0, dy)

	bgWidth = utils.Min(dx+bgWidth, sourceImage.Width()-dx)
	bgHeight = utils.Min(dy+bgHeight, sourceImage.Height()-dy)

	rects := fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" rx="%f" style="fill:%s" />`, dx, dy, bgWidth, bgHeight, radius, bgColor)
	svg := fmt.Sprintf(`<svg width="%d" height="%d" viewport="0 0 %d %d" xmlns="http://www.w3.org/2000/svg">%s</svg>`, sourceImage.Width(), sourceImage.Height(), sourceImage.Width(), sourceImage.Height(), rects)
	svgImage, err := utils.RenderSVG(svg)
	if err != nil {
		return ox, oy, overlayImg, err
	}

	_ = DumpDebugImage("svgImage", svgImage)

	sourceCopy, err := sourceImage.Copy()
	if err != nil {
		return ox, oy, overlayImg, err
	}
	if !sourceCopy.HasAlpha() {
		_ = sourceCopy.AddAlpha()
	}

	c := utils.IfNil(bg.Contrast, 1)
	if c != 1 {
		err = sourceCopy.Linear1(c, -0.5*c-0.5)
		if err != nil {
			return ox, oy, overlayImg, err
		}
	}

	b := utils.IfNil(bg.Brightness, 1)
	s := utils.IfNil(bg.Saturation, 1)
	if b != 1 || s != 1 {
		err = sourceCopy.Modulate(b, s, 0)
		if err != nil {
			return ox, oy, overlayImg, err
		}
	}

	if bg.Blur != nil && *bg.Blur > 0 {
		_ = sourceCopy.GaussianBlur(float64(*bg.Blur))
		if env.FoxyEnvironment.AlwaysPrerender {
			sourceCopy, err = utils.RenderImage(sourceCopy)
			if err != nil {
				return ox, oy, overlayImg, err
			}
		}
	}

	_ = sourceCopy.Composite(svgImage, vips.BlendModeOver, 0, 0)
	_ = DumpDebugImage("sourceCopyComposite", sourceCopy)

	err = sourceCopy.Crop(dx, dy, bgWidth, bgHeight)
	if err != nil {
		return ox, oy, overlayImg, err
	}

	hAlign := utils.IfNil(bg.HAlign, "center")
	vAlign := utils.IfNil(bg.VAlign, "center")
	oix := hPadding
	oiy := vPadding

	if hAlign == "center" {
		oix = (bgWidth - overlayImg.Width()) / 2
	} else if hAlign == "right" {
		oix = bgWidth - overlayImg.Width() - hPadding
	}

	if vAlign == "center" {
		oiy = (bgHeight - overlayImg.Height()) / 2
	} else if vAlign == "bottom" {
		oiy = bgHeight - overlayImg.Height() - vPadding
	}

	if !overlayImg.HasAlpha() {
		_ = overlayImg.AddAlpha()
	}

	err = sourceCopy.Composite(overlayImg, vips.BlendModeOver, oix, oiy)
	if err != nil {
		return ox, oy, overlayImg, err
	}

	_ = DumpDebugImage("sourceCopy", sourceCopy)

	return dx, dy, sourceCopy, nil
}

func (overlay *OverlayParams) ProcessTextOverlay(sourceId string, config *config.Config, sourceImage *vips.ImageRef, params *ImageParams, imageMeta *vision.Metadata) (*vips.ImageRef, error) {
	return sourceImage, nil
}

func (overlay *OverlayParams) ProcessImageOverlay(sourceId string, config *config.Config, sourceImage *vips.ImageRef, params *ImageParams, imageMeta *vision.Metadata) (*vips.ImageRef, error) {
	if overlay.ImageKey == nil {
		return sourceImage, nil
	}

	relativeSize := utils.IfNil(overlay.RelativeSize, true)
	width := utils.IfNil(overlay.Width, 0)
	height := utils.IfNil(overlay.Height, 0)
	minWidth := utils.IfNil(overlay.MinWidth, 0)
	minHeight := utils.IfNil(overlay.MinHeight, 0)
	maxWidth := utils.IfNil(overlay.MaxWidth, 0)
	maxHeight := utils.IfNil(overlay.MaxHeight, 0)
	if width == 0 && height == 0 {
		return sourceImage, nil
	}

	hpadding := utils.IfNil(overlay.HPadding, 0)
	vpadding := utils.IfNil(overlay.VPadding, 0)

	sw := sourceImage.Width() - (hpadding * 2)
	sh := sourceImage.Height() - (vpadding * 2)

	opacity := utils.IfNil(overlay.Opacity, 1.0)
	if opacity <= 0 {
		return sourceImage, nil
	}

	var overlayImg *vips.ImageRef

	if filepath.Ext(*overlay.ImageKey) == ".svg" {
		svgCodeBytes, err := storage.GetSourceText(config, sourceId, *overlay.ImageKey, params.Debug.DisableSourceCache)
		if err != nil {
			return sourceImage, err
		}

		svgCode := string(*svgCodeBytes)
		for k, v := range overlay.Substitutes {
			if v == "{{light-or-dark}}" {
				if len(imageMeta.DominantColors.Colors) > 0 && imageMeta.DominantColors.Lightest != nil && imageMeta.DominantColors.Darkest != nil {
					dominantColor := imageMeta.DominantColors.Colors[0]
					lDist := math.Abs(dominantColor.L - imageMeta.DominantColors.Lightest.L)
					dDist := math.Abs(dominantColor.L - imageMeta.DominantColors.Darkest.L)
					if lDist > dDist {
						newV := fmt.Sprintf("#%02x%02x%02x", imageMeta.DominantColors.Lightest.R, imageMeta.DominantColors.Lightest.G, imageMeta.DominantColors.Lightest.B)
						svgCode = strings.ReplaceAll(svgCode, k, newV)
					} else {
						newV := fmt.Sprintf("#%02x%02x%02x", imageMeta.DominantColors.Darkest.R, imageMeta.DominantColors.Darkest.G, imageMeta.DominantColors.Darkest.B)
						svgCode = strings.ReplaceAll(svgCode, k, newV)
					}
				}
			} else if v == "{{light}}" {
				if imageMeta.DominantColors.Lightest != nil {
					newV := fmt.Sprintf("#%02x%02x%02x", imageMeta.DominantColors.Lightest.R, imageMeta.DominantColors.Lightest.G, imageMeta.DominantColors.Lightest.B)
					svgCode = strings.ReplaceAll(svgCode, k, newV)
				}

			} else if v == "{{dark}}" {
				if imageMeta.DominantColors.Darkest != nil {
					newV := fmt.Sprintf("#%02x%02x%02x", imageMeta.DominantColors.Darkest.R, imageMeta.DominantColors.Darkest.G, imageMeta.DominantColors.Darkest.B)
					svgCode = strings.ReplaceAll(svgCode, k, newV)
				}
			} else if v == "{{mid}}" {
				if len(imageMeta.DominantColors.Colors) > 0 {
					colorsCopy := make([]vision.UsedColor, len(imageMeta.DominantColors.Colors))
					copy(colorsCopy, imageMeta.DominantColors.Colors)
					sort.Slice(colorsCopy, func(i, j int) bool {
						return colorsCopy[i].L > colorsCopy[j].L
					})

					midColor := colorsCopy[len(colorsCopy)/2]
					newV := fmt.Sprintf("#%02x%02x%02x", midColor.R, midColor.G, midColor.B)
					svgCode = strings.ReplaceAll(svgCode, k, newV)
				}
			} else if v == "{{dominant}}" {
				if len(imageMeta.DominantColors.Colors) > 0 {
					dominantColor := imageMeta.DominantColors.Colors[0]
					newV := fmt.Sprintf("#%02x%02x%02x", dominantColor.R, dominantColor.G, dominantColor.B)
					svgCode = strings.ReplaceAll(svgCode, k, newV)
				}
			} else {
				svgCode = strings.ReplaceAll(svgCode, k, v)
			}
		}

		oimg, err := vips.NewImageFromBuffer([]byte(svgCode))
		if err != nil {
			return sourceImage, err
		}

		overlayImg = oimg
	} else {
		oimg, err := storage.GetSourceImage(config, sourceId, *overlay.ImageKey, params.Debug.DisableSourceCache)
		if err != nil {
			return sourceImage, err
		}

		overlayImg = oimg
	}

	if utils.IfNil(overlay.Trim, false) {
		untrimmedAlpha, err := overlayImg.ExtractBandToImage(3, 1)
		if err != nil {
			return sourceImage, err
		}

		l, t, w, h, err := untrimmedAlpha.FindTrim(0, &vips.Color{R: 0, G: 0, B: 0})
		if err != nil {
			return sourceImage, err
		}

		_ = overlayImg.Crop(l, t, w, h)
	}

	rotate := utils.IfNil(overlay.Rotate, 0)
	if rotate != 0 {
		switch rotate {
		case 90:
			_ = overlayImg.Rotate(vips.Angle90)
		case 180:
			_ = overlayImg.Rotate(vips.Angle180)
		case 270:
			_ = overlayImg.Rotate(vips.Angle270)
		}

		_ = DumpDebugImage("rotate", overlayImg)
	}

	relativeCoords := utils.IfNil(overlay.RelativeCoords, true)
	x := utils.IfNil(overlay.X, 0)
	y := utils.IfNil(overlay.Y, 0)
	if relativeCoords {
		x = hpadding + int(math.Round(float64(sw)*(float64(x)/100.0)))
		y = vpadding + int(math.Round(float64(sh)*(float64(y)/100.0)))
	}

	if relativeSize {
		if width > 0 {
			width = int(float64(sw) * (float64(width) / 100.0))
		}

		if height > 0 {
			height = int(float64(sh) * (float64(height) / 100.0))
		}
	} else {
		minWidth = int(float64(sw) * (float64(minWidth) / 100.0))
		minHeight = int(float64(sh) * (float64(minHeight) / 100.0))
		maxWidth = int(float64(sw) * (float64(maxWidth) / 100.0))
		maxHeight = int(float64(sh) * (float64(maxHeight) / 100.0))
	}

	if maxWidth == 0 {
		maxWidth = math.MaxInt
	}

	if maxHeight == 0 {
		maxHeight = math.MaxInt
	}

	width = utils.Min(maxWidth, utils.Max(minWidth, width))
	height = utils.Min(maxHeight, utils.Max(minHeight, height))

	ratio := float64(overlayImg.Width()) / float64(overlayImg.Height())
	if width > 0 && height == 0 {
		height = int(math.Round(float64(width) * (1.0 / ratio)))
	} else if height > 0 && width == 0 {
		width = int(math.Round(float64(height) * ratio))
	}

	if width > sourceImage.Width() || height > sourceImage.Height() {
		fitSz := geometry.SizeToFitSize(width, height, sourceImage.Width()-(hpadding*2), sourceImage.Height()-(vpadding*2))
		width = fitSz.Width
		height = fitSz.Height
	}

	if width < minWidth && minWidth < sourceImage.Width()-(hpadding*2) {
		width = minWidth
	} else if width > maxWidth {
		width = utils.Min(maxWidth, sourceImage.Width()-(hpadding*2))
	}

	if height < minHeight && minHeight < sourceImage.Height()-(vpadding*2) {
		height = minHeight
	} else if height > maxHeight {
		height = utils.Min(maxHeight, sourceImage.Height()-(vpadding*2))
	}

	fit := utils.IfNil(overlay.Fit, "fit")
	if fit == "fit" {
		newSize := geometry.SizeToFitSize(overlayImg.Width(), overlayImg.Height(), width, height)
		width = newSize.Width
		height = newSize.Height

		_ = overlayImg.ThumbnailWithSize(width, height, vips.InterestingNone, vips.SizeBoth)
	} else if fit == "cover" {
		newSize := geometry.SizeToFillSize(overlayImg.Width(), overlayImg.Height(), width, height, true)
		width = newSize.Width
		height = newSize.Height

		_ = overlayImg.ThumbnailWithSize(width, height, vips.InterestingNone, vips.SizeBoth)
	} else if fit == "crop" {
		newSize := geometry.SizeToFillSize(overlayImg.Width(), overlayImg.Height(), width, height, true)
		dx := (overlayImg.Width() - newSize.Width) / 2
		dy := (overlayImg.Height() - newSize.Height) / 2

		_ = overlayImg.Crop(dx, dy, newSize.Width, newSize.Height)
	} else if fit == "stretch" {
		_ = overlayImg.ResizeWithVScale(float64(overlayImg.Width())/float64(width), float64(overlayImg.Height())/float64(height), vips.KernelLanczos3)
	}

	var dx, dy int
	hAnchor := utils.IfNil(overlay.HAnchor, "right")
	if hAnchor == "left" {
		dx = x
	} else if hAnchor == "center" {
		dx = x - (overlayImg.Width() / 2)
	} else if hAnchor == "right" {
		dx = x - overlayImg.Width()
	}

	vAnchor := utils.IfNil(overlay.VAnchor, "bottom")
	if vAnchor == "top" {
		dy = y
	} else if vAnchor == "center" {
		dy = y - (overlayImg.Height() / 2)
	} else if vAnchor == "bottom" {
		dy = y - overlayImg.Height()
	}

	if opacity < 1.0 {
		if !overlayImg.HasAlpha() {
			_ = overlayImg.AddAlpha()
		}

		_ = overlayImg.Linear([]float64{1.0, 1.0, 1.0, opacity}, []float64{0.0, 0.0, 0.0, 0.0})
	}

	if overlay.Background != nil {
		ndx, ndy, newOverlayImg, err := overlay.Background.Process(dx, dy, sourceImage, overlayImg, overlay, imageMeta)
		if err != nil {
			return sourceImage, err
		}

		dx = ndx
		dy = ndy
		overlayImg = newOverlayImg
	}

	if overlayImg.HasAlpha() && !sourceImage.HasAlpha() {
		_ = sourceImage.AddAlpha()
	}

	_ = sourceImage.Composite(overlayImg, vips.BlendModeOver, dx, dy)

	return sourceImage, nil
}

func (overlay *OverlayParams) Process(sourceKey string, sourceId string, config *config.Config, sourceImage *vips.ImageRef, params *ImageParams, imageMeta *vision.Metadata) (*vips.ImageRef, error) {
	if overlay.Type == "text" {
		return overlay.ProcessTextOverlay(sourceId, config, sourceImage, params, imageMeta)
	} else if overlay.Type == "image" {
		return overlay.ProcessImageOverlay(sourceId, config, sourceImage, params, imageMeta)
	}

	return sourceImage, nil
}

func (overlays Overlays) Process(sourceKey string, sourceId string, config *config.Config, sourceImage *vips.ImageRef, params *ImageParams, imageMeta *vision.Metadata) (*vips.ImageRef, error) {
	for _, overlay := range overlays {
		var err error
		sourceImage, err = overlay.Process(sourceKey, sourceId, config, sourceImage, params, imageMeta)
		if err != nil {
			return nil, err
		}
	}

	return sourceImage, nil
}

/**
/ov:<overlay_index>:type:<image|text|rect|ellipse>
/ov:<overlay_index>:url:<base64 encoded source key or path>
/ov:<overlay_index>:text:<base64 encoded text>
/ov:<overlay_index>:font:<base64 encoded fontname>
/ov:<overlay_index>:xy:<px|rel>:<x>:<y>
/ov:<overlay_index>:a:<left|center|right>:<top|center|bottom>
/ov:<overlay_index>:sz:<px|rel>:<width>:<height>
/ov:<overlay_index>:minsz:<width>:<height>
/ov:<overlay_index>:maxsz:<width>:<height>
/ov:<overlay_index>:rot:<angle>
/ov:<overlay_index>:o:<opacity%>
/ov:<overlay_index>:fit:<fit|fill|crop>
/ov:<overlay_index>:tc:<text color>
/ov:<overlay_index>:fc:<fill color>
/ov:<overlay_index>:sc:<stroke color>
/ov:<overlay_index>:sw:<stroke width>
/ov:<overlay_index>:ds:o:<drop shadow opacity%>
/ov:<overlay_index>:ds:c:<drop shadow color>
/ov:<overlay_index>:ds:bl:<drop shadow blur>
/ov:<overlay_index>:ds:xy:<x offset>:<y offset>
/ov:<overlay_index>:bg:c:<background color>
/ov:<overlay_index>:bg:bl:<background blur>
/ov:<overlay_index>:bg:sat:<background saturation>
/ov:<overlay_index>:bg:con:<background contrast>
/ov:<overlay_index>:bg:bri:<background brightness>
/ov:<overlay_index>:bg:pad:<px|rel>:<h padding>:<v padding>
*/
