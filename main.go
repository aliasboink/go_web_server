package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type apiConfig struct {
	fileserverHits int
}

func main() {
	r := chi.NewRouter()
	apiRouter := chi.NewRouter()
	adminRouter := chi.NewRouter()

	const filepathRoot = "."
	const port = "8080"

	debug := flag.Bool("debug", false, "Debug the program (deletes DB).")
	flag.Parse()
	if *debug {
		deleteDatabase("database.json")
	}

	apiCfg := apiConfig{
		fileserverHits: 0,
	}

	r.Handle("/app", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))))
	r.Handle("/app/*", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))))
	apiRouter.Get("/healthz", handlerReadiness)
	apiRouter.Get("/reset", apiCfg.handlerReset)
	apiRouter.Post("/chirps", handlerPostChirp)
	apiRouter.Get("/chirps", handlerGetChirps)
	apiRouter.Get("/chirps/{id}", handlerGetChirpWithId)
	apiRouter.Post("/users", handlerPostUser)
	adminRouter.Get("/metrics", apiCfg.handlerMetrics)

	r.Mount("/api", apiRouter)
	r.Mount("/admin", adminRouter)

	corsMux := middlewareCors(r)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: corsMux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
}

func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	myMap := map[string]interface{}{"Hits": fmt.Sprintf("%d", cfg.fileserverHits)}
	outputHTML(w, "admin/metrics.html", myMap)
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits++
		next.ServeHTTP(w, r)
	})
}
