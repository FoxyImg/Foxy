package aws

import (
	"encoding/json"
	"errors"
	"foxy/internal/config"
	"foxy/internal/geometry"
	"foxy/internal/metadata"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rekognition"
	"github.com/davidbyttow/govips/v2/vips"
)

func DetectFaces(sourceConfig config.Config, sid string, key string, sourceImage *vips.ImageRef, skipCache bool) (*metadata.Metadata, error) {
	metaFileName := strings.TrimRight(os.Getenv("CACHE_DIR"), "/") + "/" + sid + "/" + strings.TrimLeft(key, "/") + ".json"

	if !skipCache && os.Getenv("USE_CACHE") == "true" {
		_, err := os.Stat(metaFileName)
		if err == nil {
			jsonData, err := os.ReadFile(metaFileName)
			if err == nil {
				meta := metadata.Metadata{}
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
		Features:  []*string{aws.String("GENERAL_LABELS")},
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

	var result = &metadata.Metadata{}
	result.Faces = make([]metadata.Face, len(facesResult.FaceDetails))
	result.Labels = make([]metadata.Label, len(labelsResults.Labels))
	result.ModerationLabels = make([]metadata.Label, len(moderationResults.ModerationLabels))

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
			result.Faces[idx].Age = &metadata.AgeRange{
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

	result.People = metadata.GetPersonLabels(result.Labels)

	if os.Getenv("USE_CACHE") == "true" {
		metaFilePath := filepath.Dir(metaFileName)
		err = os.MkdirAll(metaFilePath, os.ModePerm)
		if err != nil {
			log.Println("MkdirAll Error:", err)
			return result, nil
		}

		jsonData, err := json.Marshal(result)
		if err != nil {
			log.Println("Marshal Error:", err)
			return result, nil
		}

		_ = os.WriteFile(metaFileName, jsonData, 0644)
	}

	return result, nil
}
