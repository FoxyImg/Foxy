package params

import (
	"foxy/internal/config"
	"foxy/internal/env"
)

func Boot() error {
	if env.FoxyEnvironment.Isolated {
		err := config.LoadSourceConfigFromJSON()
		if err != nil {
			return err
		}

		return LoadPresetsFromJSON()
	}

	return nil
}
