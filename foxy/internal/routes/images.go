package routes

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"foxy/internal/aws"
	"foxy/internal/db"
	"foxy/internal/process/images"
	"foxy/internal/utils"
	"log"
	"net/http"
	"strings"
	"time"
)

func sendImageResult(w http.ResponseWriter, format string, buffer *[]byte) {
	if format == "webp" {
		w.Header().Set("Content-Type", "image/webp")
	} else if format == "jpeg" {
		w.Header().Set("Content-Type", "image/jpeg")
	} else if format == "jpg" {
		w.Header().Set("Content-Type", "image/jpeg")
	} else if format == "png" {
		w.Header().Set("Content-Type", "image/png")
	} else if format == "avif" {
		w.Header().Set("Content-Type", "image/avif")
	}

	w.Write(*buffer)
}

func HandleImagesRoute(w http.ResponseWriter, r *http.Request) {
	defer utils.TrackTime(time.Now(), "Handle Images Route")

	w.Header().Set("Access-Control-Allow-Origin", "*")

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 2 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	accessKey := parts[1]

	source, err := base64.URLEncoding.DecodeString(parts[2])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	params, err := images.BuildParams(parts[3:])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !params.DisableRenderCache {
		cached, _ := utils.GetCachedResult(accessKey, string(source), parts[3:], params.ExportParams.Format)
		if cached != nil {
			sendImageResult(w, params.ExportParams.Format, cached)
			return
		}
	}

	sourceConfig, err := db.GetSourceConfig(accessKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if sourceConfig.SourceType != "s3" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	url, err := aws.GetSignedUrl(*sourceConfig, string(source), time.Hour*1)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	img, err := utils.GetCachedSource(accessKey, string(source), url, params.DisableSourceCache)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	if img == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	paramsJSON, err := json.Marshal(params)
	if err != nil {
		fmt.Println("Marshal JSON Error: ", err)
		return
	}

	log.Println(string(paramsJSON))

	buffer, meta, err := images.ProcessImage(*sourceConfig, accessKey, string(source), params, img)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	if params.MetaOnly {
		w.Header().Set("Content-Type", "application/json")
		metaJSON, err := json.Marshal(meta)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)
			return
		}

		w.Write(metaJSON)
		return
	}

	_ = utils.SetCachedResult(accessKey, string(source), parts[3:], params.ExportParams.Format, buffer)

	sendImageResult(w, params.ExportParams.Format, buffer)
}
