package vision

import (
	"errors"
	"foxy/internal/config"
	"foxy/internal/geometry"
	"github.com/lucasb-eyer/go-colorful"
	"log"
	"math"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rekognition"
	"github.com/davidbyttow/govips/v2/vips"
)

func RekognitionDetectFaces(sourceId string, sourceConfig *config.Config, sid string, key string, sourceImage *vips.ImageRef) (*Metadata, error) {
	var rekConfig config.S3Config
	if sourceConfig.Vision.UseSourceCredentials {
		rekConfig = sourceConfig.Source.S3Config
	} else {
		rekConfig = sourceConfig.Vision.S3Config
	}

	if rekConfig.Secret == nil || rekConfig.Key == nil || rekConfig.Region == nil {
		return nil, errors.New("rekognition key and secret not set")
	}

	sess := session.Must(session.NewSession(&aws.Config{
		Region: rekConfig.Region,
		Credentials: credentials.NewStaticCredentials(
			*rekConfig.Key,
			*rekConfig.Secret,
			"",
		),
	}))

	svc := rekognition.New(sess)

	var rekImage *rekognition.Image
	if sourceConfig.Source.Type != "s3" || (sourceConfig.Source.Type == "s3" && !sourceConfig.Vision.UseSourceCredentials) || sourceImage.Width() > 2560 || sourceImage.Height() > 2560 {
		log.Println("Image too large for rekognition, resizing")
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

			jpegImage, _, _ := sourceCopy.ExportJpeg(&vips.JpegExportParams{Quality: 80})
			rekImage = &rekognition.Image{
				Bytes: jpegImage,
			}
		} else {
			jpegImage, _, _ := sourceImage.ExportJpeg(&vips.JpegExportParams{Quality: 80})
			rekImage = &rekognition.Image{
				Bytes: jpegImage,
			}
		}
	} else {
		rekImage = &rekognition.Image{
			S3Object: &rekognition.S3Object{
				Bucket: rekConfig.Bucket,
				Name:   aws.String(strings.TrimLeft(key, "/")),
			},
		}
	}

	facesResult, err := svc.DetectFaces(&rekognition.DetectFacesInput{
		Attributes: []*string{
			aws.String("DEFAULT"),
			aws.String("FACE_OCCLUDED"),
			aws.String("AGE_RANGE"),
			aws.String("GENDER"),
		},
		Image: rekImage,
	})

	if err != nil {
		return nil, err
	}

	labelsResults, err := svc.DetectLabels(&rekognition.DetectLabelsInput{
		MaxLabels: aws.Int64(64),
		Image:     rekImage,
		Features:  []*string{aws.String("GENERAL_LABELS"), aws.String("IMAGE_PROPERTIES")},
	})

	if err != nil {
		return nil, err
	}

	moderationResults, err := svc.DetectModerationLabels(&rekognition.DetectModerationLabelsInput{
		Image: rekImage,
	})

	if err != nil {
		return nil, err
	}

	var result = &Metadata{}
	result.Faces = make([]Face, len(facesResult.FaceDetails))
	result.Labels = make([]Label, len(labelsResults.Labels))
	result.ModerationLabels = make([]Label, len(moderationResults.ModerationLabels))

	for idx, face := range facesResult.FaceDetails {
		result.Faces[idx].Box.Left = *face.BoundingBox.Left
		result.Faces[idx].Box.Top = *face.BoundingBox.Top
		result.Faces[idx].Box.Width = *face.BoundingBox.Width
		result.Faces[idx].Box.Height = *face.BoundingBox.Height
		if face.Pose != nil {
			result.Faces[idx].Rotation = *face.Pose.Roll
		}
		result.Faces[idx].Confidence = *face.Confidence
		if face.AgeRange != nil {
			result.Faces[idx].Age = &AgeRange{
				High: *face.AgeRange.High,
				Low:  *face.AgeRange.Low,
			}
		}

		if face.Gender != nil {
			result.Faces[idx].Gender = face.Gender.Value
		}
	}

	for idx, label := range labelsResults.Labels {
		hasBox := false
		boxLeft := math.MaxFloat64
		boxTop := math.MaxFloat64
		boxWidth := float64(0)
		boxHeight := float64(0)

		if label.Instances != nil {
			for _, instance := range label.Instances {
				if instance.BoundingBox != nil {
					hasBox = true
					boxLeft = math.Min(boxLeft, *instance.BoundingBox.Left)
					boxTop = math.Min(boxTop, *instance.BoundingBox.Top)
					boxWidth = math.Max(boxWidth, *instance.BoundingBox.Width)
					boxHeight = math.Max(boxHeight, *instance.BoundingBox.Height)
				}
			}
		}

		result.Labels[idx].Name = *label.Name
		if hasBox {
			result.Labels[idx].Box = &geometry.Box{
				Left:   boxLeft,
				Top:    boxTop,
				Width:  boxWidth,
				Height: boxHeight,
			}
		}

		result.Labels[idx].Confidence = *label.Confidence
	}

	for idx, label := range moderationResults.ModerationLabels {
		result.ModerationLabels[idx].Name = *label.Name
		result.ModerationLabels[idx].Confidence = *label.Confidence
	}

	result.People = GetPersonLabels(result.Labels)

	darkestVal := math.MaxFloat64
	lightestVal := 0.0
	darkestIdx := -1
	lightestIdx := -1

	if labelsResults.ImageProperties != nil {
		for idx, color := range labelsResults.ImageProperties.DominantColors {
			rgbColor := colorful.Color{R: float64(*color.Red) / 255.0, G: float64(*color.Green) / 255.0, B: float64(*color.Blue) / 255.0}
			l, _, _ := rgbColor.Lab()
			r := int(*color.Red)
			g := int(*color.Green)
			b := int(*color.Blue)
			result.DominantColors.Colors = append(result.DominantColors.Colors, UsedColor{
				Used: *color.PixelPercent,
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
			result.DominantColors.Lightest = &UsedColor{
				Used: result.DominantColors.Colors[lightestIdx].Used,
				R:    result.DominantColors.Colors[lightestIdx].R,
				G:    result.DominantColors.Colors[lightestIdx].G,
				B:    result.DominantColors.Colors[lightestIdx].B,
				L:    result.DominantColors.Colors[lightestIdx].L,
			}
		}

		if darkestIdx > -1 {
			result.DominantColors.Darkest = &UsedColor{
				Used: result.DominantColors.Colors[darkestIdx].Used,
				R:    result.DominantColors.Colors[darkestIdx].R,
				G:    result.DominantColors.Colors[darkestIdx].G,
				B:    result.DominantColors.Colors[darkestIdx].B,
				L:    result.DominantColors.Colors[darkestIdx].L,
			}
		}

		if len(result.DominantColors.Colors) > 0 {
			sort.Slice(result.DominantColors.Colors, func(i, j int) bool {
				return result.DominantColors.Colors[i].Used > result.DominantColors.Colors[j].Used
			})
		}

	}

	return result, nil
}
