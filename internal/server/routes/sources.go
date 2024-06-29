package routes

import (
	"context"
	"encoding/json"
	"foxy/internal/config"
	"foxy/internal/db"
	"foxy/internal/env"
	"foxy/internal/server/middleware"
	"github.com/jackc/pgx/v5"
	"github.com/sethvargo/go-limiter/httplimit"
	"log"
	"net/http"
)

type sourceInfo struct {
	Type         string   `json:"type"`
	Name         string   `json:"name"`
	Key          string   `json:"key"`
	SampleImages []string `json:"sampleImages"`
}

func RegisterSourceRoutes(mux *http.ServeMux, httpLimiter *httplimit.Middleware) {
	mux.Handle("OPTIONS /sources/{appId}", middleware.CorsHeaders(middleware.CorsDefaultHandler()))

	mux.Handle("GET /sources/{appId}", middleware.VerifyAuth(
		middleware.CorsHeaders(
			http.HandlerFunc(GetSourcesHandler),
		),
	))
}

func GetSourcesHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Get sources handler", r.Method, r.URL.Path)

	var sources []sourceInfo

	if env.FoxyEnvironment.Isolated {
		err := config.LoadSourceConfigFromJSON()
		if err != nil {
			log.Println("Error loading source config: ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		for key, _ := range *config.LoadedSources {
			sources = append(sources, sourceInfo{
				Name:         key,
				Type:         (*config.LoadedSources)[key].Source.Type,
				Key:          key,
				SampleImages: []string{},
			})
		}
	} else {
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
	}

	sourcesJSON, sourcesJSONError := json.Marshal(sources)
	if sourcesJSONError != nil {
		log.Println("Marshal JSON Error:", sourcesJSONError)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err := w.Write(sourcesJSON)
	if err != nil {
		log.Println("Write JSON Error:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
