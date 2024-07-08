package routes

import (
	"context"
	"encoding/json"
	"foxy/internal/db"
	"foxy/internal/env"
	"foxy/internal/params"
	"foxy/internal/server/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/gosimple/slug"
	"github.com/jackc/pgx/v5"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
)

func RegisterPresetRoutes(router chi.Router) {
	router.Group(func(router chi.Router) {
		router.Use(middleware.CorsHeaders)
		router.Use(middleware.VerifyAuth)

		router.Options("/presets/{appId}", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		router.Options("/presets/{appId}/{presetName}", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		router.Get("/presets/{appId}", GetPresetsHandler)

		if !env.FoxyEnvironment.AllowPresetManagement {
			return
		}

		if env.FoxyEnvironment.Isolated {
			router.Post("/presets/{appId}/{presetName}", PostNewIsolatedPresetHandler)
			router.Put("/presets/{appId}/{presetName}", PostNewIsolatedPresetHandler)
			router.Delete("/presets/{appId}/{presetName}", DeleteIsolatedPresetHandler)
		} else {
			router.Post("/presets/{appId}/{presetName}", PostNewPresetHandler)
			router.Put("/presets/{appId}/{presetName}", PutUpdatePresetHandler)
			router.Delete("/presets/{appId}/{presetName}", DeletePresetHandler)
		}
	})
}

func GetPresetsHandler(w http.ResponseWriter, r *http.Request) {
	if env.FoxyEnvironment.Isolated {
		presetsJSON, presetsJSONError := json.Marshal(params.PresetCache)
		if presetsJSONError != nil {
			log.Println("Marshal JSON Error:", presetsJSONError)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(presetsJSON)
		return
	}

	conn, err := db.NewConnection()
	if err != nil {
		log.Println("New Connection Error:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer conn.Release()

	appId := r.PathValue("appId")
	res, err := conn.Query(context.Background(), "SELECT name, preset FROM presets where app_id = $1", appId)
	if err != nil {
		log.Println("Query Error:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer res.Close()

	var presets = make(map[string]params.ImageParams)
	var name string
	var presetJSON string
	_, err = pgx.ForEachRow(res, []any{&name, &presetJSON}, func() error {
		var params = *params.NewImageParams()
		jsonErr := json.Unmarshal([]byte(presetJSON), &params)
		if jsonErr != nil {
			return jsonErr
		}

		presets[name] = params

		return nil
	})

	if err != nil {
		log.Println("ForEachRow Error:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	presetsJSON, presetsJSONError := json.Marshal(presets)
	if presetsJSONError != nil {
		log.Println("Marshal JSON Error:", presetsJSONError)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(presetsJSON)
}

func PostNewPresetHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := db.NewConnection()
	if err != nil {
		log.Println("New Connection Error:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer conn.Release()

	appId := r.PathValue("appId")
	presetName := slug.Make(strings.ToLower(r.PathValue("presetName")))
	presetJSON, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("Read Body Error:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	log.Println("Preset JSON:", presetJSON)

	if !json.Valid(presetJSON) {
		log.Println("Invalid JSON")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	_, err = conn.Exec(context.Background(), "INSERT INTO presets (app_id, name, preset) VALUES ($1, $2, $3)", appId, presetName, string(presetJSON))
	if err != nil {
		log.Println("Exec Error:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_, _, err = params.FetchPreset(&appId, presetName)
	if err != nil {
		log.Println("Fetch Preset Error:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func PutUpdatePresetHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := db.NewConnection()
	if err != nil {
		log.Println("New Connection Error:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer conn.Release()

	appId := r.PathValue("appId")
	presetName := slug.Make(strings.ToLower(r.PathValue("presetName")))
	presetJSON, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("Read Body Error:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !json.Valid(presetJSON) {
		log.Println("Invalid JSON")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_, _, err = params.FetchPreset(&appId, presetName)
	if err != nil {
		log.Println("Fetch Preset Error:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_, err = conn.Exec(context.Background(), "UPDATE presets SET preset = $1 WHERE app_id = $2 AND name = $3", string(presetJSON), appId, presetName)
	if err != nil {
		log.Println("Exec Error:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var mutex = &sync.Mutex{}
	mutex.Lock()
	delete(params.PresetCache, appId+":"+presetName)
	mutex.Unlock()

	_ = db.RedisDelete(appId + ":" + presetName)

	_, _, err = params.FetchPreset(&appId, presetName)
	if err != nil {
		log.Println("Fetch Preset Error:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func DeletePresetHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := db.NewConnection()
	if err != nil {
		log.Println("New Connection Error:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer conn.Release()

	appId := r.PathValue("appId")
	presetName := slug.Make(strings.ToLower(r.PathValue("presetName")))

	_, err = conn.Exec(context.Background(), "DELETE FROM presets WHERE app_id = $1 AND name = $2", appId, presetName)
	if err != nil {
		log.Println("Exec Error:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var mutex = &sync.Mutex{}
	mutex.Lock()
	delete(params.PresetCache, appId+":"+presetName)
	mutex.Unlock()

	_ = db.RedisDelete(appId + ":" + presetName)

}

func PostNewIsolatedPresetHandler(w http.ResponseWriter, r *http.Request) {
	presetName := slug.Make(strings.ToLower(r.PathValue("presetName")))
	presetJSON, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("Read Body Error:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	log.Println("Preset JSON:", presetJSON)

	if !json.Valid(presetJSON) {
		log.Println("Invalid JSON")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var newIsolatedParams params.ImageParams
	err = json.Unmarshal([]byte(presetJSON), &newIsolatedParams)
	if err != nil {
		log.Println("Invalid JSON: ", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if newIsolatedParams.SourceCrop != nil {
		newIsolatedParams.SourceCrop = nil
	}

	m := &sync.Mutex{}
	m.Lock()
	params.IsolatedPresets[presetName] = &params.Preset{
		Version: 1,
		Params:  newIsolatedParams,
	}
	m.Unlock()

	if env.FoxyEnvironment.PresetsFile != nil {
		jsonStr, err := json.MarshalIndent(&params.IsolatedPresets, "", "  ")
		if err != nil {
			log.Println("Error marshaling presets file: ", err)
		}

		err = os.WriteFile(*env.FoxyEnvironment.PresetsFile, jsonStr, 0644)
		if err != nil {
			log.Println("Error writing presets file: ", err)
		}
	}

	newParams := *params.NewImageParams()
	err = json.Unmarshal([]byte(presetJSON), &newParams)
	if err != nil {
		log.Println("Invalid JSON: ", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if newParams.SourceCrop != nil {
		newParams.SourceCrop = nil
	}

	m.Lock()
	params.PresetCache[presetName] = &params.Preset{
		Version: 1,
		Params:  newParams,
	}
	m.Unlock()

	w.WriteHeader(http.StatusOK)
}

func DeleteIsolatedPresetHandler(w http.ResponseWriter, r *http.Request) {
	presetName := slug.Make(strings.ToLower(r.PathValue("presetName")))

	var mutex = &sync.Mutex{}
	mutex.Lock()
	delete(params.PresetCache, presetName)
	delete(params.IsolatedPresets, presetName)
	mutex.Unlock()

	if env.FoxyEnvironment.PresetsFile != nil {
		jsonStr, err := json.MarshalIndent(&params.IsolatedPresets, "", "  ")
		if err != nil {
			log.Println("Error marshaling presets file: ", err)
		}

		err = os.WriteFile(*env.FoxyEnvironment.PresetsFile, jsonStr, 0644)
		if err != nil {
			log.Println("Error writing presets file: ", err)
		}
	}

	w.WriteHeader(http.StatusOK)
}
