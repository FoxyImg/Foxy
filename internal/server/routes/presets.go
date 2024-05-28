package routes

import (
	"context"
	"encoding/json"
	"foxy/internal/db"
	"foxy/internal/metadata"
	"foxy/internal/server/middleware"
	"github.com/gosimple/slug"
	"github.com/jackc/pgx/v5"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
)

func RegisterPresetRoutes() {
	http.Handle("GET /presets/{appId}", middleware.VerifyAuth(http.HandlerFunc(GetPresetsHandler)))
	http.Handle("POST /presets/{appId}/{presetName}", middleware.VerifyAuth(http.HandlerFunc(PostNewPresetHandler)))
	http.Handle("PUT /presets/{appId}/{presetName}", middleware.VerifyAuth(http.HandlerFunc(PutUpdatePresetHandler)))
	http.Handle("DELETE /presets/{appId}/{presetName}", middleware.VerifyAuth(http.HandlerFunc(DeletePresetHandler)))
}

func GetPresetsHandler(w http.ResponseWriter, r *http.Request) {
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

	var presets = make(map[string]metadata.ImageParams)
	var name string
	var presetJSON string
	_, err = pgx.ForEachRow(res, []any{&name, &presetJSON}, func() error {
		var params metadata.ImageParams
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

	_, _, err = metadata.FetchPreset(appId, presetName)
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

	_, _, err = metadata.FetchPreset(appId, presetName)
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
	delete(metadata.PresetCache, appId+":"+presetName)
	mutex.Unlock()

	_ = db.RedisDelete(appId + ":" + presetName)

	_, _, err = metadata.FetchPreset(appId, presetName)
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
	delete(metadata.PresetCache, appId+":"+presetName)
	mutex.Unlock()

	_ = db.RedisDelete(appId + ":" + presetName)

}
