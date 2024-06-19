package onnx

import (
	"bytes"
	"errors"
	"foxy/internal/env"
	"github.com/davidbyttow/govips/v2/vips"
	"github.com/nfnt/resize"
	ort "github.com/yalue/onnxruntime_go"
	"image"
	"image/color"
	"image/png"
	"log"
)

type ModelSession struct {
	ModelPath  *string
	Session    *ort.AdvancedSession
	InputInfo  *ort.InputOutputInfo
	Input      *ort.Tensor[float32]
	OutputInfo *ort.InputOutputInfo
	Output     *ort.Tensor[float32]
}

func Setup() error {
	if env.FoxyEnvironment.UseML == nil || !*env.FoxyEnvironment.UseML || env.FoxyEnvironment.OnnxLib == nil || env.FoxyEnvironment.OnnxModelHumans == nil || env.FoxyEnvironment.OnnxModelForeground == nil {
		return nil
	}

	ort.SetSharedLibraryPath(*env.FoxyEnvironment.OnnxLib)
	err := ort.InitializeEnvironment()
	return err
}

func Teardown() {
	if env.FoxyEnvironment.UseML == nil || !*env.FoxyEnvironment.UseML || env.FoxyEnvironment.OnnxLib == nil || env.FoxyEnvironment.OnnxModelHumans == nil || env.FoxyEnvironment.OnnxModelForeground == nil {
		return
	}

	_ = ort.DestroyEnvironment()
}

func NewModelSession(modelPath string) (*ModelSession, error) {
	session := &ModelSession{}
	err := session.Init(modelPath)
	return session, err
}

func (session *ModelSession) Init(modelPath string) error {
	inputs, outputs, err := ort.GetInputOutputInfo(modelPath)
	if err != nil {
		return err
	}

	if len(inputs) == 0 {
		return errors.New("expected 1 or more inputs, got 0")
	}

	if len(outputs) == 0 {
		return errors.New("expected 1 or more outputs, got 0")
	}

	inputTensor, err := ort.NewEmptyTensor[float32](inputs[0].Dimensions)
	if err != nil {
		log.Panicf("Error creating input tensor: %w", err)
	}

	outputTensor, err := ort.NewEmptyTensor[float32](outputs[0].Dimensions)
	if err != nil {
		log.Panicf("Error creating output tensor: %w", err)
	}

	options, err := ort.NewSessionOptions()
	if err != nil {
		_ = inputTensor.Destroy()
		_ = outputTensor.Destroy()
		return err
	}
	defer options.Destroy()

	if env.FoxyEnvironment.OnnxUseCoreML != nil && *env.FoxyEnvironment.OnnxUseCoreML {
		err = options.AppendExecutionProviderCoreML(0)
		if err != nil {
			_ = inputTensor.Destroy()
			_ = outputTensor.Destroy()
			return err
		}
	}

	onnxSession, err := ort.NewAdvancedSession(modelPath,
		[]string{inputs[0].Name}, []string{outputs[0].Name},
		[]ort.ArbitraryTensor{inputTensor},
		[]ort.ArbitraryTensor{outputTensor},
		options)
	if err != nil {
		_ = inputTensor.Destroy()
		_ = outputTensor.Destroy()
		return err
	}

	*session = ModelSession{
		ModelPath:  &modelPath,
		Session:    onnxSession,
		InputInfo:  &inputs[0],
		Input:      inputTensor,
		OutputInfo: &outputs[0],
		Output:     outputTensor,
	}

	return nil
}

func (session *ModelSession) Destroy() {
	_ = session.Input.Destroy()
	_ = session.Output.Destroy()
}

func (session *ModelSession) prepareInput(sourceImage *vips.ImageRef) error {
	if session.InputInfo.Dimensions[1] != 3 {
		return errors.New("unsupported number of bands")
	}

	srcCopy, err := sourceImage.Copy()
	if err != nil {
		return err
	}

	if srcCopy.Bands() > 3 {
		_ = srcCopy.ExtractBand(3, srcCopy.Bands()-3)
	}

	exportParams := vips.NewDefaultExportParams()
	exportParams.Format = vips.ImageTypePNG
	exportParams.Quality = 100
	exportParams.Compression = 0
	img, err := srcCopy.ToImage(exportParams)
	if err != nil {
		return err
	}

	xsize := int(session.InputInfo.Dimensions[2])
	ysize := int(session.InputInfo.Dimensions[3])

	channelSize := xsize * ysize
	data := session.Input.GetData()

	redChannel := data[0:channelSize]
	greenChannel := data[channelSize : channelSize*2]
	blueChannel := data[channelSize*2 : channelSize*3]

	img = resize.Resize(uint(xsize), uint(ysize), img, resize.Lanczos3)
	i := 0
	for y := 0; y < xsize; y++ {
		for x := 0; x < ysize; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			redChannel[i] = float32(r>>8) / 255.0
			greenChannel[i] = float32(g>>8) / 255.0
			blueChannel[i] = float32(b>>8) / 255.0
			i++
		}
	}

	return nil
}

func (session *ModelSession) prepareOutput(sourceWidth int, sourceHeight int) (*vips.ImageRef, error) {
	data := session.Output.GetData()

	xsize := int(session.OutputInfo.Dimensions[2])
	ysize := int(session.OutputInfo.Dimensions[3])

	outImg := image.NewGray(image.Rectangle{Min: image.Point{X: 0, Y: 0}, Max: image.Point{X: xsize, Y: ysize}})
	for x := 0; x < xsize; x++ {
		for y := 0; y < ysize; y++ {
			outImg.Set(x, y, color.Gray{Y: uint8(data[x+y*xsize] * 255.0)})
		}
	}

	buf := new(bytes.Buffer)
	err := png.Encode(buf, outImg)
	if err != nil {
		return nil, err
	}

	newImg, err := vips.NewImageFromBuffer(buf.Bytes())
	if err != nil {
		return nil, err
	}

	err = newImg.ResizeWithVScale(
		float64(sourceWidth)/float64(newImg.Width()),
		float64(sourceHeight)/float64(newImg.Height()),
		vips.KernelLanczos3,
	)
	if err != nil {
		return nil, err
	}

	return newImg, nil
}

func (session *ModelSession) ProcessImage(sourceImage *vips.ImageRef) (*vips.ImageRef, error) {
	err := session.prepareInput(sourceImage)
	if err != nil {
		return nil, err
	}

	err = session.Session.Run()
	if err != nil {
		return nil, err
	}

	return session.prepareOutput(sourceImage.Width(), sourceImage.Height())
}
