package vision

import (
	"bytes"
	"encoding/json"
	"foxy/internal/config"
	"foxy/internal/geometry"
	"github.com/davidbyttow/govips/v2/vips"
	"io"
	"math"
	"mime/multipart"
	"net/http"
)

type CompreAgeRange struct {
	Probability float64 `json:"probability"`
	High        int64   `json:"high"`
	Low         int64   `json:"low"`
}

type CompreGender struct {
	Probability float64 `json:"probability"`
	Value       string  `json:"value"`
}

type CompreBox struct {
	Probability float64 `json:"probability"`
	Left        int     `json:"x_min"`
	Top         int     `json:"y_min"`
	Right       int     `json:"x_max"`
	Bottom      int     `json:"y_max"`
}

type CompreFace struct {
	AgeRange *CompreAgeRange `json:"age"`
	Gender   *CompreGender   `json:"gender"`
	Box      *CompreBox      `json:"box"`
}

type CompreResult struct {
	Faces []CompreFace `json:"result"`
}

func CompreFaceDetectFaces(sourceId string, sourceConfig *config.Config, sid string, key string, sourceImage *vips.ImageRef) (*Metadata, error) {
	if sourceConfig.Vision.ApiKey == nil || sourceConfig.Vision.Url == nil {
		return nil, nil
	}

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

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	w, err := writer.CreateFormFile("file", "image.jpg")
	if err != nil {
		return nil, err
	}

	_, err = w.Write(imageBytes)
	if err != nil {
		return nil, err
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	r, _ := http.NewRequest("POST", *sourceConfig.Vision.Url+"/api/v1/detection/detect?face_plugins=age,gender", body)
	r.Header.Set("Content-Type", writer.FormDataContentType())
	r.Header.Set("x-api-key", *sourceConfig.Vision.ApiKey)

	client := &http.Client{}
	resp, err := client.Do(r)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var meta = Metadata{
		Width:            sourceImage.Width(),
		Height:           sourceImage.Height(),
		Faces:            make([]Face, 0),
		Labels:           make([]Label, 0),
		People:           make([]Label, 0),
		ModerationLabels: make([]Label, 0),
	}

	if resp.StatusCode != 200 {
		if resp.StatusCode == 400 {
			return &meta, nil
		}

		return nil, nil
	}

	bodyJSON, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result CompreResult
	err = json.Unmarshal(bodyJSON, &result)
	if err != nil {
		return nil, err
	}

	for _, face := range result.Faces {
		if face.Box == nil {
			continue
		}

		newFace := Face{
			Box: geometry.Box{
				Left:   float64(face.Box.Left) / float64(nw),
				Top:    float64(face.Box.Top) / float64(nh),
				Width:  float64(face.Box.Right-face.Box.Left) / float64(nw),
				Height: float64(face.Box.Bottom-face.Box.Top) / float64(nh),
			},
			Confidence: face.Box.Probability,
		}

		if face.AgeRange != nil {
			newFace.Age = &AgeRange{
				High: face.AgeRange.High,
				Low:  face.AgeRange.Low,
			}
		}

		if face.Gender != nil {
			newFace.Gender = &face.Gender.Value
		}

		meta.Faces = append(meta.Faces, newFace)
	}

	return &meta, nil
}
