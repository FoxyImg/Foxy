package vision

import (
	"bytes"
	vision "cloud.google.com/go/vision/apiv1"
	"context"
	"foxy/internal/config"
	"foxy/internal/geometry"
	"foxy/internal/utils"
	"github.com/davidbyttow/govips/v2/vips"
	"github.com/lucasb-eyer/go-colorful"
	pb "google.golang.org/genproto/googleapis/cloud/vision/v1"
	"math"
	"slices"
	"sort"
)

var personMids = []string{
	"/m/01g317",  //Person
	"/m/04yx4",   //Man
	"/m/03bt1vf", //Woman
	"/m/05r655",  //Girl
	"/m/01bl7v",  //Boy
	"/m/02p0tk3", //Human body
}

func processEntityAnnotations(nw int, nh int, annotations []*pb.EntityAnnotation) *[]Label {
	var labels = make([]Label, 0)

	for _, annotation := range annotations {
		if annotation.BoundingPoly == nil {
			continue
		}

		if len(annotation.BoundingPoly.Vertices) == 4 {
			minX := math.MaxInt32
			minY := math.MaxInt32
			maxX := 0
			maxY := 0
			for _, vertex := range annotation.BoundingPoly.Vertices {
				minX = utils.Min(minX, int(vertex.X))
				minY = utils.Min(minY, int(vertex.Y))
				maxX = utils.Max(maxX, int(vertex.X))
				maxY = utils.Max(maxY, int(vertex.Y))
			}

			label := Label{
				Name:       annotation.Description,
				Confidence: float64(annotation.Score),
				Box: &geometry.Box{
					Left:   float64(minX) / float64(nw),
					Top:    float64(minY) / float64(nh),
					Width:  float64(maxX-minX) / float64(nw),
					Height: float64(maxY-minY) / float64(nh),
				},
			}

			labels = append(labels, label)
		}
	}

	return &labels
}

