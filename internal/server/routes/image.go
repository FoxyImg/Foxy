package routes

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"foxy/internal/config"
	"foxy/internal/env"
	"foxy/internal/params"
	"foxy/internal/storage"
	"foxy/internal/utils"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func RegisterImageRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /{accessKey}/{source}/{params...}", GetImageHandler)
	mux.HandleFunc("GET /{accessKey}/{source}", GetImageHandler)
}

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

func GetImageHandler(w http.ResponseWriter, r *http.Request) {
	defer utils.TrackTime(time.Now(), "Handle Images Route: "+r.URL.Path)

	w.Header().Set("Access-Control-Allow-Origin", "*")

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 2 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	sourceId := parts[1]

	sourceConfig, err := config.GetSourceConfigFromCache(sourceId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	for len(parts[2])%4 != 0 {
		parts[2] += "="
	}
	source, err := base64.URLEncoding.DecodeString(parts[2])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var checkSig = true
	var imageParams *params.ImageParams

	if len(parts) >= 4 && strings.HasPrefix(parts[3], "@") {
		p, version, paramsErr := params.FetchPreset(sourceConfig.AppId, parts[3][1:])
		if paramsErr != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		checkSig = false
		imageParams = p
		parts = append(parts, "version:"+strconv.Itoa(version))
	}

	if imageParams == nil {
		p, paramsErr := params.BuildParams(parts[3:])
		if paramsErr != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		imageParams = p
	}

	if checkSig && env.FoxyEnvironment.RequireSignatureValidation {
		if sourceConfig.Secret == nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if !r.URL.Query().Has("s") {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if !utils.VerifySignature(*sourceConfig.Secret, r.URL.Query().Get("s"), strings.TrimRight(r.URL.Path, "/")) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
	}

	if r.URL.Query().Has("preset") {
		paramsJSON, jsonErr := json.Marshal(imageParams)
		if jsonErr != nil {
			fmt.Println("Marshal JSON Error: ", jsonErr)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(paramsJSON)
		return
	}

	if imageParams.Debug != nil && !imageParams.Debug.DisableRenderCache {
		cached, _ := storage.GetCachedResult(sourceId, string(source), parts[3:], utils.IfNil(imageParams.Export.Format, "jpg"))
		if cached != nil {
			sendImageResult(w, utils.IfNil(imageParams.Export.Format, "jpg"), cached)
			return
		}
	}

	img, err := storage.GetSourceImage(sourceConfig, sourceId, string(source), imageParams.Debug != nil && imageParams.Debug.DisableSourceCache)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	if img == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	buffer, meta, err := params.ProcessImage(sourceId, sourceConfig, sourceId, string(source), imageParams, img)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	if imageParams.MetaOnly {
		metaJSON, err := json.Marshal(meta)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(metaJSON)
		return
	}

	_ = storage.SetCachedResult(sourceId, string(source), parts[3:], utils.IfNil(imageParams.Export.Format, "jpg"), buffer)

	sendImageResult(w, utils.IfNil(imageParams.Export.Format, "jpg"), buffer)
}
