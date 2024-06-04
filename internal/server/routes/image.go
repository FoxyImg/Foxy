package routes

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"foxy/internal/config"
	"foxy/internal/env"
	"foxy/internal/process/images"
	"foxy/internal/storage"
	"foxy/internal/utils"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func RegisterImageRoutes() {
	http.HandleFunc("GET /{accessKey}/{source}/{params...}", GetImageHandler)
	http.HandleFunc("GET /{accessKey}/{source}", GetImageHandler)
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

	accessKey := parts[1]

	sourceConfig, err := config.GetSourceConfigFromCache(accessKey)
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
	var params *images.ImageParams

	if len(parts) >= 4 && strings.HasPrefix(parts[3], "@") {
		p, version, paramsErr := images.FetchPreset(*sourceConfig.AppId, parts[3][1:])
		if paramsErr != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		checkSig = false
		params = p
		parts = append(parts, "version:"+strconv.Itoa(version))
	}

	if params == nil {
		p, paramsErr := images.BuildParams(parts[3:])
		if paramsErr != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		params = p
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
		paramsJSON, jsonErr := json.Marshal(params)
		if jsonErr != nil {
			fmt.Println("Marshal JSON Error: ", jsonErr)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(paramsJSON)
		return
	}

	if !params.Debug.DisableRenderCache {
		cached, _ := storage.GetCachedResult(accessKey, string(source), parts[3:], utils.IfNil(params.Export.Format, "jpg"))
		if cached != nil {
			sendImageResult(w, utils.IfNil(params.Export.Format, "jpg"), cached)
			return
		}
	}

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

	buffer, meta, err := images.ProcessImage(*sourceConfig, accessKey, string(source), params, img)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	if params.MetaOnly {
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

	_ = storage.SetCachedResult(accessKey, string(source), parts[3:], utils.IfNil(params.Export.Format, "jpg"), buffer)

	sendImageResult(w, utils.IfNil(params.Export.Format, "jpg"), buffer)
}