func GoogleCloudVisionDetectFaces(sourceId string, sourceConfig *config.Config, sid string, key string, sourceImage *vips.ImageRef) (*Metadata, error) {
	var nw, nh int
	var imageBytes []byte
	if sourceImage.Width() > 2560 || sourceImage.Height() > 2560 {
		sourceCopy, err := sourceImage.Copy()
		if err != nil {
			return nil, err
		}
		defer sourceCopy.Close()

		scale := math.Min(0.5, 2560.0/float64(sourceImage.Width()))
		err = sourceCopy.Resize(scale, vips.KernelLanczos3)
		if err != nil {
			return nil, err
		}

		nw = sourceCopy.Width()
		nh = sourceCopy.Height()

		jpegImage, _, err := sourceCopy.ExportJpeg(&vips.JpegExportParams{Quality: 80})
		if err != nil {
			return nil, err
		}

		imageBytes = jpegImage
	} else {
		nw = sourceImage.Width()
		nh = sourceImage.Height()

		jpegImage, _, err := sourceImage.ExportJpeg(&vips.JpegExportParams{Quality: 80})
		if err != nil {
			return nil, err
		}

		imageBytes = jpegImage
	}

	ctx := context.Background()
	client, err := vision.NewImageAnnotatorClient(ctx)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	image, err := vision.NewImageFromReader(bytes.NewReader(imageBytes))
	if err != nil {
		return nil, err
	}

	annotations, err := client.DetectFaces(ctx, image, nil, 10)
	if err != nil {
		return nil, err
	}

	var meta = Metadata{
		Width:            sourceImage.Width(),
		Height:           sourceImage.Height(),
		Faces:            make([]Face, 0),
		Labels:           make([]Label, 0),
		People:           make([]Label, 0),
		ModerationLabels: make([]Label, 0),
	}

	for _, annotation := range annotations {
		if annotation.BoundingPoly == nil {
			continue
		}

		if len(annotation.BoundingPoly.Vertices) == 4 {
			minX := math.MaxInt32
			minY := math.MaxInt32
			maxX := 0
			maxY := 0
			for _, vertex := range annotation.BoundingPoly.Vertices {
				minX = utils.Min(minX, int(vertex.X))
				minY = utils.Min(minY, int(vertex.Y))
				maxX = utils.Max(maxX, int(vertex.X))
				maxY = utils.Max(maxY, int(vertex.Y))
			}

			face := Face{
				Confidence: float64(annotation.DetectionConfidence),
				Box: geometry.Box{
					Left:   float64(minX) / float64(nw),
					Top:    float64(minY) / float64(nh),
					Width:  float64(maxX-minX) / float64(nw),
					Height: float64(maxY-minY) / float64(nh),
				},
			}

			meta.Faces = append(meta.Faces, face)
		}
	}

	labels, err := client.DetectLabels(ctx, image, nil, 20)
	if err != nil {
		return nil, err
	}
	for _, label := range labels {
		newLabel := Label{
			Name:       label.Description,
			Confidence: float64(label.Score),
		}

		if label.BoundingPoly != nil {
			if len(label.BoundingPoly.Vertices) == 4 {
				minX := math.MaxInt32
				minY := math.MaxInt32
				maxX := 0
				maxY := 0
				for _, vertex := range label.BoundingPoly.Vertices {
					minX = utils.Min(minX, int(vertex.X))
					minY = utils.Min(minY, int(vertex.Y))
					maxX = utils.Max(maxX, int(vertex.X))
					maxY = utils.Max(maxY, int(vertex.Y))
				}

				newLabel.Box = &geometry.Box{
					Left:   float64(minX) / float64(nw),
					Top:    float64(minY) / float64(nh),
					Width:  float64(maxX-minX) / float64(nw),
					Height: float64(maxY-minY) / float64(nh),
				}

			}
		}

		meta.Labels = append(meta.Labels, newLabel)
	}

	objects, err := client.LocalizeObjects(ctx, image, nil)
	if err != nil {
		return nil, err
	}

	for _, object := range objects {
		newLabel := Label{
			Name:       object.Name,
			Confidence: float64(object.Score),
		}

		if object.BoundingPoly != nil {
			if len(object.BoundingPoly.Vertices) == 4 {
				minX := math.MaxInt32
				minY := math.MaxInt32
				maxX := 0
				maxY := 0
				for _, vertex := range object.BoundingPoly.Vertices {
					minX = utils.Min(minX, int(vertex.X))
					minY = utils.Min(minY, int(vertex.Y))
					maxX = utils.Max(maxX, int(vertex.X))
					maxY = utils.Max(maxY, int(vertex.Y))
				}

				newLabel.Box = &geometry.Box{
					Left:   float64(minX) / float64(nw),
					Top:    float64(minY) / float64(nh),
					Width:  float64(maxX-minX) / float64(nw),
					Height: float64(maxY-minY) / float64(nh),
				}

			} else if len(object.BoundingPoly.NormalizedVertices) == 4 {
				minX := math.MaxFloat64
				minY := math.MaxFloat64
				maxX := 0.0
				maxY := 0.0

				for _, vertex := range object.BoundingPoly.NormalizedVertices {
					minX = math.Min(minX, float64(vertex.X))
					minY = math.Min(minY, float64(vertex.Y))
					maxX = math.Max(maxX, float64(vertex.X))
					maxY = math.Max(maxY, float64(vertex.Y))
				}

				newLabel.Box = &geometry.Box{
					Left:   minX,
					Top:    minY,
					Width:  maxX - minX,
					Height: maxY - minY,
				}

			}
		}

		if slices.Contains(personMids, object.Mid) && newLabel.Box != nil {
			meta.People = append(meta.People, newLabel)
		} else {
			meta.Labels = append(meta.Labels, newLabel)
		}
	}

	webSafe, err := client.DetectSafeSearch(ctx, image, nil)
	if err != nil {
		return nil, err
	}

	if webSafe != nil {
		if webSafe.Adult >= pb.Likelihood_LIKELY {
			meta.ModerationLabels = append(meta.ModerationLabels, Label{
				Name:       "Adult",
				Confidence: float64(webSafe.Adult) * 20.0,
			})
		}

		if webSafe.Violence >= pb.Likelihood_LIKELY {
			meta.ModerationLabels = append(meta.ModerationLabels, Label{
				Name:       "Violence",
				Confidence: float64(webSafe.Violence) * 20.0,
			})
		}

		if webSafe.Racy >= pb.Likelihood_LIKELY {
			meta.ModerationLabels = append(meta.ModerationLabels, Label{
				Name:       "Racy",
				Confidence: float64(webSafe.Racy) * 20.0,
			})
		}
	}

	imageProps, err := client.DetectImageProperties(ctx, image, nil)
	if err != nil {
		return nil, err
	}

	darkestVal := math.MaxFloat64
	lightestVal := 0.0
	darkestIdx := -1
	lightestIdx := -1

	for idx, quantized := range imageProps.DominantColors.Colors {
		color := quantized.Color
		rgbColor := colorful.Color{R: float64(color.Red), G: float64(color.Green), B: float64(color.Blue)}
		l, _, _ := rgbColor.Lab()
		r := int(color.Red) & 0xff
		g := int(color.Green) & 0xff
		b := int(color.Blue) & 0xff
		meta.DominantColors.Colors = append(meta.DominantColors.Colors, UsedColor{
			Used: float64(quantized.PixelFraction) * 100,
			R:    r,
			G:    g,
			B:    b,
			L:    l,
		})

		if l < darkestVal {
			darkestVal = l
			darkestIdx = idx
		}

		if l > lightestVal {
			lightestVal = l
			lightestIdx = idx
		}
	}

	if lightestIdx > -1 {
		meta.DominantColors.Lightest = &UsedColor{
			Used: meta.DominantColors.Colors[lightestIdx].Used,
			R:    meta.DominantColors.Colors[lightestIdx].R,
			G:    meta.DominantColors.Colors[lightestIdx].G,
			B:    meta.DominantColors.Colors[lightestIdx].B,
			L:    meta.DominantColors.Colors[lightestIdx].L,
		}
	}

	if darkestIdx > -1 {
		meta.DominantColors.Darkest = &UsedColor{
			Used: meta.DominantColors.Colors[darkestIdx].Used,
			R:    meta.DominantColors.Colors[darkestIdx].R,
			G:    meta.DominantColors.Colors[darkestIdx].G,
			B:    meta.DominantColors.Colors[darkestIdx].B,
			L:    meta.DominantColors.Colors[darkestIdx].L,
		}
	}

	if len(imageProps.DominantColors.Colors) > 0 {
		sort.Slice(meta.DominantColors.Colors, func(i, j int) bool {
			return meta.DominantColors.Colors[i].Used > meta.DominantColors.Colors[j].Used
		})
	}

	return &meta, nil
}
