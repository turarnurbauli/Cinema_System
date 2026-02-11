package main

import (
	"cinema-system/handler"
	"cinema-system/middleware"
	"cinema-system/repository"
	"cinema-system/service"
	"context"
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

	// JWT secret для аутентификации.
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET is not set; please configure a secret key for JWT")
	}

	// Пользователи и роли.
	userRepo := repository.NewUserRepo(ctx, client, dbName)
	userSvc := service.NewUserService(userRepo)

	// Создаём дефолтных пользователей, если их ещё нет.
	if err := userSvc.EnsureUserWithRole(ctx, "admin", "1234", "Admin", service.RoleAdmin); err != nil {
		log.Fatalf("failed to ensure default admin: %v", err)
	}
	if err := userSvc.EnsureUserWithRole(ctx, "cashier", "1234", "Cashier", service.RoleCashier); err != nil {
		log.Fatalf("failed to ensure default cashier: %v", err)
	}

	authHandler := handler.NewAuthHandler(userSvc, jwtSecret)

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
    :root {
      color-scheme: dark;
      --bg: #020617;
      --bg-soft: #0b1120;
      --bg-card: #020617;
      --bg-elevated: #030712;
      --accent: #f97316;
      --accent-soft: rgba(248, 171, 94, 0.1);
      --accent-strong: #fbbf24;
      --text: #e5e7eb;
      --muted: #9ca3af;
      --border-subtle: #1f2937;
      --danger: #ef4444;
      --success: #22c55e;
      --seat-free: #0f172a;
      --seat-selected: #22c55e;
      --seat-unavailable: #4b5563;
    }

    * { box-sizing: border-box; }
    body {
      margin: 0;
      min-height: 100vh;
      font-family: system-ui, -apple-system, BlinkMacSystemFont, "SF Pro Text", sans-serif;
      background: radial-gradient(circle at top, #111827 0, #020617 45%, #020617 100%);
      color: var(--text);
      display: flex;
      flex-direction: column;
    }

    .shell {
      max-width: 1280px;
      margin: 0 auto;
      padding: 24px 16px 32px;
      flex: 1;
      display: flex;
      flex-direction: column;
      gap: 24px;
    }

    header {
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 16px;
    }
    .brand {
      display: flex;
      align-items: center;
      gap: 12px;
    }
    .brand-logo {
      width: 40px;
      height: 40px;
      border-radius: 12px;
      background: conic-gradient(from 200deg, #f97316, #facc15, #22c55e, #0ea5e9, #a855f7, #f97316);
      display: flex;
      align-items: center;
      justify-content: center;
      box-shadow: 0 0 0 1px rgba(15, 23, 42, 0.8), 0 18px 45px rgba(15, 23, 42, 0.9);
    }
    .brand-logo span {
      font-weight: 800;
      font-size: 20px;
      color: #020617;
    }
    .brand-text h1 {
      margin: 0;
      font-size: 22px;
      font-weight: 700;
      letter-spacing: 0.03em;
    }
    .brand-text p {
      margin: 2px 0 0;
      font-size: 13px;
      color: var(--muted);
    }

    .header-actions {
      display: flex;
      align-items: center;
      gap: 8px;
    }
    .pill {
      padding: 4px 10px;
      border-radius: 999px;
      border: 1px solid rgba(148, 163, 184, 0.4);
      font-size: 11px;
      text-transform: uppercase;
      letter-spacing: 0.08em;
      color: var(--muted);
      background: rgba(15, 23, 42, 0.9);
    }
    .pill strong {
      color: var(--accent-strong);
    }

    main {
      display: grid;
      grid-template-columns: minmax(0, 3fr) minmax(0, 2.5fr);
      gap: 24px;
      align-items: flex-start;
    }
    @media (max-width: 980px) {
      main {
        grid-template-columns: minmax(0, 1fr);
      }
    }

    .panel {
      border-radius: 20px;
      background: radial-gradient(circle at top left, rgba(15, 23, 42, 0.9), #020617 60%);
      border: 1px solid rgba(15, 23, 42, 0.9);
      box-shadow:
        0 24px 60px rgba(15, 23, 42, 0.95),
        inset 0 0 0 1px rgba(15, 23, 42, 0.85);
      padding: 16px 18px 18px;
    }

    .panel-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 8px;
    }
    .panel-header h2 {
      margin: 0;
      font-size: 16px;
    }
    .panel-header p {
      margin: 0;
      font-size: 12px;
      color: var(--muted);
    }

    /* Movies list */
    .movies-list {
      display: flex;
      flex-direction: column;
      gap: 12px;
      max-height: 540px;
      overflow: auto;
      padding-right: 6px;
    }
    .movie-card {
      display: grid;
      grid-template-columns: auto minmax(0, 1fr);
      gap: 14px;
      padding: 12px 12px;
      border-radius: 16px;
      background: rgba(15, 23, 42, 0.9);
      border: 1px solid rgba(31, 41, 55, 0.9);
      cursor: pointer;
      transition: border-color 0.15s ease, transform 0.1s ease, box-shadow 0.15s ease, background 0.15s ease;
    }
    .movie-card:hover {
      border-color: rgba(248, 171, 94, 0.9);
      transform: translateY(-1px);
      box-shadow: 0 18px 40px rgba(8, 47, 73, 0.7);
      background: radial-gradient(circle at top left, rgba(30, 64, 175, 0.4), rgba(15, 23, 42, 0.9));
    }
    .movie-card.active {
      border-color: var(--accent-strong);
      box-shadow: 0 18px 45px rgba(248, 171, 94, 0.6);
      background: radial-gradient(circle at top left, rgba(248, 171, 94, 0.25), rgba(15, 23, 42, 0.95));
    }
    .movie-poster {
      width: 64px;
      height: 92px;
      border-radius: 12px;
      background: linear-gradient(135deg, #1d4ed8, #a855f7);
      display: flex;
      align-items: flex-end;
      justify-content: center;
      overflow: hidden;
      font-size: 26px;
      font-weight: 700;
      color: rgba(15, 23, 42, 0.95);
      box-shadow:
        0 14px 30px rgba(15, 23, 42, 0.9),
        0 0 0 1px rgba(15, 23, 42, 0.7);
    }
    .movie-poster span {
      width: 100%;
      text-align: center;
      padding: 6px 8px;
      background: linear-gradient(to top, rgba(248, 250, 252, 0.9), transparent);
    }
    .movie-info h3 {
      margin: 0 0 2px;
      font-size: 15px;
    }
    .movie-meta {
      font-size: 12px;
      color: var(--muted);
      margin-bottom: 6px;
    }
    .movie-tags {
      display: flex;
      flex-wrap: wrap;
      gap: 6px;
      margin-top: 4px;
    }
    .tag {
      padding: 3px 8px;
      border-radius: 999px;
      font-size: 11px;
      border: 1px solid rgba(55, 65, 81, 0.9);
      color: var(--muted);
      background: rgba(15, 23, 42, 0.9);
    }
    .tag-rating {
      border-color: rgba(250, 204, 21, 0.8);
      color: var(--accent-strong);
      background: rgba(251, 191, 36, 0.08);
    }

    /* Booking panel */
    .booking-layout {
      display: grid;
      grid-template-rows: auto auto minmax(0, 1fr) auto;
      gap: 12px;
    }
    .booking-header-main h2 {
      margin: 0 0 4px;
      font-size: 17px;
    }
    .booking-header-main p {
      margin: 0;
      font-size: 13px;
      color: var(--muted);
    }

    .schedule-row {
      display: flex;
      flex-wrap: wrap;
      gap: 8px;
      margin-top: 6px;
    }
    .chip {
      padding: 4px 10px;
      border-radius: 999px;
      border: 1px solid rgba(55, 65, 81, 0.9);
      font-size: 12px;
      color: var(--muted);
      cursor: pointer;
      background: rgba(15, 23, 42, 0.9);
      transition: border-color 0.15s ease, background 0.15s ease, color 0.15s ease;
    }
    .chip span {
      opacity: 0.85;
    }
    .chip.active {
      border-color: var(--accent-strong);
      background: var(--accent-soft);
      color: var(--accent-strong);
    }

    .hall-row {
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 12px;
      padding: 8px 10px;
      border-radius: 12px;
      background: rgba(15, 23, 42, 0.7);
      border: 1px solid rgba(30, 64, 175, 0.4);
    }
    .hall-name {
      font-size: 13px;
      color: var(--muted);
    }
    .hall-name strong {
      color: var(--accent-strong);
    }

    /* Seat map */
    .seat-map-wrapper {
      border-radius: 16px;
      background: radial-gradient(circle at top, rgba(15, 23, 42, 0.9), rgba(2, 6, 23, 1));
      border: 1px solid rgba(15, 23, 42, 0.9);
      padding: 12px 14px 14px;
      display: flex;
      flex-direction: column;
      gap: 8px;
      min-height: 220px;
    }
    .screen-label {
      text-align: center;
      font-size: 11px;
      letter-spacing: 0.2em;
      text-transform: uppercase;
      color: var(--muted);
      margin-bottom: 4px;
    }
    .screen-bar {
      height: 4px;
      border-radius: 999px;
      background: linear-gradient(to right, rgba(148, 163, 184, 0), rgba(248, 250, 252, 0.7), rgba(148, 163, 184, 0));
      margin: 0 auto 6px;
      max-width: 210px;
      box-shadow: 0 0 20px rgba(248, 250, 252, 0.25);
    }
    .seat-grid {
      display: grid;
      gap: 4px;
      justify-content: center;
    }
    .seat-row {
      display: flex;
      align-items: center;
      gap: 4px;
    }
    .seat-row-label {
      width: 16px;
      font-size: 10px;
      text-align: right;
      color: var(--muted);
      margin-right: 4px;
    }
    .seat {
      width: 18px;
      height: 18px;
      border-radius: 4px;
      border: 1px solid rgba(30, 64, 175, 0.7);
      background: var(--seat-free);
      cursor: pointer;
      position: relative;
      display: inline-flex;
      align-items: center;
      justify-content: center;
      transition: transform 0.08s ease, box-shadow 0.1s ease, background 0.12s ease, border-color 0.12s ease;
    }
    .seat:hover {
      transform: translateY(-1px);
      box-shadow: 0 8px 18px rgba(15, 23, 42, 0.9);
    }
    .seat.selected {
      background: var(--seat-selected);
      border-color: rgba(34, 197, 94, 0.95);
      box-shadow: 0 0 0 1px rgba(22, 163, 74, 0.8), 0 10px 18px rgba(22, 163, 74, 0.35);
    }
    .seat.unavailable {
      background: var(--seat-unavailable);
      border-color: rgba(75, 85, 99, 0.9);
      cursor: default;
      opacity: 0.6;
    }
    .seat span {
      font-size: 8px;
      color: #e5e7eb;
    }

    .seat-legend {
      display: flex;
      flex-wrap: wrap;
      gap: 10px;
      font-size: 11px;
      color: var(--muted);
      margin-top: 4px;
    }
    .legend-item {
      display: inline-flex;
      align-items: center;
      gap: 4px;
    }
    .legend-swatch {
      width: 12px;
      height: 12px;
      border-radius: 3px;
      border: 1px solid rgba(55, 65, 81, 0.9);
    }
    .legend-swatch.free { background: var(--seat-free); }
    .legend-swatch.selected { background: var(--seat-selected); border-color: rgba(22, 163, 74, 0.9); }
    .legend-swatch.unavailable { background: var(--seat-unavailable); }

    /* Ticket summary */
    .summary {
      display: flex;
      flex-direction: column;
      gap: 6px;
      font-size: 13px;
    }
    .summary-row {
      display: flex;
      justify-content: space-between;
      gap: 8px;
    }
    .summary-row strong {
      font-size: 14px;
    }
    .summary-list {
      max-height: 120px;
      overflow: auto;
      padding-right: 4px;
    }
    .ticket-item {
      display: grid;
      grid-template-columns: minmax(0, 1.5fr) minmax(0, 1.5fr) auto;
      gap: 6px;
      align-items: center;
      padding: 6px 8px;
      border-radius: 10px;
      background: rgba(15, 23, 42, 0.85);
      border: 1px solid rgba(31, 41, 55, 0.9);
      margin-bottom: 4px;
    }
    .ticket-item small {
      display: block;
      font-size: 11px;
      color: var(--muted);
    }
    .ticket-type {
      width: 100%;
      border-radius: 999px;
      border: 1px solid rgba(55, 65, 81, 0.9);
      background: #020617;
      color: var(--text);
      font-size: 12px;
      padding: 4px 8px;
    }
    .ticket-price {
      font-size: 13px;
      color: var(--accent-strong);
      text-align: right;
    }

    button {
      border-radius: 999px;
      border: none;
      padding: 8px 14px;
      font-size: 13px;
      font-weight: 600;
      cursor: pointer;
      background: var(--success);
      color: #022c22;
      display: inline-flex;
      align-items: center;
      gap: 6px;
      box-shadow: 0 14px 35px rgba(16, 185, 129, 0.35);
    }
    button:disabled {
      opacity: 0.5;
      cursor: default;
      box-shadow: none;
    }
    .btn-outline {
      background: transparent;
      color: var(--muted);
      border: 1px solid rgba(75, 85, 99, 0.9);
      box-shadow: none;
    }

    footer {
      max-width: 1280px;
      margin: 0 auto;
      padding: 0 16px 20px;
      font-size: 11px;
      color: var(--muted);
      display: flex;
      justify-content: space-between;
      gap: 8px;
    }
  </style>
</head>
<body>
  <div class="shell">
    <header>
      <div class="brand">
        <div class="brand-logo"><span>CS</span></div>
        <div class="brand-text">
          <h1>Cinema System</h1>
          <p>Онлайн‑выбор фильма, зала и мест — как в настоящем кинотеатре.</p>
        </div>
      </div>
      <div class="header-actions">
        <div class="pill"><strong>Assignment 4</strong> · ADP‑2</div>
        <div class="pill">Backend: Go · DB: MongoDB Atlas</div>
      </div>
    </header>

    <main>
      <section class="panel">
        <div class="panel-header">
          <div>
            <h2>Афиша фильмов</h2>
            <p>Загружается из <code>/api/movies</code>. Нажми на фильм, чтобы выбрать сеанс.</p>
          </div>
        </div>
        <div id="movies" class="movies-list"></div>
    </section>

      <section class="panel">
        <div class="booking-layout">
          <div class="booking-header">
            <div class="booking-header-main">
              <h2 id="booking-title">Выберите фильм</h2>
              <p id="booking-subtitle">После выбора фильма появятся доступные залы и время сеанса.</p>
            </div>
          </div>

          <div>
            <div class="hall-row">
              <div class="hall-name">
                Зал: <strong id="hall-label">–</strong>
                <small id="time-label" style="display:block;margin-top:2px;color:var(--muted);font-size:11px;">Время не выбрано</small>
              </div>
              <div class="schedule-row" id="schedule-row"></div>
            </div>
          </div>

          <div class="seat-map-wrapper">
            <div>
              <div class="screen-label">Экран</div>
              <div class="screen-bar"></div>
            </div>
            <div id="seat-grid" class="seat-grid"></div>
            <div class="seat-legend">
              <div class="legend-item">
                <div class="legend-swatch free"></div><span>Свободное место</span>
              </div>
              <div class="legend-item">
                <div class="legend-swatch selected"></div><span>Выбрано</span>
              </div>
              <div class="legend-item">
                <div class="legend-swatch unavailable"></div><span>Занято (для демо не используется)</span>
              </div>
            </div>
          </div>

          <div class="summary">
            <div class="summary-row">
              <div>
                <strong>Ваши билеты</strong>
                <div id="summary-caption" class="summary-caption" style="font-size:11px;color:var(--muted);margin-top:2px;">
                  Выберите места в зале: щёлкните по креслам на схеме.
                </div>
              </div>
              <div>
                <div class="ticket-price" id="total-price">0 ₸</div>
              </div>
            </div>
            <div id="tickets-list" class="summary-list"></div>
            <div class="summary-row">
              <button id="btn-book" disabled>Забронировать (демо)</button>
              <button id="btn-clear" class="btn-outline" type="button">Сбросить выбор</button>
            </div>
          </div>
        </div>
      </section>
    </main>
  </div>

  <footer>
    <span>Demo UI · Нет реальной оплаты, только учебный выбор билетов.</span>
    <span>Team: Alkhan Almas &amp; Nurbauli Turar</span>
  </footer>

  <script>
    const moviesEl = document.getElementById("movies");
    const bookingTitle = document.getElementById("booking-title");
    const bookingSubtitle = document.getElementById("booking-subtitle");
    const seatGridEl = document.getElementById("seat-grid");
    const scheduleRowEl = document.getElementById("schedule-row");
    const hallLabelEl = document.getElementById("hall-label");
    const timeLabelEl = document.getElementById("time-label");
    const ticketsListEl = document.getElementById("tickets-list");
    const summaryCaptionEl = document.getElementById("summary-caption");
    const totalPriceEl = document.getElementById("total-price");
    const btnBook = document.getElementById("btn-book");
    const btnClear = document.getElementById("btn-clear");

    const ticketTypes = {
      adult:  { label: "Взрослый", price: 2500 },
      student:{ label: "Студент", price: 1900 },
      child:  { label: "Детский", price: 1600 }
    };

    const seatConfig = { rows: 8, cols: 12 };
    let movies = [];
    let selectedMovie = null;
    let selectedShow = null; // { hall, time }
    const selectedSeats = new Map(); // key -> { rowLabel, seatNumber, type }

    function rowLabel(index) {
      return String.fromCharCode("A".charCodeAt(0) + index);
    }

    function generateScheduleForMovie(movie, index) {
      // Просто детерминированный "рандом": 3 сеанса, залы 1–8
      const baseHour = 11 + (index % 3) * 3;
      const hall1 = (index % 8) + 1;
      const hall2 = ((index + 3) % 8) + 1;
      const hall3 = ((index + 5) % 8) + 1;
      const time1 = String(baseHour).padStart(2,"0") + ":00";
      const time2 = String(baseHour+3).padStart(2,"0") + ":30";
      const time3 = String(baseHour+6).padStart(2,"0") + ":15";
      return [
        { hall: hall1, time: time1 },
        { hall: hall2, time: time2 },
        { hall: hall3, time: time3 },
      ];
    }

    async function loadMovies() {
      moviesEl.innerHTML = "<p style='color:var(--muted);font-size:13px;'>Загружаем афишу...</p>";
      try {
        const res = await fetch("/api/movies");
        if (!res.ok) throw new Error("HTTP " + res.status);
        const data = await res.json();
        if (!Array.isArray(data) || data.length === 0) {
          moviesEl.innerHTML = "<p style='color:var(--muted);font-size:13px;'>Фильмы пока не добавлены.</p>";
          return;
        }
        movies = data;
        renderMovies();
      } catch (e) {
        moviesEl.innerHTML = "<p style='color:var(--muted);font-size:13px;'>Не удалось загрузить фильмы.</p>";
      }
    }

    function renderMovies() {
        moviesEl.innerHTML = "";
      movies.forEach((m, index) => {
        const card = document.createElement("div");
        card.className = "movie-card";
        card.dataset.id = m.id;

        const poster = document.createElement("div");
        poster.className = "movie-poster";
        const initial = (m.title || "?").trim().charAt(0).toUpperCase();
        poster.innerHTML = "<span>" + initial + "</span>";

        const info = document.createElement("div");
        info.className = "movie-info";
        const h3 = document.createElement("h3");
        h3.textContent = m.title || "(Без названия)";
        const meta = document.createElement("div");
        meta.className = "movie-meta";
        meta.textContent =
          (m.genre || "Жанр не указан") +
          " · " +
          (m.duration ? (m.duration + " мин") : "длительность неизвестна");

        const tags = document.createElement("div");
        tags.className = "movie-tags";

        const tagRating = document.createElement("span");
        tagRating.className = "tag tag-rating";
        tagRating.textContent = "Рейтинг: " + (m.rating ?? "-");
        tags.appendChild(tagRating);

        const tagId = document.createElement("span");
        tagId.className = "tag";
        tagId.textContent = "ID #" + m.id;
        tags.appendChild(tagId);

        info.appendChild(h3);
        if (m.description) {
          const p = document.createElement("p");
          p.style.margin = "2px 0 0";
          p.style.fontSize = "12px";
          p.style.color = "var(--muted)";
          p.textContent = m.description;
          info.appendChild(p);
        }
        info.appendChild(meta);
        info.appendChild(tags);

        card.appendChild(poster);
        card.appendChild(info);

        card.addEventListener("click", () => selectMovie(m, index, card));

        moviesEl.appendChild(card);
      });
    }

    function selectMovie(movie, index, cardEl) {
      selectedMovie = movie;
      selectedShow = null;
      selectedSeats.clear();
      updateSummary();
      hallLabelEl.textContent = "–";
      timeLabelEl.textContent = "Время не выбрано";
      scheduleRowEl.innerHTML = "";
      seatGridEl.innerHTML = "";

      document.querySelectorAll(".movie-card").forEach(c => c.classList.remove("active"));
      if (cardEl) cardEl.classList.add("active");

      bookingTitle.textContent = movie.title || "Выбранный фильм";
      bookingSubtitle.textContent =
        (movie.genre || "Жанр не указан") +
        " · " +
        (movie.duration ? (movie.duration + " мин") : "длительность неизвестна") +
        (movie.rating ? " · Рейтинг: " + movie.rating : "");

      const schedule = generateScheduleForMovie(movie, index);
      schedule.forEach((s, i) => {
        const chip = document.createElement("button");
        chip.type = "button";
        chip.className = "chip";
        chip.innerHTML = "<span>" + s.time + " · Зал " + s.hall + "</span>";
        chip.addEventListener("click", () => {
          document.querySelectorAll(".chip").forEach(c => c.classList.remove("active"));
          chip.classList.add("active");
          selectedShow = s;
          hallLabelEl.textContent = "Зал " + s.hall;
          timeLabelEl.textContent = "Сегодня, " + s.time;
          selectedSeats.clear();
          renderSeatGrid();
          updateSummary();
        });
        if (i === 0) {
          // Выбираем первый по умолчанию
          chip.click();
        }
        scheduleRowEl.appendChild(chip);
      });
    }

    function renderSeatGrid() {
      seatGridEl.innerHTML = "";
      if (!selectedShow) {
        seatGridEl.innerHTML = "<p style='color:var(--muted);font-size:12px;text-align:center;margin-top:12px;'>Сначала выберите время и зал.</p>";
                return;
              }
      const frag = document.createDocumentFragment();
      for (let r = 0; r < seatConfig.rows; r++) {
        const row = document.createElement("div");
        row.className = "seat-row";
        const label = document.createElement("div");
        label.className = "seat-row-label";
        const rLabel = rowLabel(r);
        label.textContent = rLabel;
        row.appendChild(label);

        for (let c = 1; c <= seatConfig.cols; c++) {
          const key = rLabel + c;
          const seat = document.createElement("button");
          seat.type = "button";
          seat.className = "seat";
          seat.dataset.key = key;
          seat.innerHTML = "<span></span>";
          seat.addEventListener("click", () => toggleSeat(key, rLabel, c, seat));
          row.appendChild(seat);
        }
        frag.appendChild(row);
      }
      seatGridEl.appendChild(frag);
    }

    function toggleSeat(key, rowLabelValue, seatNumber, seatEl) {
      if (!selectedShow) return;
      if (selectedSeats.has(key)) {
        selectedSeats.delete(key);
        seatEl.classList.remove("selected");
      } else {
        selectedSeats.set(key, {
          rowLabel: rowLabelValue,
          seatNumber: seatNumber,
          type: "adult"
        });
        seatEl.classList.add("selected");
      }
      updateSummary();
    }

    function updateSummary() {
      ticketsListEl.innerHTML = "";
      if (selectedSeats.size === 0 || !selectedShow || !selectedMovie) {
        summaryCaptionEl.textContent = "Выберите один или несколько мест на схеме.";
        totalPriceEl.textContent = "0 ₸";
        btnBook.disabled = true;
        return;
      }
      summaryCaptionEl.textContent =
        "Зал " + selectedShow.hall + ", " + selectedShow.time + " · " + selectedSeats.size + " мест(а)";

      let total = 0;
      selectedSeats.forEach((ticket, key) => {
        const container = document.createElement("div");
        container.className = "ticket-item";

        const place = document.createElement("div");
        place.innerHTML = "Ряд " + ticket.rowLabel + ", место " + ticket.seatNumber +
          "<small>Билет " + key + "</small>";

        const typeSel = document.createElement("select");
        typeSel.className = "ticket-type";
        for (const [code, t] of Object.entries(ticketTypes)) {
          const opt = document.createElement("option");
          opt.value = code;
          opt.textContent = t.label;
          if (code === ticket.type) opt.selected = true;
          typeSel.appendChild(opt);
        }
        typeSel.addEventListener("change", () => {
          ticket.type = typeSel.value;
          renderSummary(); // пересчёт цен
        });

        const price = document.createElement("div");
        price.className = "ticket-price";
        price.dataset.key = key;

        container.appendChild(place);
        container.appendChild(typeSel);
        container.appendChild(price);

        ticketsListEl.appendChild(container);
      });

      function renderSummary() {
        let sum = 0;
        selectedSeats.forEach((ticket, key) => {
          const type = ticketTypes[ticket.type] || ticketTypes.adult;
          const itemPrice = type.price;
          sum += itemPrice;
          const priceEl = ticketsListEl.querySelector('.ticket-price[data-key="' + key + '"]');
          if (priceEl) {
            priceEl.textContent = itemPrice.toLocaleString("ru-RU") + " ₸";
          }
        });
        total = sum;
        totalPriceEl.textContent = total.toLocaleString("ru-RU") + " ₸";
        btnBook.disabled = selectedSeats.size === 0;
      }
      renderSummary();
    }

    btnClear.addEventListener("click", () => {
      selectedSeats.clear();
      document.querySelectorAll(".seat.selected").forEach(s => s.classList.remove("selected"));
      updateSummary();
    });

    btnBook.addEventListener("click", () => {
      if (!selectedMovie || !selectedShow || selectedSeats.size === 0) return;
      alert("Демо: бронь создана.\n\nФильм: " + selectedMovie.title +
        "\nЗал: " + selectedShow.hall +
        "\nВремя: " + selectedShow.time +
        "\nМест: " + selectedSeats.size +
        "\nСумма: " + totalPriceEl.textContent +
        "\n\n(В учебной версии данные не сохраняются в базе.)");
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
	// Защита методов изменения фильмов: только роль admin может создавать/обновлять/удалять.
	protectedMovies := middleware.RequireRoleForMethods(
		movieHandler,
		jwtSecret,
		map[string][]string{
			http.MethodPost:   {"admin"},
			http.MethodPut:    {"admin"},
			http.MethodDelete: {"admin"},
		},
	)

	http.Handle("/api/movies", protectedMovies)
	http.Handle("/api/movies/", protectedMovies)

	// Аутентификация.
	http.HandleFunc("/api/auth/register", authHandler.Register)
	http.HandleFunc("/api/auth/login", authHandler.Login)

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
