package utils

import (
	"os"
)

func TempFileName(prefix string) (*string, error) {
	tempImage, err := os.CreateTemp("", prefix)
	if err != nil {
		return nil, err
	}

	tempName := tempImage.Name()

	_ = tempImage.Close()
	_ = os.Remove(tempImage.Name())

	return &tempName, nil
}
