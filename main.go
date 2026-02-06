package main

import (
	"cinema-system/handler"
	"cinema-system/repository"
	"cinema-system/service"
	"fmt"
	"log"
	"net/http"
	"time"
)

const port = ":8080"

func main() {
	// Repository → Service → Handler (Assignment 3 architecture)
	repo := repository.NewMovieRepo()
	svc := service.NewMovieService(repo)
	movieHandler := handler.NewMovieHandler(svc)

	// At least one goroutine: background worker (e.g. heartbeat logger)
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			log.Println("[background] cinema system running")
		}
	}()

	// Routes: 3+ endpoints (list, get by id, create, update, delete = 5)
	http.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	http.Handle("/api/movies", movieHandler)
	http.Handle("/api/movies/", movieHandler)

	fmt.Println("Cinema System – Assignment 4 (Milestone 2)")
	fmt.Println("Server listening on http://localhost" + port)
	fmt.Println("  GET  /health         – health check")
	fmt.Println("  GET  /api/movies     – list movies")
	fmt.Println("  GET  /api/movies/:id – get movie")
	fmt.Println("  POST /api/movies     – create movie (JSON body)")
	fmt.Println("  PUT  /api/movies/:id – update movie")
	fmt.Println("  DELETE /api/movies/:id – delete movie")
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}
}
