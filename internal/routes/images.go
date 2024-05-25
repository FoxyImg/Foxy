package routes

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"foxy/internal/config"
	"foxy/internal/metadata"
	"foxy/internal/process/images"
	"foxy/internal/storage"
	"foxy/internal/utils"
	"log"
	"net/http"
	"os"
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

	_, _ = w.Write(*buffer)
}

func HandleImagesRoute(w http.ResponseWriter, r *http.Request) {
	defer utils.TrackTime(time.Now(), "Handle Images Route: "+r.URL.Path)

	w.Header().Set("Access-Control-Allow-Origin", "*")

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 2 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	accessKey := parts[1]

	sourceConfig, err := config.GetSourceConfigFromCache(accessKey)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if os.Getenv("REQUIRE_SIG_VALIDATION") == "true" {
		if sourceConfig.Secret == nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if !r.URL.Query().Has("s") {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if !utils.VerifySignature(*sourceConfig.Secret, r.URL.Query().Get("s"), r.URL.Path) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
	}

	source, err := base64.URLEncoding.DecodeString(parts[2])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	params, err := metadata.BuildParams(parts[3:])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !params.Debug.DisableRenderCache {
		cached, _ := storage.GetCachedResult(accessKey, string(source), parts[3:], params.ExportParams.Format)
		if cached != nil {
			sendImageResult(w, params.ExportParams.Format, cached)
			return
		}
	}

	//if sourceConfig.Source.Type != "s3" {
	//	w.WriteHeader(http.StatusBadRequest)
	//	return
	//}

	//url, err := aws.GetSignedUrl(*sourceConfig, string(source), time.Hour*1)
	//if err != nil {
	//	w.WriteHeader(http.StatusInternalServerError)
	//	log.Println(err)
	//	return
	//}

	//img, err := storage.GetCachedSource(accessKey, string(source), url, params.Debug.DisableSourceCache)
	//if err != nil {
	//	w.WriteHeader(http.StatusInternalServerError)
	//	log.Println(err)
	//	return
	//}

	img, err := storage.GetSourceImage(sourceConfig, accessKey, string(source), params.Debug.DisableSourceCache)
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

		_, _ = w.Write(metaJSON)
		return
	}

	_ = storage.SetCachedResult(accessKey, string(source), parts[3:], params.ExportParams.Format, buffer)

	sendImageResult(w, params.ExportParams.Format, buffer)
}
