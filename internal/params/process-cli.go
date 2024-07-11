package params

import (
	"bytes"
	"encoding/json"
	"fmt"
	"foxy/internal/utils"
	"foxy/internal/vision"
	"log"
	"os"
	"os/exec"
)

func ProcessImageCLI(
	sourceId string,
	key string,
	params *ImageParams,
) (*[]byte, *vision.Metadata, error) {
	foxyCLI, err := os.Executable()
	if err != nil {
		log.Println(err)
		return nil, nil, err
	}

	tempImage, err := utils.TempFileName("tmpimage-")
	if err != nil {
		log.Println(err)
		return nil, nil, err
	}

	tempMeta, err := utils.TempFileName("tmpmeta-")
	if err != nil {
		log.Println(err)
		return nil, nil, err
	}

	tempParams, err := os.CreateTemp("", "tmpparams-")
	if err != nil {
		log.Println(err)
		return nil, nil, err
	}

	paramsJSON, err := json.Marshal(params)
	if err != nil {
		log.Println(err)
		return nil, nil, err
	}

	_, _ = tempParams.Write(paramsJSON)
	_ = tempParams.Close()
	defer os.Remove(tempParams.Name())

	cmd := exec.Command(foxyCLI, "process", sourceId, sourceId, key, tempParams.Name(), *tempImage, *tempMeta)
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	if err = cmd.Run(); err != nil {
		fmt.Println("out:", outb.String(), "err:", errb.String())
		log.Println(err)
		return nil, nil, err
	}
	fmt.Println("out:", outb.String(), "err:", errb.String())

	meta := vision.Metadata{}
	_, err = os.Stat(*tempMeta)
	if err == nil {
		metaJSON, err := os.ReadFile(*tempMeta)
		if err == nil {
			err = json.Unmarshal(metaJSON, &meta)
			if err != nil {
				log.Println(err)
				return nil, nil, err
			}
		}

		_ = os.Remove(*tempMeta)
	}

	if params.MetaOnly {
		_ = os.Remove(*tempImage)
		return nil, &meta, nil
	}

	_, err = os.Stat(*tempImage)
	if err != nil {
		log.Println(err)
		return nil, nil, err
	}

	img, err := os.ReadFile(*tempImage)
	_ = os.Remove(*tempImage)

	return &img, &meta, nil
}
