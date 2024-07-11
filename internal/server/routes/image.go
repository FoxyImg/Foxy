package routes

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"foxy/internal/config"
	"foxy/internal/env"
	"foxy/internal/ffmpeg"
	"foxy/internal/params"
	"foxy/internal/server/middleware"
	"foxy/internal/storage"
	"foxy/internal/utils"
	"foxy/internal/vision"
	"github.com/davidbyttow/govips/v2/vips"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func RegisterImageRoutes(router chi.Router) {
	router.Group(func(router chi.Router) {
		router.Use(middleware.CorsHeaders)

		router.Options("/{accessKey}/{source}/{params...}", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		router.Options("/{accessKey}/{source...}", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		router.Get("/{accessKey}/*", GetImageHandler)
	})
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
	defer utils.TrackTime(time.Now(), "Handle Images Route: "+r.URL.Path+"?"+r.URL.RawQuery)

	w.Header().Set("Access-Control-Allow-Origin", "*")

	ffmpegUtility := ffmpeg.NewFfmpeg(nil, nil)

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

	imgixMode := utils.IfNil(sourceConfig.ImgixMode, false)

	var source string
	if imgixMode {
		source = strings.Join(parts[2:], "/")
		if strings.Contains(source, "%") {
			unescape, unescapeErr := url.QueryUnescape(source)
			if unescapeErr == nil {
				source = unescape
			}
		}
	} else {
		for len(parts[2])%4 != 0 {
			parts[2] += "="
		}
		sourceBytes, err := base64.URLEncoding.DecodeString(parts[2])
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		source = string(sourceBytes)
	}

	isVideo := ffmpegUtility.IsVideo(source)

	var checkSig = true
	var imageParams *params.ImageParams

	//TODO: Preset handling
	//if len(parts) >= 4 && strings.HasPrefix(parts[3], "@") {
	//	p, version, paramsErr := params.FetchPreset(sourceConfig.AppId, parts[3][1:])
	//	if paramsErr != nil {
	//		w.WriteHeader(http.StatusBadRequest)
	//		return
	//	}
	//
	//	checkSig = false
	//	imageParams = p
	//	parts = append(parts, "version:"+strconv.Itoa(version))
	//}

	if imgixMode {
		p, paramsErr := params.BuildParamsFromQuery(r.URL.Query())
		if paramsErr != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		imageParams = p
	} else {
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

		if imgixMode {
			if !utils.VerifySignatureFromQuery(*sourceConfig.Secret, source, r.URL.Query()) {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
		} else {
			if !utils.VerifySignature(*sourceConfig.Secret, r.URL.Query().Get("s"), strings.TrimRight(r.URL.Path, "/")) {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
		}
	}

	if r.URL.Query().Has("showpreset") {
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
		cached, _ := storage.GetCachedResult(sourceConfig, sourceId, source, imageParams, utils.IfNil(imageParams.Export.Format, "jpg"))
		if cached != nil {
			http.ServeFile(w, r, *cached)
			return
		}
	}

	var img *vips.ImageRef
	var videoMeta *vision.VideoMetadata
	if isVideo {
		var ffmeta *ffmpeg.Meta
		ffmeta, source, img, err = imageParams.Video.ProcessFrame(source, sourceId, sourceConfig, nil, imageParams, nil)
		if ffmeta != nil {
			videoMeta = ffmeta.GetVideoMetadata()
		}
	} else {
		img, err = storage.GetSourceImage(sourceConfig, sourceId, source, imageParams.Debug != nil && imageParams.Debug.DisableSourceCache)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)
			return
		}
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

	if meta != nil && videoMeta != nil {
		meta.Video = videoMeta
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

	log.Println("Set Cached Result", source)
	_ = storage.SetCachedResult(sourceConfig, sourceId, source, imageParams, utils.IfNil(imageParams.Export.Format, "jpg"), buffer)

	if *env.FoxyEnvironment.CacheTTL > 0 {
		expires := time.Now().Add(*env.FoxyEnvironment.CacheTTL)
		w.Header().Add("Expires", strings.Replace(expires.Format(time.RFC1123), "UTC", "GMT", -1))
		w.Header().Set("Cache-Control", fmt.Sprintf("public, s-maxage=%d, max-age=%d, no-transform", *env.FoxyEnvironment.CacheTTL, *env.FoxyEnvironment.CacheTTL))
	} else {
		w.Header().Set("Cache-Control", "private, no-cache, no-store, must-revalidate")
	}

	sendImageResult(w, utils.IfNil(imageParams.Export.Format, "jpg"), buffer)
}
