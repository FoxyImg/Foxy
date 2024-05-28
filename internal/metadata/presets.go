package metadata

import (
	"context"
	"encoding/json"
	"foxy/internal/db"
	"log"
	"sync"
	"time"
)

type Preset struct {
	Version int         `json:"version"`
	Params  ImageParams `json:"params"`
}

var PresetCache = make(map[string]*Preset)

func FetchPreset(appId string, presetId string) (*ImageParams, int, error) {
	presetKey := appId + ":" + presetId
	if PresetCache[presetKey] != nil {
		return &PresetCache[presetKey].Params, PresetCache[presetKey].Version, nil
	}

	presetConfigJSON, err := db.RedisGet(presetKey)
	if presetConfigJSON != nil {
		var preset Preset
		err = json.Unmarshal([]byte(*presetConfigJSON), &preset)
		if err == nil {
			var mutex = &sync.Mutex{}
			mutex.Lock()
			PresetCache[presetKey] = &preset
			mutex.Unlock()

			return &preset.Params, preset.Version, nil
		}
	}

	pg, err := db.NewClient()
	if err != nil {
		return nil, 0, err
	}

	conn, err := pg.Acquire(context.Background())
	if err != nil {
		return nil, 0, err
	}

	var presetParamsJSON string
	var version int
	res := conn.QueryRow(context.Background(), "SELECT preset, version FROM presets where app_id = $1 and name = $2", appId, presetId)
	err = res.Scan(&presetParamsJSON, &version)
	if err != nil {
		return nil, 0, err
	}

	conn.Release()

	var params = *NewImageParams()
	err = json.Unmarshal([]byte(presetParamsJSON), &params)
	if err != nil {
		return nil, 0, err
	}

	preset := Preset{
		Version: version,
		Params:  params,
	}

	presetJSON, err := json.Marshal(preset)
	if err != nil {
		return nil, 0, err
	}

	err = db.RedisSet(presetKey, string(presetJSON), 60*time.Minute)
	if err != nil {
		log.Println("RedisSet Error:", err)
	}

	var mutex = &sync.Mutex{}
	mutex.Lock()
	PresetCache[presetKey] = &preset
	mutex.Unlock()

	return &preset.Params, preset.Version, nil
}
