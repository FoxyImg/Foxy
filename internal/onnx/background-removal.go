package onnx

import "foxy/internal/env"

var HumanBackgroundRemoval *ModelSession = nil
var GenericBackgroundRemoval *ModelSession = nil

func InitBackgroundRemoval() error {
	if env.FoxyEnvironment.UseML == nil || !*env.FoxyEnvironment.UseML || env.FoxyEnvironment.OnnxLib == nil || env.FoxyEnvironment.OnnxModelHumans == nil || env.FoxyEnvironment.OnnxModelForeground == nil {
		return nil
	}

	var err error
	HumanBackgroundRemoval, err = NewModelSession(*env.FoxyEnvironment.OnnxModelHumans)
	if err != nil {
		return err
	}

	GenericBackgroundRemoval, err = NewModelSession(*env.FoxyEnvironment.OnnxModelForeground)
	if err != nil {
		return err
	}

	return nil
}
