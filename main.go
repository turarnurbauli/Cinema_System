package main

import (
	"context"
	"cinema-system/handler"
	"cinema-system/repository"
	"cinema-system/service"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const port = ":8080"

func main() {
	// Подключение к MongoDB Atlas (или локальной MongoDB) через переменную окружения MONGODB_URI.
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		log.Fatal("MONGODB_URI is not set; please configure a MongoDB Atlas connection string")
	}

	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("failed to connect to MongoDB: %v", err)
	}
	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			log.Printf("error disconnecting MongoDB client: %v", err)
		}
	}()

	// Имя базы можно переопределить через переменную окружения MONGODB_DB, по умолчанию "cinema".
	dbName := os.Getenv("MONGODB_DB")
	if dbName == "" {
		dbName = "cinema"
	}

	repo, err := repository.NewMovieRepo(ctx, client, dbName)
	if err != nil {
		log.Fatalf("failed to initialise movie repository: %v", err)
	}

	// Repository → Service → Handler (Assignment 3 architecture)
	svc := service.NewMovieService(repo)
	movieHandler := handler.NewMovieHandler(svc)

	// Simple HTML homepage so root "/" is not empty.
	http.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <title>Cinema System – Movies</title>
  <style>
    body { font-family: system-ui, sans-serif; margin: 0; padding: 24px; background: #0b1120; color: #e5e7eb; }
    h1 { margin-top: 0; color: #facc15; }
    .card { background: #020617; border-radius: 12px; padding: 16px 20px; margin-bottom: 12px; border: 1px solid #1f2937; }
    .badge { display: inline-block; padding: 2px 8px; border-radius: 999px; font-size: 12px; background: #1d4ed8; color: #e5e7eb; }
    .muted { color: #9ca3af; font-size: 14px; }
    .layout { display: grid; grid-template-columns: 2fr 1fr; gap: 24px; align-items: flex-start; }
    @media (max-width: 800px) { .layout { grid-template-columns: 1fr; } }
    label { display: block; margin-top: 8px; font-size: 14px; }
    input, textarea { width: 100%; padding: 6px 8px; border-radius: 6px; border: 1px solid #4b5563; background: #020617; color: #e5e7eb; font-size: 14px; }
    button { margin-top: 12px; padding: 8px 14px; border-radius: 999px; border: none; background: #22c55e; color: #022c22; font-weight: 600; cursor: pointer; }
    button:disabled { opacity: 0.5; cursor: default; }
    .chips { margin-top: 4px; font-size: 13px; color: #9ca3af; }
    .endpoint { font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace; font-size: 12px; }
    .actions { margin-top: 8px; display: flex; gap: 8px; flex-wrap: wrap; }
    .btn-secondary { background: #0ea5e9; color: #022c22; }
    .btn-danger { background: #ef4444; color: #fef2f2; }
  </style>
</head>
<body>
  <h1>Cinema System – Assignment 4</h1>
  <p class="muted">Simple UI on top of JSON API. Data is stored in MongoDB (Atlas/local).</p>

  <div class="layout">
    <section>
      <h2>Movies</h2>
      <p class="muted">Loaded from <span class="endpoint">GET /api/movies</span>.</p>
      <div id="movies"></div>
    </section>
    <aside>
      <h2>Add movie</h2>
      <form id="movie-form">
        <label>Title <input name="title" required /></label>
        <label>Description <textarea name="description" rows="2"></textarea></label>
        <label>Duration (minutes) <input name="duration" type="number" min="1" required /></label>
        <label>Genre <input name="genre" required /></label>
        <label>Rating (0–10) <input name="rating" type="number" min="0" max="10" step="0.1" required /></label>
        <button type="submit">Create movie</button>
        <div id="status" class="chips"></div>
      </form>

      <h3>API endpoints</h3>
      <ul class="chips">
        <li><span class="endpoint">GET /health</span> – health check</li>
        <li><span class="endpoint">GET /api/movies</span> – list movies</li>
        <li><span class="endpoint">GET /api/movies/:id</span> – get movie</li>
        <li><span class="endpoint">POST /api/movies</span> – create (JSON)</li>
        <li><span class="endpoint">PUT /api/movies/:id</span> – update</li>
        <li><span class="endpoint">DELETE /api/movies/:id</span> – delete</li>
      </ul>
    </aside>
  </div>

  <script>
    const moviesEl = document.getElementById("movies");
    const form = document.getElementById("movie-form");
    const statusEl = document.getElementById("status");

    async function loadMovies() {
      moviesEl.innerHTML = "<p class='muted'>Loading...</p>";
      try {
        const res = await fetch("/api/movies");
        if (!res.ok) throw new Error("HTTP " + res.status);
        const data = await res.json();
        if (!Array.isArray(data) || data.length === 0) {
          moviesEl.innerHTML = "<p class='muted'>No movies yet. Add one on the right.</p>";
          return;
        }
        moviesEl.innerHTML = "";
        data.forEach(m => {
          const div = document.createElement("div");
          div.className = "card";
          div.innerHTML =
            "<div class='badge'>#" + m.id + "</div>" +
            "<h3>" + (m.title || "(no title)") + "</h3>" +
            "<p class='muted'>" + (m.description || "No description") + "</p>" +
            "<p class='chips'>Genre: " + (m.genre || "-") + " · Duration: " + (m.duration || 0) + " min · Rating: " + (m.rating ?? "-") + "</p>" +
            "<div class='actions'>" +
              "<button data-action='edit' data-id='" + m.id + "' class='btn-secondary'>Edit movie</button>" +
              "<button data-action='delete' data-id='" + m.id + "' class='btn-danger'>Delete</button>" +
            "</div>";
          moviesEl.appendChild(div);
        });

        moviesEl.querySelectorAll("button[data-action='delete']").forEach(btn => {
          btn.addEventListener("click", async () => {
            const id = btn.getAttribute("data-id");
            if (!confirm("Delete movie #" + id + "?")) return;
            try {
              const res = await fetch("/api/movies/" + id, { method: "DELETE" });
              if (!res.ok && res.status !== 204) throw new Error("HTTP " + res.status);
              loadMovies();
            } catch {
              alert("Failed to delete movie");
            }
          });
        });

        moviesEl.querySelectorAll("button[data-action='edit']").forEach(btn => {
          btn.addEventListener("click", async () => {
            const id = btn.getAttribute("data-id");
            try {
              // Fetch current movie
              const resGet = await fetch("/api/movies/" + id);
              if (!resGet.ok) throw new Error("HTTP " + resGet.status);
              const movie = await resGet.json();

              const newTitle = prompt("Title:", movie.title || "");
              if (newTitle === null) return;
              const newDescription = prompt("Description:", movie.description || "");
              if (newDescription === null) return;
              const newDurationStr = prompt("Duration (minutes):", String(movie.duration || 0));
              if (newDurationStr === null) return;
              const newDuration = Number(newDurationStr);
              if (Number.isNaN(newDuration) || newDuration <= 0) {
                alert("Invalid duration");
                return;
              }
              const newGenre = prompt("Genre:", movie.genre || "");
              if (newGenre === null) return;
              const newRatingStr = prompt("Rating (0-10):", String(movie.rating ?? ""));
              if (newRatingStr === null) return;
              const newRating = Number(newRatingStr);
              if (Number.isNaN(newRating) || newRating < 0 || newRating > 10) {
                alert("Invalid rating");
                return;
              }

              movie.title = newTitle;
              movie.description = newDescription;
              movie.duration = newDuration;
              movie.genre = newGenre;
              movie.rating = newRating;

              const resPut = await fetch("/api/movies/" + id, {
                method: "PUT",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(movie)
              });
              if (!resPut.ok) throw new Error("HTTP " + resPut.status);
              loadMovies();
            } catch {
              alert("Failed to update movie");
            }
          });
        });
      } catch (e) {
        moviesEl.innerHTML = "<p class='muted'>Failed to load movies.</p>";
      }
    }

    form.addEventListener("submit", async (e) => {
      e.preventDefault();
      const formData = new FormData(form);
      const movie = {
        title: formData.get("title"),
        description: formData.get("description"),
        duration: Number(formData.get("duration")),
        genre: formData.get("genre"),
        rating: Number(formData.get("rating"))
      };
      statusEl.textContent = "Saving...";
      try {
        const res = await fetch("/api/movies", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(movie)
        });
        if (!res.ok) throw new Error("HTTP " + res.status);
        await res.json();
        statusEl.textContent = "Movie created.";
        form.reset();
        loadMovies();
      } catch (e) {
        statusEl.textContent = "Failed to create movie.";
      }
    });

    loadMovies();
  </script>
</body>
</html>`))
	})

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
