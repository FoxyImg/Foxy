package params

import (
	"fmt"
	"foxy/internal/env"
	"foxy/internal/utils"
	"github.com/davidbyttow/govips/v2/vips"
	"log"
	"os"
	"time"
)

func DumpDebugImage(id string, sourceImage *vips.ImageRef) error {
	if !*env.FoxyEnvironment.DebugImages {
		return nil
	}

	defer utils.TrackTime(time.Now(), "Dump Debug Image")

	filename := fmt.Sprintf("./debugimg/debug-%s-%s.png", time.Now().Format("20060102150405"), id)

	copiedImage, err := sourceImage.Copy()
	if err != nil {
		log.Println(err)
		return err
	}

	data, _, err := copiedImage.ExportPng(nil)
	if err != nil {
		log.Println(err)
		return err
	}

	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}
