package routes

import (
	"context"
	"encoding/json"
	"foxy/internal/db"
	"foxy/internal/server/middleware"
	"github.com/jackc/pgx/v5"
	"log"
	"net/http"
)

type sourceInfo struct {
	Type         string   `json:"type"`
	Name         string   `json:"name"`
	Key          string   `json:"key"`
	SampleImages []string `json:"sampleImages"`
}

func RegisterSourceRoutes(mux *http.ServeMux) {
	mux.Handle("OPTIONS /sources/{appId}", middleware.CorsHeaders(middleware.CorsDefaultHandler()))

	mux.Handle("GET /sources/{appId}", middleware.VerifyAuth(
		middleware.CorsHeaders(
			http.HandlerFunc(GetSourcesHandler),
		),
	))
}

func GetSourcesHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Get sources handler", r.Method, r.URL.Path)

	conn, err := db.NewConnection()
	if err != nil {
		log.Println("New Connection Error:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer conn.Release()

	appId := r.PathValue("appId")
	res, err := conn.Query(context.Background(), "SELECT name, type, sid FROM sources where app_id = $1", appId)
	if err != nil {
		log.Println("Query Error:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer res.Close()

	var sources []sourceInfo
	var name string
	var sourceType string
	var sourceKey string
	_, _ = pgx.ForEachRow(res, []any{&name, &sourceType, &sourceKey}, func() error {
		sources = append(sources, sourceInfo{
			Name:         name,
			Type:         sourceType,
			Key:          sourceKey,
			SampleImages: []string{},
		})

		return nil
	})

	sourcesJSON, sourcesJSONError := json.Marshal(sources)
	if sourcesJSONError != nil {
		log.Println("Marshal JSON Error:", sourcesJSONError)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(sourcesJSON)
}
