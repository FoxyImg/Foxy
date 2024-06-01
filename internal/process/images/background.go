package images

import "foxy/internal/utils"

type BackgroundOptions struct {
	Color *string `json:"color,omitempty"`
}

func (opts *BackgroundOptions) Params() []string {
	return []string{"bg"}
}

func (opts *BackgroundOptions) ParseParams(param string, options []string) (needsVision bool) {
	needsVision = false
	if len(options) == 1 {
		opts.Color = utils.Ptr(options[0])
	}

	return
}
