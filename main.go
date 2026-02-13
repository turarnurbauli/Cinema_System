package main

import (
	"cinema-system/handler"
	"cinema-system/middleware"
	"cinema-system/model"
	"cinema-system/repository"
	"cinema-system/service"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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
	if err := userSvc.EnsureUserWithRole(ctx, "customer", "1234", "Guest", service.RoleCustomer); err != nil {
		log.Fatalf("failed to ensure default customer: %v", err)
	}

	authHandler := handler.NewAuthHandler(userSvc, jwtSecret)
	profileHandler := handler.NewProfileHandler(userSvc)

	repo, err := repository.NewMovieRepo(ctx, client, dbName)
	if err != nil {
		log.Fatalf("failed to initialise movie repository: %v", err)
	}

	hallRepo, err := repository.NewHallRepo(ctx, client, dbName)
	if err != nil {
		log.Fatalf("failed to initialise hall repository: %v", err)
	}
	seatRepo, err := repository.NewSeatRepo(ctx, client, dbName)
	if err != nil {
		log.Fatalf("failed to initialise seat repository: %v", err)
	}
	sessionRepo, err := repository.NewSessionRepo(ctx, client, dbName)
	if err != nil {
		log.Fatalf("failed to initialise session repository: %v", err)
	}
	bookingRepo, err := repository.NewBookingRepo(ctx, client, dbName)
	if err != nil {
		log.Fatalf("failed to initialise booking repository: %v", err)
	}
	// Ensure seats exist for all halls (idempotent). Fix VIP to last 2 rows only.
	halls, _ := hallRepo.GetAll()
	for _, h := range halls {
		_ = seatRepo.EnsureSeatsForHall(ctx, h)
		_ = seatRepo.FixVipSeatsForHall(ctx, h)
	}

	// Repository → Service → Handler (Assignment 3 architecture)
	svc := service.NewMovieService(repo)
	movieHandler := handler.NewMovieHandler(svc)

	sessionSvc := service.NewSessionService(sessionRepo, hallRepo, seatRepo, bookingRepo)
	sessionHandler := handler.NewSessionHandler(sessionSvc)
	bookingSvc := service.NewBookingService(bookingRepo, sessionRepo, seatRepo)
	bookingHandler := handler.NewBookingHandler(bookingSvc, userRepo)
	clientHandler := handler.NewClientHandler(userRepo, bookingRepo)

	// Simple HTML homepage so root "/" is not empty.
	http.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1, viewport-fit=cover" />
  <title>NotFlix – Cinema &amp; Tickets</title>
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
      padding: 20px 14px 28px;
      flex: 1;
      display: flex;
      flex-direction: column;
      gap: 20px;
    }

    header {
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 12px;
    }
    .brand {
      display: flex;
      align-items: center;
      gap: 10px;
    }
    .brand-logo {
      width: 36px;
      height: 36px;
      border-radius: 10px;
      background: conic-gradient(from 200deg, #f97316, #facc15, #22c55e, #0ea5e9, #a855f7, #f97316);
      display: flex;
      align-items: center;
      justify-content: center;
      box-shadow: 0 0 0 1px rgba(15, 23, 42, 0.8), 0 18px 45px rgba(15, 23, 42, 0.9);
    }
    .brand-logo span {
      font-weight: 800;
      font-size: 18px;
      color: #020617;
    }
    .brand-text h1 {
      margin: 0;
      font-size: 20px;
      font-weight: 700;
      letter-spacing: 0.03em;
    }
    .brand-text p {
      margin: 2px 0 0;
      font-size: 12px;
      color: var(--muted);
    }

    .header-actions {
      display: flex;
      align-items: center;
      gap: 6px;
    }
    .avatar-button {
      border: none;
      background: transparent;
      padding: 0;
      cursor: pointer;
    }
    .avatar-circle {
      width: 34px;
      height: 34px;
      border-radius: 999px;
      background: linear-gradient(135deg, #f97316, #facc15);
      display: flex;
      align-items: center;
      justify-content: center;
      color: #020617;
      font-weight: 700;
      font-size: 15px;
      overflow: hidden;
    }
    .avatar-circle img {
      width: 100%;
      height: 100%;
      object-fit: cover;
    }
    .user-menu-dropdown {
      position: absolute;
      top: 110%;
      right: 0;
      background: var(--bg-elevated);
      border-radius: 10px;
      border: 1px solid var(--border-subtle);
      min-width: 160px;
      padding: 4px 0;
      box-shadow: 0 14px 36px rgba(15, 23, 42, 0.9);
      z-index: 20;
    }
    .user-menu-item {
      padding: 6px 10px;
      font-size: 12px;
      cursor: pointer;
    }
    .user-menu-item:hover {
      background: rgba(148, 163, 184, 0.15);
    }
    .user-menu-separator {
      height: 1px;
      background: var(--border-subtle);
      margin: 4px 0;
    }
    .header-search {
      display: flex;
      align-items: center;
      gap: 5px;
      padding: 3px 8px;
      border-radius: 999px;
      background: rgba(15, 23, 42, 0.9);
      border: 1px solid rgba(148, 163, 184, 0.35);
      font-size: 11px;
      color: var(--muted);
    }
    .header-search input {
      border: none;
      outline: none;
      background: transparent;
      color: var(--text);
      font-size: 11px;
      width: 140px;
    }
    .header-search input::placeholder {
      color: var(--muted);
    }
    .header-search-icon {
      font-size: 13px;
    }
    .header-sort {
      display: flex;
      align-items: center;
      gap: 4px;
      padding: 4px 8px;
      border-radius: 999px;
      border: 1px solid rgba(148, 163, 184, 0.55);
      background: radial-gradient(circle at top, rgba(15,23,42,0.95), #020617);
      font-size: 11px;
      color: var(--muted);
    }
    .header-sort select {
      background: transparent;
      border: none;
      outline: none;
      color: var(--text);
      font-size: 11px;
      padding: 0 2px;
      cursor: pointer;
    }
    .header-sort select option {
      background: #020617;
      color: var(--text);
      font-size: 11px;
    }
    .pill {
      padding: 3px 8px;
      border-radius: 999px;
      border: 1px solid rgba(148, 163, 184, 0.4);
      font-size: 10px;
      text-transform: uppercase;
      letter-spacing: 0.06em;
      color: var(--muted);
      background: rgba(15, 23, 42, 0.9);
    }
    .pill strong {
      color: var(--accent-strong);
    }

    main {
      display: grid;
      grid-template-columns: 1fr;
      gap: 20px;
      align-items: flex-start;
    }

    .panel {
      border-radius: 16px;
      background: radial-gradient(circle at top left, rgba(15, 23, 42, 0.9), #020617 60%);
      border: 1px solid rgba(15, 23, 42, 0.9);
      box-shadow:
        0 20px 48px rgba(15, 23, 42, 0.95),
        inset 0 0 0 1px rgba(15, 23, 42, 0.85);
      padding: 14px 16px 16px;
    }

    .panel-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 6px;
    }
    .panel-header h2 {
      margin: 0;
      font-size: 15px;
    }
    .panel-header p {
      margin: 0;
      font-size: 11px;
      color: var(--muted);
    }

    /* Movies list: 3 columns, centered by default */
    .movies-panel {
      max-width: 1120px;
      margin-left: auto;
      margin-right: auto;
    }
    .app-layout.right-panel-open .movies-panel {
      max-width: none;
      margin-left: 0;
      margin-right: 0;
    }
    .booking-panel {
      max-width: 1120px;
      margin-left: auto;
      margin-right: auto;
    }
    .movies-list {
      display: grid;
      grid-template-columns: repeat(3, 1fr);
      gap: 14px;
      max-height: 580px;
      overflow: auto;
      padding-right: 4px;
    }
    .movie-card {
      display: grid;
      grid-template-columns: auto minmax(0, 1fr);
      gap: 14px;
      padding: 14px;
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
      width: 80px;
      height: 116px;
      border-radius: 12px;
      background: linear-gradient(135deg, #1d4ed8, #a855f7);
      display: flex;
      align-items: flex-end;
      justify-content: center;
      overflow: hidden;
      font-size: 28px;
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
    .movie-poster img {
      width: 100%;
      height: 100%;
      object-fit: cover;
      border-radius: 14px;
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
      margin: 0 0 2px;
      font-size: 15px;
    }
    .booking-header-main p {
      margin: 0;
      font-size: 12px;
      color: var(--muted);
    }

    .schedule-row {
      display: flex;
      flex-wrap: wrap;
      gap: 6px;
      margin-top: 4px;
    }
    .chip {
      padding: 5px 10px;
      border-radius: 999px;
      border: 1px solid rgba(55, 65, 81, 0.9);
      font-size: 11px;
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
      gap: 10px;
      padding: 6px 10px;
      border-radius: 10px;
      background: rgba(15, 23, 42, 0.7);
      border: 1px solid rgba(30, 64, 175, 0.4);
    }
    .hall-name {
      font-size: 12px;
      color: var(--muted);
    }
    .hall-name strong {
      color: var(--accent-strong);
    }

    /* Seat map */
    .seat-map-wrapper {
      border-radius: 14px;
      background: radial-gradient(circle at top, rgba(15, 23, 42, 0.9), rgba(2, 6, 23, 1));
      border: 1px solid rgba(15, 23, 42, 0.9);
      padding: 10px 12px 12px;
      display: flex;
      flex-direction: column;
      gap: 6px;
      min-height: 200px;
    }
    .screen-label {
      text-align: center;
      font-size: 10px;
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
      gap: 8px;
      font-size: 10px;
      color: var(--muted);
      margin-top: 2px;
    }
    .legend-item {
      display: inline-flex;
      align-items: center;
      gap: 3px;
    }
    .legend-swatch {
      width: 10px;
      height: 10px;
      border-radius: 2px;
      border: 1px solid rgba(55, 65, 81, 0.9);
    }
    .legend-swatch.free { background: var(--seat-free); }
    .legend-swatch.selected { background: var(--seat-selected); border-color: rgba(22, 163, 74, 0.9); }
    .legend-swatch.unavailable { background: var(--seat-unavailable); }
    .legend-swatch.seat-vip { border-color: rgba(250, 204, 21, 0.8); background: linear-gradient(135deg, rgba(250, 204, 21, 0.15), var(--seat-free)); }
    .seat.seat-vip { border-color: rgba(250, 204, 21, 0.8); background: linear-gradient(135deg, rgba(250, 204, 21, 0.15), var(--seat-free)); }
    .seat.seat-vip.selected { background: var(--seat-selected); border-color: rgba(34, 197, 94, 0.95); }

    #auth-modal-overlay {
      position: fixed; inset: 0; background: rgba(2, 6, 23, 0.95); z-index: 9999;
      display: flex; align-items: center; justify-content: center; padding: 20px;
    }
    #auth-modal-overlay.hidden { display: none; }
    #auth-modal-box {
      background: var(--bg-elevated); border: 1px solid var(--border-subtle);
      border-radius: 16px; padding: 22px; max-width: 340px; width: 100%;
      box-shadow: 0 20px 48px rgba(0,0,0,0.5);
    }
    #auth-modal-box h2 { margin: 0 0 16px; font-size: 18px; text-align: center; }
    #auth-modal-box .auth-tabs { display: flex; gap: 6px; margin-bottom: 16px; }
    #auth-modal-box .auth-tabs button {
      flex: 1; padding: 8px; border-radius: 8px; border: 1px solid var(--border-subtle);
      background: var(--bg); color: var(--muted); font-size: 13px; cursor: pointer;
    }
    #auth-modal-box .auth-tabs button.active { background: var(--accent-soft); color: var(--accent-strong); border-color: var(--accent); }
    #auth-modal-box .auth-form { display: none; }
    #auth-modal-box .auth-form.active { display: block; }
    #auth-modal-box .auth-form input {
      width: 100%; margin-bottom: 8px; padding: 8px 10px; border-radius: 8px;
      border: 1px solid var(--border-subtle); background: var(--bg); color: var(--text); box-sizing: border-box; font-size: 14px;
    }
    #auth-modal-box .auth-form button[type="button"] { width: 100%; margin-top: 6px; padding: 8px; font-size: 13px; }
    #app-content { min-height: 100vh; }

    /* Ticket summary */
    .summary {
      display: flex;
      flex-direction: column;
      gap: 4px;
      font-size: 12px;
    }
    .summary-row {
      display: flex;
      justify-content: space-between;
      gap: 6px;
    }
    .summary-row strong {
      font-size: 13px;
    }
    .summary-list {
      max-height: 120px;
      overflow: auto;
      padding-right: 4px;
    }
    .ticket-item {
      display: grid;
      grid-template-columns: minmax(0, 1fr) minmax(0, 1fr) auto;
      gap: 6px;
      align-items: center;
      padding: 6px 8px;
      border-radius: 8px;
      background: rgba(15, 23, 42, 0.85);
      border: 1px solid rgba(31, 41, 55, 0.9);
      margin-bottom: 4px;
    }
    .ticket-item .ticket-type-wrap { min-width: 0; }
    .ticket-item .ticket-type { min-width: 90px; }
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
      font-size: 12px;
      color: var(--accent-strong);
      text-align: right;
    }

    button {
      border-radius: 999px;
      border: none;
      padding: 7px 12px;
      font-size: 12px;
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

    .app-layout {
      display: flex;
      gap: 0;
      flex: 1;
      min-height: 0;
      align-items: stretch;
    }
    .main-content {
      flex: 1;
      min-width: 0;
      display: flex;
      flex-direction: column;
      gap: 20px;
    }
    #right-panel {
      width: 0;
      overflow: hidden;
      transition: width 0.25s ease;
      flex-shrink: 0;
      border-left: none;
      background: var(--bg-soft);
    }
    #right-panel.open {
      width: 360px;
      min-width: 360px;
      border-left: 1px solid var(--border-subtle);
    }
    .app-layout.right-panel-open .main-content {
      max-width: calc(100% - 360px);
    }
    .app-layout.right-panel-open .movies-list {
      grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
    }
    .app-layout.right-panel-booking.right-panel-open .main-content {
      max-width: calc(100% - 480px);
    }
    .app-layout.right-panel-booking.right-panel-open .movies-list {
      grid-template-columns: 1fr;
    }
    .app-layout.right-panel-booking #right-panel.open {
      width: 480px;
      min-width: 480px;
    }
    .app-layout.right-panel-booking .right-panel-inner {
      width: 480px;
    }
    .right-panel-inner {
      width: 360px;
      height: 100%;
      min-height: 360px;
      padding: 16px;
      overflow: auto;
      display: flex;
      flex-direction: column;
      gap: 14px;
    }
    .right-panel-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 6px;
    }
    .right-panel-header h3 { margin: 0; font-size: 15px; }
    .right-panel-content input, .right-panel-content select, .right-panel-content textarea {
      width: 100%;
      margin-bottom: 6px;
      padding: 7px 9px;
      border-radius: 8px;
      border: 1px solid var(--border-subtle);
      background: var(--bg);
      color: var(--text);
      box-sizing: border-box;
    }
    .right-panel-content textarea { resize: vertical; min-height: 52px; }
    .right-panel-content .row { margin-bottom: 8px; }
    .right-panel-content button { margin-right: 6px; margin-top: 2px; }
    .film-list-admin { list-style: none; padding: 0; margin: 0 0 12px; }
    .film-list-admin li {
      display: flex; align-items: center; justify-content: space-between; gap: 8px;
      padding: 8px 10px; border-radius: 10px; background: rgba(15,23,42,0.8);
      margin-bottom: 6px; border: 1px solid var(--border-subtle);
    }
    .film-list-admin li span { flex: 1; font-size: 13px; }
    .film-list-admin .btn-sm { padding: 4px 10px; font-size: 12px; }

    footer {
      max-width: 1280px;
      margin: 0 auto;
      padding: 0 14px 18px;
      font-size: 10px;
      color: var(--muted);
      display: flex;
      justify-content: space-between;
      gap: 6px;
    }

    /* Mobile-friendly */
    @media (max-width: 768px) {
      .shell { padding: 12px 12px 24px; gap: 16px; }
      header { flex-wrap: wrap; gap: 10px; }
      .brand-text h1 { font-size: 18px; }
      .brand-text p { font-size: 12px; }
      .header-actions { flex-wrap: wrap; gap: 6px; }
      .header-search { margin-right: 0; }
      .header-search input { width: 120px; }
      .pill { font-size: 10px; padding: 6px 8px; }
      .movies-list {
        grid-template-columns: 1fr;
        max-height: none;
        gap: 12px;
      }
      .movie-card {
        grid-template-columns: auto minmax(0, 1fr);
        padding: 12px;
        gap: 12px;
      }
      .movie-poster { width: 72px; height: 104px; font-size: 26px; }
      .movie-info h3 { font-size: 15px; }
      .panel-header { flex-wrap: wrap; gap: 8px; }
      .panel { padding: 12px 14px; }
      .booking-layout { padding: 12px; }
      .chip { min-height: 44px; padding: 10px 14px; font-size: 14px; }
      button:not(.btn-sm):not(.avatar-button), .btn-outline:not(.btn-sm) { min-height: 44px; padding: 10px 16px; }
      #right-panel.open { width: 100%; min-width: 100%; max-width: 100%; }
      .app-layout.right-panel-open .main-content { max-width: 0; overflow: hidden; }
      .app-layout.right-panel-booking #right-panel.open { width: 100%; min-width: 100%; }
      .app-layout.right-panel-booking .right-panel-inner { width: 100%; }
      .right-panel-inner { padding: 14px; min-height: 320px; }
      .seat-grid { gap: 6px; }
      .seat { min-width: 32px; min-height: 32px; font-size: 11px; }
      footer { flex-direction: column; text-align: center; padding: 12px 16px; }
    }
    @media (max-width: 480px) {
      .shell { padding: 12px 10px 24px; padding-left: max(10px, env(safe-area-inset-left)); padding-right: max(10px, env(safe-area-inset-right)); }
      .brand-logo { width: 32px; height: 32px; }
      .brand-logo span { font-size: 16px; }
      .header-search input { width: 100px; }
      .avatar-circle { width: 32px; height: 32px; font-size: 14px; }
      .user-menu-dropdown { min-width: 160px; }
      .movie-card { grid-template-columns: 1fr; }
      .movie-poster { width: 100%; height: 160px; font-size: 48px; }
      #auth-modal-box { max-width: 94vw; max-height: 90vh; overflow: auto; }
    }
  </style>
</head>
<body>
  <div id="auth-modal-overlay">
    <div id="auth-modal-box">
      <h2>NotFlix</h2>
      <p style="text-align:center;color:var(--muted);font-size:13px;margin:0 0 16px;">Log in or sign up</p>
      <div class="auth-tabs">
        <button type="button" id="modal-tab-login" class="active">Log in</button>
        <button type="button" id="modal-tab-register">Sign up</button>
      </div>
      <div id="modal-login-form" class="auth-form active">
        <small style="display:block;margin-bottom:8px;color:var(--muted);font-size:11px;">Test: customer / 1234</small>
        <input type="text" id="modal-login-email" placeholder="Email" autocomplete="username" />
        <input type="password" id="modal-login-password" placeholder="Password" autocomplete="current-password" />
        <button type="button" id="modal-btn-login">Log in</button>
      </div>
      <div id="modal-register-form" class="auth-form">
        <input type="text" id="modal-register-name" placeholder="Your name *" />
        <input type="text" id="modal-register-email" placeholder="Email *" autocomplete="email" />
        <input type="password" id="modal-register-password" placeholder="Password *" autocomplete="new-password" />
        <button type="button" id="modal-btn-register">Sign up</button>
      </div>
    </div>
  </div>
  <div id="app-content" style="display:none;">
  <div class="shell">
    <header>
      <div class="brand">
        <div class="brand-logo"><span>NF</span></div>
        <div class="brand-text">
          <h1>NotFlix</h1>
          <p>Choose film, hall and seats.</p>
        </div>
      </div>
      <div class="header-actions">
        <div class="header-search" style="margin-right:8px;">
          <span class="header-search-icon">🔍</span>
          <input type="text" id="movie-search" placeholder="Search films..." />
        </div>
        <div id="role-pill" style="display:none;" class="pill"></div>
        <div id="user-menu" style="position:relative;display:none;margin-right:8px;">
          <button type="button" id="user-avatar-btn" class="avatar-button">
            <div id="user-avatar" class="avatar-circle">U</div>
          </button>
          <div id="user-dropdown" class="user-menu-dropdown" style="display:none;">
            <div class="user-menu-item" id="menu-my-profile">My profile</div>
            <div class="user-menu-item" id="menu-my-purchases">My purchases</div>
            <div class="user-menu-separator"></div>
            <div class="user-menu-item" id="menu-logout">Log out</div>
          </div>
        </div>
        <div id="auth-bar">
          <button type="button" id="btn-register" class="btn-outline" style="margin-right:8px;">Sign up</button>
          <button type="button" id="btn-login" class="btn-outline" style="margin-right:8px;">Log in</button>
          <div id="register-form" style="display:none;position:absolute;top:100%;right:0;margin-top:8px;padding:14px;background:var(--bg-elevated);border-radius:12px;border:1px solid var(--border-subtle);z-index:11;min-width:220px;">
            <div style="font-size:13px;font-weight:600;margin-bottom:10px;">New customer sign up</div>
            <input type="text" id="register-name" placeholder="Your name *" style="display:block;margin-bottom:6px;padding:8px 10px;border-radius:8px;border:1px solid var(--border-subtle);background:var(--bg);color:var(--text);width:100%;box-sizing:border-box;" />
            <input type="text" id="register-email" placeholder="Email *" autocomplete="email" style="display:block;margin-bottom:6px;padding:8px 10px;border-radius:8px;border:1px solid var(--border-subtle);background:var(--bg);color:var(--text);width:100%;box-sizing:border-box;" />
            <input type="password" id="register-password" placeholder="Password *" autocomplete="new-password" style="display:block;margin-bottom:10px;padding:8px 10px;border-radius:8px;border:1px solid var(--border-subtle);background:var(--bg);color:var(--text);width:100%;box-sizing:border-box;" />
            <button type="button" id="btn-do-register">Sign up</button>
          </div>
          <div id="login-form" style="display:none;position:absolute;top:100%;right:0;margin-top:8px;padding:12px;background:var(--bg-elevated);border-radius:12px;border:1px solid var(--border-subtle);z-index:10;">
            <small style="display:block;margin-bottom:6px;color:var(--muted);font-size:11px;">Test: customer / 1234</small>
            <input type="text" id="login-email" placeholder="Email" autocomplete="username" style="display:block;margin-bottom:6px;padding:6px 10px;border-radius:8px;border:1px solid var(--border-subtle);background:var(--bg);color:var(--text);width:200px;" />
            <input type="password" id="login-password" placeholder="Password" autocomplete="current-password" style="display:block;margin-bottom:8px;padding:6px 10px;border-radius:8px;border:1px solid var(--border-subtle);background:var(--bg);color:var(--text);width:200px;" />
            <button type="button" id="btn-do-login">Log in</button>
          </div>
        </div>
        <div class="pill">Go · MongoDB</div>
      </div>
    </header>

    <div class="app-layout">
    <div class="main-content">
    <main>
      <section class="panel movies-panel">
        <div class="panel-header" style="display:flex;align-items:flex-start;justify-content:space-between;gap:12px;">
          <div>
            <h2>Movies</h2>
            <p>Click a film to choose a session.</p>
          </div>
          <div style="display:flex;align-items:center;gap:8px;">
            <div class="header-sort">
              <span>Sort</span>
              <select id="movie-sort">
                <option value="default">Default</option>
                <option value="title">Title A–Z</option>
                <option value="rating">Rating ↓</option>
                <option value="duration">Duration ↓</option>
              </select>
            </div>
            <button type="button" id="btn-add-movie-in-movies" class="btn-outline" style="display:none;flex-shrink:0;">Add film</button>
          </div>
        </div>
        <div id="movies" class="movies-list"></div>
    </section>

      <section id="booking-section" class="panel booking-panel">
        <div class="booking-layout">
          <div class="booking-header">
            <div class="booking-header-main">
              <h2 id="booking-title">Choose a film</h2>
              <p id="booking-subtitle">Sessions and times will appear after you select a film.</p>
            </div>
          </div>

          <div>
            <div class="hall-row">
              <div class="hall-name">
                Hall: <strong id="hall-label">–</strong>
                <small id="time-label" style="display:block;margin-top:2px;color:var(--muted);font-size:11px;">Time not selected</small>
              </div>
              <div class="schedule-row" id="schedule-row"></div>
            </div>
          </div>

          <div class="seat-map-wrapper">
            <div>
              <div class="screen-label">Screen</div>
              <div class="screen-bar"></div>
            </div>
            <div id="seat-grid" class="seat-grid"></div>
            <div class="seat-legend">
              <div class="legend-item">
                <div class="legend-swatch free"></div><span>Available</span>
              </div>
              <div class="legend-item">
                <div class="legend-swatch selected"></div><span>Selected</span>
              </div>
              <div class="legend-item">
                <div class="legend-swatch unavailable"></div><span>Taken</span>
              </div>
              <div class="legend-item">
                <div class="legend-swatch seat-vip"></div><span>VIP (2× adult)</span>
              </div>
            </div>
          </div>

          <div class="summary">
            <div class="summary-row">
              <div>
                <strong>Your tickets</strong>
                <div id="summary-caption" class="summary-caption" style="font-size:11px;color:var(--muted);margin-top:2px;">
                  Click seats on the map to select.
                </div>
              </div>
              <div>
                <div class="ticket-price" id="total-price">0 ₸</div>
              </div>
            </div>
            <div id="tickets-list" class="summary-list"></div>
            <div class="summary-row">
              <button id="btn-book" disabled>Book tickets</button>
              <button id="btn-clear" class="btn-outline" type="button">Clear selection</button>
            </div>
          </div>
        </div>
      </section>
      <div id="booking-section-anchor"></div>
    </main>

  <section id="admin-panel" class="panel" style="display:none;margin-top:16px;">
    <div class="panel-header"><h2>Admin panel</h2></div>
    <div style="margin-top:0;">
      <h3 style="font-size:14px;margin:0 0 8px;">Sessions</h3>
      <button type="button" id="btn-add-session-sidebar" class="btn-outline" style="margin-bottom:10px;">Add session</button>
      <p style="font-size:12px;color:var(--muted);margin:0 0 8px;">Edit by ID: <button type="button" id="btn-edit-session-by-id" class="btn-outline btn-sm">Edit session</button></p>
    </div>
    <div style="margin-top:12px;">
      <button type="button" id="btn-load-all-bookings" class="btn-outline">Show all bookings</button>
      <div id="all-bookings-list" style="margin-top:8px;max-height:200px;overflow:auto;"></div>
    </div>
    <div style="margin-top:16px;">
      <h3 style="font-size:14px;margin:0 0 8px;">Client data</h3>
      <input type="number" id="admin-client-id" placeholder="User ID" style="width:120px;margin-right:8px;padding:8px;border-radius:8px;border:1px solid var(--border-subtle);background:var(--bg);color:var(--text);" />
      <button type="button" id="btn-load-client-admin" class="btn-outline">Load</button>
      <div id="admin-client-data" style="margin-top:8px;padding:10px;border-radius:12px;background:var(--bg-soft);border:1px solid var(--border-subtle);min-height:40px;font-size:13px;"></div>
    </div>
  </section>
  <section id="cashier-panel" class="panel" style="display:none;margin-top:16px;">
    <div class="panel-header"><h2>Cashier panel</h2></div>
    <p style="font-size:13px;color:var(--muted);margin:0 0 12px;">Bookings, cancel/refund, change seats, free seats by session.</p>
    <div>
      <h3 style="font-size:14px;margin:0 0 8px;">All bookings</h3>
      <button type="button" id="btn-cashier-bookings" class="btn-outline">Refresh list</button>
      <div id="cashier-bookings-list" style="margin-top:8px;max-height:220px;overflow:auto;"></div>
    </div>
    <div style="margin-top:16px;">
      <h3 style="font-size:14px;margin:0 0 8px;">Free seats by session</h3>
      <input type="number" id="cashier-session-id" placeholder="Session ID" style="width:100px;margin-right:8px;padding:8px;border-radius:8px;border:1px solid var(--border-subtle);background:var(--bg);color:var(--text);" />
      <button type="button" id="btn-cashier-free-seats" class="btn-outline">Show seats</button>
      <div id="cashier-free-seats-result" style="margin-top:8px;padding:10px;border-radius:12px;background:var(--bg-soft);border:1px solid var(--border-subtle);min-height:40px;font-size:13px;"></div>
    </div>
    <div style="margin-top:16px;">
      <h3 style="font-size:14px;margin:0 0 8px;">Client data</h3>
      <input type="number" id="cashier-client-id" placeholder="User ID" style="width:120px;margin-right:8px;padding:8px;border-radius:8px;border:1px solid var(--border-subtle);background:var(--bg);color:var(--text);" />
      <button type="button" id="btn-load-client-cashier" class="btn-outline">Load</button>
      <div id="cashier-client-data" style="margin-top:8px;padding:10px;border-radius:12px;background:var(--bg-soft);border:1px solid var(--border-subtle);min-height:40px;font-size:13px;"></div>
    </div>
  </section>
  </div>

  <aside id="right-panel">
    <div class="right-panel-inner">
      <div class="right-panel-header">
        <h3 id="right-panel-title">Edit</h3>
        <button type="button" id="right-panel-close" class="btn-outline btn-sm">Close</button>
      </div>
      <div id="right-panel-content"></div>
    </div>
  </aside>
  </div>
  </div>

  <footer>
    <span>NotFlix · Course project.</span>
    <span>Team: Nurbauli Turar &amp; Alkhan Almas</span>
  </footer>
  </div>

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
      adult:  { label: "Adult", price: 2500 },
      student:{ label: "Student", price: 1900 },
      child:  { label: "Child", price: 1600 }
    };
    const vipSeatPrice = 5000;

    let movies = [];
    let halls = [];
    let selectedMovie = null;
    let selectedSession = null; // { id, movieId, hallId, startTime, price }
    let seatsData = []; // { id, rowNumber, seatNumber, seatType, booked }
    const selectedSeats = new Map(); // seatId -> { rowNumber, seatNumber, seatType }

    function rowLabel(index) {
      return String.fromCharCode("A".charCodeAt(0) + index);
    }

    async function loadHalls() {
      try {
        const res = await fetch("/api/halls");
        if (res.ok) { halls = await res.json(); if (!halls) halls = []; }
      } catch (e) { halls = []; }
    }
    function hallName(hallId) {
      const h = halls.find(x => x.id === hallId);
      return h ? h.name : "Hall " + hallId;
    }

    async function loadMovies() {
      moviesEl.innerHTML = "<p style='color:var(--muted);font-size:13px;'>Loading movies...</p>";
      try {
        const res = await fetch("/api/movies");
        if (!res.ok) throw new Error("HTTP " + res.status);
        const data = await res.json();
        if (!Array.isArray(data)) {
          movies = [];
        } else {
          movies = data;
        }
        applyMovieSearchFilter();
      } catch (e) {
        moviesEl.innerHTML = "<p style='color:var(--muted);font-size:13px;'>Failed to load movies.</p>";
      }
    }

    function renderMovies(list) {
      const arr = Array.isArray(list) ? list : movies;
      moviesEl.innerHTML = "";
      if (!arr.length) {
        moviesEl.innerHTML = "<p style='color:var(--muted);font-size:13px;'>No movies found.</p>";
        return;
      }
      arr.forEach((m, index) => {
        const card = document.createElement("div");
        card.className = "movie-card";
        card.dataset.id = m.id;

        const poster = document.createElement("div");
        poster.className = "movie-poster";
        const initial = (m.title || "?").trim().charAt(0).toUpperCase();
        if (m.posterUrl && m.posterUrl.trim()) {
          var img = document.createElement("img");
          img.src = m.posterUrl.trim();
          img.alt = m.title || "";
          img.onerror = function() { img.style.display = "none"; poster.innerHTML = "<span>" + initial + "</span>"; };
          poster.appendChild(img);
        } else {
          poster.innerHTML = "<span>" + initial + "</span>";
        }

        const info = document.createElement("div");
        info.className = "movie-info";
        const h3 = document.createElement("h3");
        h3.textContent = m.title || "(Untitled)";
        const meta = document.createElement("div");
        meta.className = "movie-meta";
        meta.textContent =
          (m.genre || "Genre not specified") +
          " · " +
          (m.duration ? (m.duration + " min") : "duration unknown");

        const tags = document.createElement("div");
        tags.className = "movie-tags";

        const tagRating = document.createElement("span");
        tagRating.className = "tag tag-rating";
        tagRating.textContent = "Rating: " + (m.rating ?? "-");
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

    async function selectMovie(movie, index, cardEl) {
      selectedMovie = movie;
      selectedSession = null;
      selectedSeats.clear();
      seatsData = [];
      updateSummary();
      hallLabelEl.textContent = "–";
      timeLabelEl.textContent = "Time not selected";
      scheduleRowEl.innerHTML = "<span style=\"color:var(--muted);font-size:13px;\">Loading sessions...</span>";
      seatGridEl.innerHTML = "";

      document.querySelectorAll(".movie-card").forEach(c => c.classList.remove("active"));
      if (cardEl) cardEl.classList.add("active");

      var role = localStorage.getItem("cinema_role");
      if (role === "admin") {
        openEditFilmPanel(movie);
      } else if (role === "customer" || role === "cashier") {
        openBookingRightPanel(movie);
      }

      bookingTitle.textContent = movie.title || "Selected film";
      bookingSubtitle.textContent =
        (movie.genre || "Genre not specified") +
        " · " +
        (movie.duration ? (movie.duration + " min") : "duration unknown") +
        (movie.rating ? " · Rating: " + movie.rating : "");

      try {
        const res = await fetch("/api/sessions?movieId=" + movie.id);
        if (!res.ok) throw new Error("HTTP " + res.status);
        const sessions = await res.json();
        scheduleRowEl.innerHTML = "";
        if (!sessions || sessions.length === 0) {
          scheduleRowEl.innerHTML = "<span style=\"color:var(--muted);font-size:13px;\">No sessions.</span>";
          return;
        }
        sessions.forEach((sess, i) => {
          const t = (sess.startTime || "").replace("Z","").slice(11, 16) || "—";
          const chip = document.createElement("button");
          chip.type = "button";
          chip.className = "chip";
          chip.dataset.sessionId = sess.id;
          chip.innerHTML = "<span>" + t + " · " + hallName(sess.hallId) + "</span>";
          chip.addEventListener("click", () => selectSession(sess, chip));
          scheduleRowEl.appendChild(chip);
          if (i === 0) selectSession(sess, chip);
        });
      } catch (e) {
        scheduleRowEl.innerHTML = "<span style=\"color:var(--muted);font-size:13px;\">Failed to load sessions.</span>";
      }
    }

    async function selectSession(sess, chipEl) {
      document.querySelectorAll(".chip").forEach(c => c.classList.remove("active"));
      if (chipEl) chipEl.classList.add("active");
      selectedSession = sess;
      selectedSeats.clear();
      hallLabelEl.textContent = hallName(sess.hallId);
      timeLabelEl.textContent = (sess.startTime || "").slice(0, 16).replace("T", " ");
      seatGridEl.innerHTML = "<span style=\"color:var(--muted);font-size:13px;\">Loading seats...</span>";
      try {
        const res = await fetch("/api/sessions/" + sess.id + "/seats");
        if (!res.ok) throw new Error("HTTP " + res.status);
        seatsData = await res.json();
        if (!seatsData) seatsData = [];
        renderSeatGrid();
        updateSummary();
      } catch (e) {
        seatGridEl.innerHTML = "<span style=\"color:var(--muted);font-size:13px;\">Failed to load seats.</span>";
      }
    }

    function renderSeatGrid() {
      seatGridEl.innerHTML = "";
      if (!selectedSession || !seatsData.length) {
        seatGridEl.innerHTML = "<p style='color:var(--muted);font-size:12px;text-align:center;margin-top:12px;'>Select a session first.</p>";
        return;
      }
      const byRow = {};
      seatsData.forEach(seat => {
        const r = seat.rowNumber || 0;
        if (!byRow[r]) byRow[r] = [];
        byRow[r].push(seat);
      });
      const rowNums = Object.keys(byRow).map(Number).sort((a,b) => a - b);
      const frag = document.createDocumentFragment();
      rowNums.forEach(rNum => {
        const row = document.createElement("div");
        row.className = "seat-row";
        const label = document.createElement("div");
        label.className = "seat-row-label";
        label.textContent = rNum;
        row.appendChild(label);
        byRow[rNum].sort((a,b) => (a.seatNumber || 0) - (b.seatNumber || 0)).forEach(s => {
          const seatEl = document.createElement("button");
          seatEl.type = "button";
          seatEl.className = "seat" + (s.seatType === "vip" ? " seat-vip" : "");
          seatEl.dataset.seatId = s.id;
          seatEl.title = s.seatType === "vip" ? "VIP" : "";
          seatEl.innerHTML = "<span></span>";
          if (s.booked) { seatEl.classList.add("unavailable"); seatEl.disabled = true; }
          else seatEl.addEventListener("click", () => toggleSeat(s, seatEl));
          row.appendChild(seatEl);
        });
        frag.appendChild(row);
      });
      seatGridEl.appendChild(frag);
    }

    function toggleSeat(s, seatEl) {
      if (!selectedSession) return;
      if (selectedSeats.has(s.id)) {
        selectedSeats.delete(s.id);
        seatEl.classList.remove("selected");
      } else {
        selectedSeats.set(s.id, { rowNumber: s.rowNumber, seatNumber: s.seatNumber, seatType: s.seatType || "regular", ticketType: "adult" });
        seatEl.classList.add("selected");
      }
      updateSummary();
    }

    function setSeatTicketType(seatId, ticketType) {
      var info = selectedSeats.get(seatId);
      if (info) { info.ticketType = ticketType; updateSummary(); }
    }
    function updateSummary() {
      ticketsListEl.innerHTML = "";
      if (selectedSeats.size === 0 || !selectedSession || !selectedMovie) {
        summaryCaptionEl.textContent = "Select one or more seats on the map.";
        totalPriceEl.textContent = "0";
        btnBook.disabled = true;
        return;
      }
      summaryCaptionEl.textContent =
        hallName(selectedSession.hallId) + ", " + (selectedSession.startTime || "").slice(11, 16) + " · " + selectedSeats.size + " seat(s)";

      var total = 0;
      selectedSeats.forEach((info, seatId) => {
        var isVip = info.seatType === "vip";
        var price = isVip ? vipSeatPrice : (ticketTypes[info.ticketType || "adult"] ? ticketTypes[info.ticketType || "adult"].price : ticketTypes.adult.price);
        total += price;
        var container = document.createElement("div");
        container.className = "ticket-item";
        var typeBlock = isVip ? "<span class=\"ticket-type-label\">VIP (" + vipSeatPrice.toLocaleString("en-US") + ")</span>" : "<select class=\"ticket-type\" data-seat-id=\"" + seatId + "\"><option value=\"adult\"" + ((info.ticketType || "adult") === "adult" ? " selected" : "") + ">Adult (" + ticketTypes.adult.price + ")</option><option value=\"student\"" + ((info.ticketType || "adult") === "student" ? " selected" : "") + ">Student (" + ticketTypes.student.price + ")</option><option value=\"child\"" + ((info.ticketType || "adult") === "child" ? " selected" : "") + ">Child (" + ticketTypes.child.price + ")</option></select>";
        container.innerHTML = "<div>Row " + info.rowNumber + ", seat " + info.seatNumber + "<small>ID " + seatId + "</small></div><div class=\"ticket-type-wrap\">" + typeBlock + "</div><div class=\"ticket-price\">" + price.toLocaleString("en-US") + "</div>";
        ticketsListEl.appendChild(container);
      });
      ticketsListEl.querySelectorAll(".ticket-type").forEach(function(sel) {
        sel.addEventListener("change", function() { setSeatTicketType(parseInt(sel.dataset.seatId, 10), sel.value); });
      });
      totalPriceEl.textContent = total.toLocaleString("en-US");
      btnBook.disabled = false;
    }

    btnClear.addEventListener("click", () => {
      selectedSeats.clear();
      document.querySelectorAll(".seat.selected").forEach(s => s.classList.remove("selected"));
      updateSummary();
    });

    btnBook.addEventListener("click", async () => {
      if (!selectedMovie || !selectedSession || selectedSeats.size === 0) return;
      const token = localStorage.getItem("cinema_token");
      if (!token) {
        alert("Please log in to book (use Log in in the header).");
        return;
      }
      const seatIds = Array.from(selectedSeats.keys());
      const ticketTypes = seatIds.map(function(sid) { var info = selectedSeats.get(sid); return (info && info.ticketType) ? info.ticketType : "adult"; });
      try {
        const res = await fetch("/api/bookings", {
          method: "POST",
          headers: { "Content-Type": "application/json", "Authorization": "Bearer " + token },
          body: JSON.stringify({ sessionId: selectedSession.id, seatIds: seatIds, ticketTypes: ticketTypes })
        });
        const data = await res.json().catch(() => ({}));
        if (!res.ok) {
          alert("Error: " + (data.error || res.status));
          return;
        }
        alert("Booking created. Total: " + (data.totalPrice || 0).toLocaleString("en-US"));
        selectedSeats.clear();
        document.querySelectorAll(".seat.selected").forEach(s => s.classList.remove("selected"));
        selectSession(selectedSession, document.querySelector(".chip.active"));
      } catch (e) {
        alert("Network error.");
      }
    });

    function updateRoleUI(role) {
      const pill = document.getElementById("role-pill");
      const adminPanel = document.getElementById("admin-panel");
      const cashierPanel = document.getElementById("cashier-panel");
      const bookingSection = document.getElementById("booking-section");
      const btnRegister = document.getElementById("btn-register");
      const btnAddMovie = document.getElementById("btn-add-movie-in-movies");
      if (role) {
        pill.style.display = "inline-block";
        pill.textContent = role === "admin" ? "ADMIN" : role === "cashier" ? "CASHIER" : "GUEST";
        adminPanel.style.display = role === "admin" ? "block" : "none";
        cashierPanel.style.display = role === "cashier" ? "block" : "none";
        if (bookingSection) bookingSection.style.display = (role === "admin" || role === "customer" || role === "cashier") ? "none" : "block";
        if (btnRegister) btnRegister.style.display = "none";
        if (btnAddMovie) btnAddMovie.style.display = role === "admin" ? "inline-block" : "none";
        updateUserAvatarUI();
      } else {
        pill.style.display = "none";
        adminPanel.style.display = "none";
        cashierPanel.style.display = "none";
        if (bookingSection) bookingSection.style.display = "block";
        if (btnRegister) btnRegister.style.display = "inline-block";
        if (btnAddMovie) btnAddMovie.style.display = "none";
        updateUserAvatarUI();
      }
    }
    function showApp() {
      document.getElementById("auth-modal-overlay").classList.add("hidden");
      document.getElementById("app-content").style.display = "block";
    }
    function showAuthModal() {
      document.getElementById("auth-modal-overlay").classList.remove("hidden");
      document.getElementById("app-content").style.display = "none";
    }
    (function initAuth() {
      const token = localStorage.getItem("cinema_token");
      const role = localStorage.getItem("cinema_role");
      if (token) {
        showApp();
        document.getElementById("btn-login").textContent = "Log out";
        updateRoleUI(role || "customer");
      } else {
        showAuthModal();
      }
      updateUserAvatarUI();
    })();

    // Movie search in header
    const movieSearchInput = document.getElementById("movie-search");
    const movieSortSelect = document.getElementById("movie-sort");
    function applyMovieSearchFilter() {
      if (!Array.isArray(movies)) {
        movies = [];
      }
      let list = movies.slice();
      if (movieSearchInput) {
        const q = movieSearchInput.value.trim().toLowerCase();
        if (q) {
          list = list.filter(function(m) {
            const title = (m.title || "").toLowerCase();
            const desc = (m.description || "").toLowerCase();
            const genre = (m.genre || "").toLowerCase();
            return title.includes(q) || desc.includes(q) || genre.includes(q);
          });
        }
      }
      if (movieSortSelect) {
        const sort = movieSortSelect.value;
        if (sort === "title") {
          list.sort(function(a, b) {
            return (a.title || "").localeCompare(b.title || "");
          });
        } else if (sort === "rating") {
          list.sort(function(a, b) {
            return (b.rating || 0) - (a.rating || 0);
          });
        } else if (sort === "duration") {
          list.sort(function(a, b) {
            return (b.duration || 0) - (a.duration || 0);
          });
        }
      }
      renderMovies(list);
    }
    if (movieSearchInput) {
      movieSearchInput.addEventListener("input", function() {
        applyMovieSearchFilter();
      });
    }
    if (movieSortSelect) {
      movieSortSelect.addEventListener("change", function() {
        applyMovieSearchFilter();
      });
    }

    document.getElementById("modal-tab-login").addEventListener("click", function() {
      document.getElementById("modal-tab-login").classList.add("active");
      document.getElementById("modal-tab-register").classList.remove("active");
      document.getElementById("modal-login-form").classList.add("active");
      document.getElementById("modal-register-form").classList.remove("active");
    });
    document.getElementById("modal-tab-register").addEventListener("click", function() {
      document.getElementById("modal-tab-register").classList.add("active");
      document.getElementById("modal-tab-login").classList.remove("active");
      document.getElementById("modal-register-form").classList.add("active");
      document.getElementById("modal-login-form").classList.remove("active");
    });

    document.getElementById("modal-btn-login").addEventListener("click", async () => {
      const email = document.getElementById("modal-login-email").value.trim();
      const password = document.getElementById("modal-login-password").value;
      if (!email || !password) { alert("Enter email and password."); return; }
      try {
        const res = await fetch("/api/auth/login", { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ email: email, password: password }) });
        const text = await res.text();
        let data = {};
        try { data = JSON.parse(text); } catch (_) {}
        if (!res.ok) { alert(data.error || "Login error"); return; }
        if (!data.token) { alert("Server did not return a token."); return; }
        localStorage.setItem("cinema_token", data.token);
        const role = (data.user && data.user.role) ? data.user.role : "customer";
        localStorage.setItem("cinema_role", role);
        saveUserFromAuthResponse(data);
        showApp();
        document.getElementById("btn-login").textContent = "Log out";
        updateRoleUI(role);
        document.getElementById("modal-login-email").value = "";
        document.getElementById("modal-login-password").value = "";
        loadHalls();
        loadMovies();
      } catch (e) { alert("Network error: " + e.message); }
    });
    document.getElementById("modal-btn-register").addEventListener("click", async () => {
      const name = document.getElementById("modal-register-name").value.trim();
      const email = document.getElementById("modal-register-email").value.trim();
      const password = document.getElementById("modal-register-password").value;
      if (!name) { alert("Enter your name."); return; }
      if (!email) { alert("Enter email."); return; }
      if (!password) { alert("Enter password."); return; }
      try {
        const res = await fetch("/api/auth/register", { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ name: name, email: email, password: password }) });
        const text = await res.text();
        let data = {};
        try { data = JSON.parse(text); } catch (_) {}
        if (!res.ok) { alert(data.error || "Registration error"); return; }
        if (data.token) {
          localStorage.setItem("cinema_token", data.token);
          const role = (data.user && data.user.role) ? data.user.role : "customer";
          localStorage.setItem("cinema_role", role);
          saveUserFromAuthResponse(data);
          showApp();
          document.getElementById("btn-login").textContent = "Log out";
          updateRoleUI(role);
          document.getElementById("modal-register-name").value = "";
          document.getElementById("modal-register-email").value = "";
          document.getElementById("modal-register-password").value = "";
          loadHalls();
          loadMovies();
        }
        alert("Registration successful. Welcome!");
      } catch (e) { alert("Network error: " + e.message); }
    });
    document.getElementById("btn-register").addEventListener("click", () => {
      document.getElementById("login-form").style.display = "none";
      const form = document.getElementById("register-form");
      form.style.display = form.style.display === "none" ? "block" : "none";
    });
    document.getElementById("btn-login").addEventListener("click", () => {
      if (localStorage.getItem("cinema_token")) {
        localStorage.removeItem("cinema_token");
        localStorage.removeItem("cinema_role");
        document.getElementById("btn-login").textContent = "Log in";
        updateRoleUI(null);
        showAuthModal();
        return;
      }
      document.getElementById("register-form").style.display = "none";
      const form = document.getElementById("login-form");
      form.style.display = form.style.display === "none" ? "block" : "none";
    });
    document.getElementById("btn-do-login").addEventListener("click", async () => {
      const email = document.getElementById("login-email").value.trim();
      const password = document.getElementById("login-password").value;
      if (!email || !password) { alert("Enter email and password."); return; }
      const btn = document.getElementById("btn-do-login");
      btn.disabled = true;
      try {
        const res = await fetch("/api/auth/login", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ email: email, password: password })
        });
        const text = await res.text();
        let data = {};
        try { data = JSON.parse(text); } catch (_) {}
        if (!res.ok) {
          alert(data.error || text || "Login error (code " + res.status + ")");
          btn.disabled = false;
          return;
        }
        if (!data.token) {
          alert("Server did not return a token.");
          btn.disabled = false;
          return;
        }
        localStorage.setItem("cinema_token", data.token);
        const role = (data.user && data.user.role) ? data.user.role : "customer";
        localStorage.setItem("cinema_role", role);
        saveUserFromAuthResponse(data);
        document.getElementById("login-form").style.display = "none";
        document.getElementById("btn-login").textContent = "Log out";
        document.getElementById("login-email").value = "";
        document.getElementById("login-password").value = "";
        updateRoleUI(role);
      } catch (e) {
        alert("Network error: " + e.message);
      }
      btn.disabled = false;
    });

    document.getElementById("btn-do-register").addEventListener("click", async () => {
      const name = document.getElementById("register-name").value.trim();
      const email = document.getElementById("register-email").value.trim();
      const password = document.getElementById("register-password").value;
      if (!name) { alert("Enter your name."); return; }
      if (!email) { alert("Enter email."); return; }
      if (!password) { alert("Enter password."); return; }
      const btn = document.getElementById("btn-do-register");
      btn.disabled = true;
      try {
        const res = await fetch("/api/auth/register", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ name: name, email: email, password: password })
        });
        const text = await res.text();
        let data = {};
        try { data = JSON.parse(text); } catch (_) {}
        if (!res.ok) {
          alert(data.error || text || "Registration error");
          btn.disabled = false;
          return;
        }
        if (data.token) {
          localStorage.setItem("cinema_token", data.token);
          const role = (data.user && data.user.role) ? data.user.role : "customer";
          localStorage.setItem("cinema_role", role);
          saveUserFromAuthResponse(data);
          document.getElementById("register-form").style.display = "none";
          document.getElementById("btn-login").textContent = "Log out";
          updateRoleUI(role);
          document.getElementById("register-name").value = "";
          document.getElementById("register-email").value = "";
          document.getElementById("register-password").value = "";
        }
        alert("Registration successful. You are logged in.");
      } catch (e) { alert("Network error: " + e.message); }
      btn.disabled = false;
    });

    function authHeaders() {
      const t = localStorage.getItem("cinema_token");
      return t ? { "Content-Type": "application/json", "Authorization": "Bearer " + t } : { "Content-Type": "application/json" };
    }
    function authHeadersForUpload() {
      const t = localStorage.getItem("cinema_token");
      return t ? { "Authorization": "Bearer " + t } : {};
    }
    function saveUserFromAuthResponse(data) {
      if (data && data.user) {
        if (data.user.name) localStorage.setItem("cinema_name", data.user.name);
        if (data.user.email) localStorage.setItem("cinema_email", data.user.email);
        if (data.user.avatarUrl) localStorage.setItem("cinema_avatar", data.user.avatarUrl);
      }
    }
    function updateUserAvatarUI() {
      const menu = document.getElementById("user-menu");
      const authBar = document.getElementById("auth-bar");
      const token = localStorage.getItem("cinema_token");
      const name = localStorage.getItem("cinema_name") || "";
      const avatarUrl = localStorage.getItem("cinema_avatar") || "";
      const avatar = document.getElementById("user-avatar");
      if (!menu || !avatar) return;
      if (token) {
        if (authBar) authBar.style.display = "none";
        menu.style.display = "inline-block";
        const initial = (name || "U").trim().charAt(0).toUpperCase();
        avatar.innerHTML = initial;
        avatar.style.backgroundImage = "";
        if (avatarUrl) {
          avatar.innerHTML = "";
          const img = document.createElement("img");
          img.src = avatarUrl;
          img.alt = name || "User";
          img.onerror = function() {
            avatar.innerHTML = initial;
            img.remove();
          };
          avatar.appendChild(img);
        }
      } else {
        if (authBar) authBar.style.display = "flex";
        menu.style.display = "none";
      }
    }
    const rightPanel = document.getElementById("right-panel");
    const rightPanelTitle = document.getElementById("right-panel-title");
    const rightPanelContent = document.getElementById("right-panel-content");
    var rightPanelShowingBooking = false;
    function ensureBookingSectionInMain() {
      var bookingSection = document.getElementById("booking-section");
      var anchor = document.getElementById("booking-section-anchor");
      var main = document.querySelector("main");
      if (bookingSection && anchor && main && bookingSection.parentNode === rightPanelContent) {
        main.insertBefore(bookingSection, anchor);
        rightPanelShowingBooking = false;
        var role = localStorage.getItem("cinema_role");
        if (role === "customer" || role === "cashier") bookingSection.style.display = "none";
        else bookingSection.style.display = "block";
      }
    }
    function openRightPanel(title, html) {
      ensureBookingSectionInMain();
      rightPanelShowingBooking = false;
      rightPanelTitle.textContent = title;
      rightPanelContent.innerHTML = html;
      rightPanel.classList.add("open");
      document.querySelector(".app-layout").classList.add("right-panel-open");
    }
    function openBookingRightPanel(movie) {
      var bookingSection = document.getElementById("booking-section");
      var anchor = document.getElementById("booking-section-anchor");
      if (!bookingSection || !anchor) return;
      rightPanelTitle.textContent = movie.title ? (movie.title + " – Choose session and seats") : "Choose session and seats";
      rightPanelContent.innerHTML = "";
      rightPanelContent.appendChild(bookingSection);
      bookingSection.style.display = "block";
      rightPanelShowingBooking = true;
      rightPanel.classList.add("open");
      var layout = document.querySelector(".app-layout");
      layout.classList.add("right-panel-open", "right-panel-booking");
    }
    function closeRightPanel() {
      if (rightPanelShowingBooking) {
        var bookingSection = document.getElementById("booking-section");
        var anchor = document.getElementById("booking-section-anchor");
        var main = document.querySelector("main");
        if (bookingSection && anchor && main) {
          main.insertBefore(bookingSection, anchor);
          var role = localStorage.getItem("cinema_role");
          if (role === "customer" || role === "cashier") bookingSection.style.display = "none";
        }
        rightPanelShowingBooking = false;
      } else {
        ensureBookingSectionInMain();
        rightPanelContent.innerHTML = "";
      }
      rightPanel.classList.remove("open");
      var layout = document.querySelector(".app-layout");
      layout.classList.remove("right-panel-open", "right-panel-booking");
    }
    document.getElementById("right-panel-close").addEventListener("click", closeRightPanel);

    async function openMyPurchasesPanel() {
      ensureBookingSectionInMain();
      rightPanelShowingBooking = false;
      rightPanelTitle.textContent = "My purchases";
      rightPanelContent.innerHTML = "<p style=\"color:var(--muted);font-size:13px;\">Loading...</p>";
      rightPanel.classList.add("open");
      var layout = document.querySelector(".app-layout");
      layout.classList.add("right-panel-open");
      try {
        var res = await fetch("/api/bookings", { headers: authHeaders() });
        var list = await res.json().catch(function() { return []; });
        if (!res.ok) {
          rightPanelContent.innerHTML = "<p style=\"color:var(--danger);\">Could not load bookings.</p>";
          return;
        }
        if (!list || list.length === 0) {
          rightPanelContent.innerHTML = "<p style=\"color:var(--muted);font-size:13px;\">No bookings yet.</p>";
          return;
        }
        var html = "<div class=\"right-panel-content\"><p style=\"font-size:12px;color:var(--muted);margin-bottom:10px;\">Your booking history</p>";
        list.forEach(function(b) {
          var date = (b.createdAt || "").slice(0, 16).replace("T", " ");
          html += "<div class=\"ticket-item\" style=\"margin-bottom:8px;\"><div><strong>#" + b.id + "</strong> · Session " + b.sessionId + " · " + (b.status || "—") + "</div><div style=\"font-size:12px;color:var(--muted);margin-top:2px;\">" + date + "</div><div class=\"ticket-price\" style=\"margin-top:4px;\">" + (b.totalPrice != null ? b.totalPrice.toLocaleString("en-US") : "—") + "</div></div>";
        });
        html += "</div>";
        rightPanelContent.innerHTML = html;
      } catch (e) {
        rightPanelContent.innerHTML = "<p style=\"color:var(--danger);\">Network error.</p>";
      }
    }
    async function openMyProfilePanel() {
      ensureBookingSectionInMain();
      rightPanelShowingBooking = false;
      rightPanelTitle.textContent = "My profile";
      rightPanelContent.innerHTML = "<p style=\"color:var(--muted);font-size:13px;\">Loading profile...</p>";
      rightPanel.classList.add("open");
      var layout = document.querySelector(".app-layout");
      layout.classList.add("right-panel-open");

      try {
        const res = await fetch("/api/me", { headers: authHeaders() });
        const u = await res.json().catch(function() { return {}; });
        if (!res.ok) {
          rightPanelContent.innerHTML = "<p style=\"color:var(--danger);font-size:13px;\">"
            + (u.error || "Could not load profile.") + "</p>";
          return;
        }
        const name = u.name || "";
        const email = u.email || "";
        const avatarUrl = u.avatarUrl || "";
        const role = u.role || "";

        rightPanelContent.innerHTML = "";
        var container = document.createElement("div");
        container.className = "right-panel-content";

        var header = document.createElement("div");
        header.style.marginBottom = "12px";
        header.style.display = "flex";
        header.style.alignItems = "center";
        header.style.gap = "12px";

        var avatar = document.createElement("div");
        avatar.className = "avatar-circle";
        avatar.style.width = "48px";
        avatar.style.height = "48px";
        avatar.style.fontSize = "18px";
        var initial = (name || "U").trim().charAt(0).toUpperCase();
        if (avatarUrl) {
          var img = document.createElement("img");
          img.src = avatarUrl;
          img.alt = name || "User";
          img.onerror = function() { avatar.textContent = initial; img.remove(); };
          avatar.appendChild(img);
        } else {
          avatar.textContent = initial;
        }

        var meta = document.createElement("div");
        meta.innerHTML = "<div style=\"font-size:13px;font-weight:600;\">" + (name || "User") + "</div>"
          + "<div style=\"font-size:11px;color:var(--muted);\">Role: <strong>" + (role || "—") + "</strong></div>";

        header.appendChild(avatar);
        header.appendChild(meta);
        container.appendChild(header);

        function addRow(labelText, inputEl) {
          var row = document.createElement("div");
          row.className = "row";
          var label = document.createElement("label");
          label.textContent = labelText;
          row.appendChild(label);
          row.appendChild(inputEl);
          container.appendChild(row);
        }

        var nameInput = document.createElement("input");
        nameInput.type = "text";
        nameInput.id = "rp-prof-name";
        nameInput.value = name;
        addRow("Name", nameInput);

        var emailInput = document.createElement("input");
        emailInput.type = "email";
        emailInput.id = "rp-prof-email";
        emailInput.value = email;
        addRow("Email (login)", emailInput);

        var avatarInput = document.createElement("input");
        avatarInput.type = "url";
        avatarInput.id = "rp-prof-avatar";
        avatarInput.placeholder = "https://... or /posters/avatar.jpg";
        avatarInput.value = avatarUrl;
        addRow("Avatar URL", avatarInput);

        var fileRow = document.createElement("div");
        fileRow.className = "row";
        var fileLabel = document.createElement("label");
        fileLabel.textContent = "Or upload avatar";
        var fileInput = document.createElement("input");
        fileInput.type = "file";
        fileInput.id = "rp-prof-avatar-file";
        fileInput.accept = "image/*";
        var uploadBtn = document.createElement("button");
        uploadBtn.type = "button";
        uploadBtn.id = "rp-prof-upload-avatar";
        uploadBtn.className = "btn-outline";
        uploadBtn.textContent = "Upload";
        fileRow.appendChild(fileLabel);
        fileRow.appendChild(fileInput);
        fileRow.appendChild(uploadBtn);
        container.appendChild(fileRow);

        var currentPwInput = document.createElement("input");
        currentPwInput.type = "password";
        currentPwInput.id = "rp-prof-current-pw";
        currentPwInput.placeholder = "Leave empty to keep password";
        addRow("Current password", currentPwInput);

        var newPwInput = document.createElement("input");
        newPwInput.type = "password";
        newPwInput.id = "rp-prof-new-pw";
        newPwInput.placeholder = "New password";
        addRow("New password", newPwInput);

        var saveBtn = document.createElement("button");
        saveBtn.type = "button";
        saveBtn.id = "rp-prof-save";
        saveBtn.textContent = "Save changes";
        container.appendChild(saveBtn);

        rightPanelContent.appendChild(container);

        uploadBtn.addEventListener("click", async function() {
          if (!fileInput || !fileInput.files || !fileInput.files[0]) {
            alert("Choose an image file first.");
            return;
          }
          var fd = new FormData();
          fd.append("poster", fileInput.files[0]);
          try {
            var res2 = await fetch("/api/upload-poster", { method: "POST", headers: authHeadersForUpload(), body: fd });
            var d2 = await res2.json().catch(function() { return {}; });
            if (!res2.ok) {
              alert(d2.error || "Upload failed.");
              return;
            }
            if (avatarInput && d2.url) {
              avatarInput.value = d2.url;
              alert("Avatar uploaded. Click Save changes to apply.");
            }
          } catch (e) {
            alert("Error: " + e.message);
          }
        });

        saveBtn.addEventListener("click", async function() {
          var body = {
            name: nameInput.value.trim(),
            email: emailInput.value.trim(),
            avatarUrl: avatarInput.value.trim(),
            currentPassword: currentPwInput.value,
            newPassword: newPwInput.value
          };
          try {
            var res3 = await fetch("/api/me", { method: "PUT", headers: authHeaders(), body: JSON.stringify(body) });
            var d3 = await res3.json().catch(function() { return {}; });
            if (!res3.ok) {
              alert(d3.error || "Could not save profile.");
              return;
            }
            saveUserFromAuthResponse({ user: d3 });
            updateUserAvatarUI();
            alert("Profile updated.");
            openMyProfilePanel();
          } catch (e) {
            alert("Error: " + e.message);
          }
        });
      } catch (e) {
        rightPanelContent.innerHTML = "<p style=\"color:var(--danger);font-size:13px;\">Error: " + e.message + "</p>";
      }
    }

    const userAvatarBtn = document.getElementById("user-avatar-btn");
    const userDropdown = document.getElementById("user-dropdown");
    if (userAvatarBtn && userDropdown) {
      userAvatarBtn.addEventListener("click", function(e) {
        e.stopPropagation();
        userDropdown.style.display = (userDropdown.style.display === "none" || !userDropdown.style.display) ? "block" : "none";
      });
      document.addEventListener("click", function() {
        userDropdown.style.display = "none";
      });
      userDropdown.addEventListener("click", function(e) { e.stopPropagation(); });
      document.getElementById("menu-my-profile").addEventListener("click", function() {
        userDropdown.style.display = "none";
        openMyProfilePanel();
      });
      document.getElementById("menu-my-purchases").addEventListener("click", function() {
        userDropdown.style.display = "none";
        openMyPurchasesPanel();
      });
      document.getElementById("menu-logout").addEventListener("click", function() {
        userDropdown.style.display = "none";
        localStorage.removeItem("cinema_token");
        localStorage.removeItem("cinema_role");
        localStorage.removeItem("cinema_name");
        localStorage.removeItem("cinema_email");
        localStorage.removeItem("cinema_avatar");
        document.getElementById("btn-login").textContent = "Log in";
        updateRoleUI(null);
        updateUserAvatarUI();
        showAuthModal();
      });
    }

    function openEditFilmPanel(m) {
      if (!m) return;
      var descEdit = (m.description || "").replace(/"/g, "&quot;").replace(/</g, "&lt;").replace(/>/g, "&gt;");
      var posterEdit = (m.posterUrl || "").replace(/"/g, "&quot;").replace(/</g, "&lt;").replace(/>/g, "&gt;");
      openRightPanel("Edit film", "<div class=\"right-panel-content\"><input type=\"hidden\" id=\"rp-edit-movie-id\" value=\"" + m.id + "\" /><div class=\"row\"><label>Title</label><input type=\"text\" id=\"rp-edit-movie-title\" value=\"" + (m.title || "").replace(/"/g, "&quot;") + "\" placeholder=\"Title\" /></div><div class=\"row\"><label>Genre</label><input type=\"text\" id=\"rp-edit-movie-genre\" value=\"" + (m.genre || "").replace(/"/g, "&quot;") + "\" placeholder=\"Genre\" /></div><div class=\"row\"><label>Description</label><textarea id=\"rp-edit-movie-description\" rows=\"3\" placeholder=\"Short description\">" + descEdit + "</textarea></div><div class=\"row\"><label>Duration (min)</label><input type=\"number\" id=\"rp-edit-movie-duration\" value=\"" + (m.duration || "") + "\" /></div><div class=\"row\"><label>Rating</label><input type=\"number\" id=\"rp-edit-movie-rating\" step=\"0.1\" value=\"" + (m.rating || "") + "\" /></div><div class=\"row\"><label>Poster URL</label><input type=\"url\" id=\"rp-edit-movie-posterUrl\" value=\"" + posterEdit + "\" placeholder=\"https://... or /posters/name.jpg\" /></div><div class=\"row\"><label>Or upload file</label><input type=\"file\" id=\"rp-edit-movie-file\" accept=\"image/*\" /><button type=\"button\" id=\"rp-edit-upload-poster\" class=\"btn-outline\">Upload</button></div><button type=\"button\" id=\"rp-btn-save-movie\">Save film</button> <button type=\"button\" id=\"rp-btn-delete-movie\" class=\"btn-outline\" style=\"margin-left:8px;\">Delete film</button></div>");
      rightPanelContent.querySelector("#rp-btn-save-movie").addEventListener("click", async () => {
        const id = rightPanelContent.querySelector("#rp-edit-movie-id").value;
        const title = rightPanelContent.querySelector("#rp-edit-movie-title").value.trim();
        const genre = rightPanelContent.querySelector("#rp-edit-movie-genre").value.trim();
        const description = (rightPanelContent.querySelector("#rp-edit-movie-description") && rightPanelContent.querySelector("#rp-edit-movie-description").value) ? rightPanelContent.querySelector("#rp-edit-movie-description").value.trim() : "";
        const duration = parseInt(rightPanelContent.querySelector("#rp-edit-movie-duration").value, 10) || 0;
        const rating = parseFloat(rightPanelContent.querySelector("#rp-edit-movie-rating").value) || 0;
        const posterUrl = (rightPanelContent.querySelector("#rp-edit-movie-posterUrl") && rightPanelContent.querySelector("#rp-edit-movie-posterUrl").value) ? rightPanelContent.querySelector("#rp-edit-movie-posterUrl").value.trim() : "";
        try {
          const res = await fetch("/api/movies/" + id, { method: "PUT", headers: authHeaders(), body: JSON.stringify({ title: title, genre: genre, duration: duration, rating: rating, description: description, posterUrl: posterUrl }) });
          const data = await res.json().catch(() => ({}));
          if (!res.ok) { alert(data.error || "Error"); return; }
          alert("Film saved.");
          closeRightPanel();
          loadMovies();
        } catch (e) { alert("Error: " + e.message); }
      });
      rightPanelContent.querySelector("#rp-btn-delete-movie").addEventListener("click", async () => {
        const id = rightPanelContent.querySelector("#rp-edit-movie-id").value;
        if (!id || !confirm("Delete this film?")) return;
        try {
          const res = await fetch("/api/movies/" + id, { method: "DELETE", headers: authHeaders() });
          if (!res.ok) { var d = await res.json().catch(() => ({})); alert(d.error || "Error"); return; }
          alert("Film deleted.");
          closeRightPanel();
          loadMovies();
        } catch (e) { alert("Error: " + e.message); }
      });
      var editFileInput = rightPanelContent.querySelector("#rp-edit-movie-file");
      var editPosterUrlInput = rightPanelContent.querySelector("#rp-edit-movie-posterUrl");
      rightPanelContent.querySelector("#rp-edit-upload-poster").addEventListener("click", async () => {
        if (!editFileInput || !editFileInput.files || !editFileInput.files[0]) { alert("Choose an image file first."); return; }
        var fd = new FormData();
        fd.append("poster", editFileInput.files[0]);
        try {
          var res = await fetch("/api/upload-poster", { method: "POST", headers: authHeadersForUpload(), body: fd });
          var data = await res.json().catch(function() { return {}; });
          if (!res.ok) { alert(data.error || "Upload failed."); return; }
          if (editPosterUrlInput && data.url) { editPosterUrlInput.value = data.url; alert("Poster uploaded. Click Save film to apply."); }
        } catch (e) { alert("Error: " + e.message); }
      });
    }
    document.getElementById("btn-add-movie-in-movies").addEventListener("click", () => {
      openRightPanel("Add film", "<div class=\"right-panel-content\"><div class=\"row\"><label>Title</label><input type=\"text\" id=\"rp-new-movie-title\" placeholder=\"Title\" /></div><div class=\"row\"><label>Genre</label><input type=\"text\" id=\"rp-new-movie-genre\" placeholder=\"Genre\" /></div><div class=\"row\"><label>Description</label><textarea id=\"rp-new-movie-description\" rows=\"3\" placeholder=\"Short description\"></textarea></div><div class=\"row\"><label>Duration (min)</label><input type=\"number\" id=\"rp-new-movie-duration\" placeholder=\"0\" /></div><div class=\"row\"><label>Rating</label><input type=\"number\" id=\"rp-new-movie-rating\" step=\"0.1\" placeholder=\"0\" /></div><div class=\"row\"><label>Poster URL</label><input type=\"url\" id=\"rp-new-movie-posterUrl\" placeholder=\"https://... or /posters/name.jpg\" /></div><div class=\"row\"><label>Or upload file</label><input type=\"file\" id=\"rp-new-movie-file\" accept=\"image/*\" /><button type=\"button\" id=\"rp-new-upload-poster\" class=\"btn-outline\">Upload</button></div><button type=\"button\" id=\"rp-btn-add-movie\">Add film</button></div>");
      var newFileInput = rightPanelContent.querySelector("#rp-new-movie-file");
      var newPosterUrlInput = rightPanelContent.querySelector("#rp-new-movie-posterUrl");
      rightPanelContent.querySelector("#rp-new-upload-poster").addEventListener("click", async () => {
        if (!newFileInput || !newFileInput.files || !newFileInput.files[0]) { alert("Choose an image file first."); return; }
        var fd = new FormData();
        fd.append("poster", newFileInput.files[0]);
        try {
          var res = await fetch("/api/upload-poster", { method: "POST", headers: authHeadersForUpload(), body: fd });
          var data = await res.json().catch(function() { return {}; });
          if (!res.ok) { alert(data.error || "Upload failed."); return; }
          if (newPosterUrlInput && data.url) { newPosterUrlInput.value = data.url; alert("Poster uploaded."); }
        } catch (e) { alert("Error: " + e.message); }
      });
      rightPanelContent.querySelector("#rp-btn-add-movie").addEventListener("click", async () => {
        const title = rightPanelContent.querySelector("#rp-new-movie-title").value.trim();
        const genre = rightPanelContent.querySelector("#rp-new-movie-genre").value.trim();
        const description = (rightPanelContent.querySelector("#rp-new-movie-description") && rightPanelContent.querySelector("#rp-new-movie-description").value) ? rightPanelContent.querySelector("#rp-new-movie-description").value.trim() : "";
        const duration = parseInt(rightPanelContent.querySelector("#rp-new-movie-duration").value, 10) || 0;
        const rating = parseFloat(rightPanelContent.querySelector("#rp-new-movie-rating").value) || 0;
        const posterUrl = (rightPanelContent.querySelector("#rp-new-movie-posterUrl") && rightPanelContent.querySelector("#rp-new-movie-posterUrl").value) ? rightPanelContent.querySelector("#rp-new-movie-posterUrl").value.trim() : "";
        if (!title) { alert("Enter film title."); return; }
        try {
          const res = await fetch("/api/movies", { method: "POST", headers: authHeaders(), body: JSON.stringify({ title: title, genre: genre, duration: duration, rating: rating, description: description, posterUrl: posterUrl }) });
          const data = await res.json().catch(() => ({}));
          if (!res.ok) { alert(data.error || "Error"); return; }
          alert("Film added.");
          closeRightPanel();
          loadMovies();
        } catch (e) { alert("Error: " + e.message); }
      });
    });
    document.getElementById("btn-add-session-sidebar").addEventListener("click", () => {
      const t = new Date().toISOString().slice(0,11) + "14:00:00Z";
      openRightPanel("Add session", "<div class=\"right-panel-content\"><div class=\"row\"><label>Film ID</label><input type=\"number\" id=\"rp-new-session-movieId\" placeholder=\"Film ID\" /></div><div class=\"row\"><label>Hall ID (1-4)</label><input type=\"number\" id=\"rp-new-session-hallId\" placeholder=\"1\" /></div><div class=\"row\"><label>Time (ISO)</label><input type=\"text\" id=\"rp-new-session-time\" placeholder=\"" + t + "\" value=\"" + t + "\" /></div><div class=\"row\"><label>Price</label><input type=\"number\" id=\"rp-new-session-price\" placeholder=\"2500\" value=\"2500\" /></div><button type=\"button\" id=\"rp-btn-add-session\">Add session</button></div>");
      rightPanelContent.querySelector("#rp-btn-add-session").addEventListener("click", async () => {
        const movieId = parseInt(rightPanelContent.querySelector("#rp-new-session-movieId").value, 10) || 0;
        const hallId = parseInt(rightPanelContent.querySelector("#rp-new-session-hallId").value, 10) || 0;
        const startTime = rightPanelContent.querySelector("#rp-new-session-time").value.trim() || t;
        const price = parseFloat(rightPanelContent.querySelector("#rp-new-session-price").value) || 2500;
        if (!movieId || !hallId) { alert("Enter film ID and hall ID."); return; }
        try {
          const res = await fetch("/api/sessions", { method: "POST", headers: authHeaders(), body: JSON.stringify({ movieId: movieId, hallId: hallId, startTime: startTime, price: price }) });
          const data = await res.json().catch(() => ({}));
          if (!res.ok) { alert(data.error || "Error"); return; }
          alert("Session added.");
          closeRightPanel();
          loadMovies();
        } catch (e) { alert("Error: " + e.message); }
      });
    });
    document.getElementById("btn-edit-session-by-id").addEventListener("click", () => {
      openRightPanel("Edit session", "<div class=\"right-panel-content\"><div class=\"row\"><label>Session ID</label><input type=\"number\" id=\"rp-edit-session-id\" placeholder=\"Session ID\" /></div><button type=\"button\" id=\"rp-btn-load-session\" class=\"btn-outline\">Load</button><div id=\"rp-session-fields\" style=\"display:none;\"><div class=\"row\"><label>Film ID</label><input type=\"number\" id=\"rp-edit-session-movieId\" /></div><div class=\"row\"><label>Hall ID</label><input type=\"number\" id=\"rp-edit-session-hallId\" /></div><div class=\"row\"><label>Time</label><input type=\"text\" id=\"rp-edit-session-time\" /></div><div class=\"row\"><label>Price</label><input type=\"number\" id=\"rp-edit-session-price\" /></div><button type=\"button\" id=\"rp-btn-save-session\">Save session</button></div></div>");
      rightPanelContent.querySelector("#rp-btn-load-session").addEventListener("click", async () => {
        const id = rightPanelContent.querySelector("#rp-edit-session-id").value.trim();
        if (!id) { alert("Enter session ID."); return; }
        try {
          const res = await fetch("/api/sessions/" + id);
          if (!res.ok) { alert("Session not found."); return; }
          const s = await res.json();
          rightPanelContent.querySelector("#rp-edit-session-movieId").value = s.movieId || "";
          rightPanelContent.querySelector("#rp-edit-session-hallId").value = s.hallId || "";
          rightPanelContent.querySelector("#rp-edit-session-time").value = s.startTime || "";
          rightPanelContent.querySelector("#rp-edit-session-price").value = s.price || "";
          rightPanelContent.querySelector("#rp-session-fields").style.display = "block";
        } catch (e) { alert("Error: " + e.message); }
      });
      rightPanelContent.querySelector("#rp-btn-save-session").addEventListener("click", async () => {
        const id = rightPanelContent.querySelector("#rp-edit-session-id").value.trim();
        if (!id) { alert("Enter session ID."); return; }
        const movieId = parseInt(rightPanelContent.querySelector("#rp-edit-session-movieId").value, 10) || 0;
        const hallId = parseInt(rightPanelContent.querySelector("#rp-edit-session-hallId").value, 10) || 0;
        const startTime = rightPanelContent.querySelector("#rp-edit-session-time").value.trim();
        const price = parseFloat(rightPanelContent.querySelector("#rp-edit-session-price").value) || 0;
        try {
          const res = await fetch("/api/sessions/" + id, { method: "PUT", headers: authHeaders(), body: JSON.stringify({ movieId: movieId, hallId: hallId, startTime: startTime, price: price }) });
          const data = await res.json().catch(() => ({}));
          if (!res.ok) { alert(data.error || "Error"); return; }
          alert("Session saved.");
          closeRightPanel();
        } catch (e) { alert("Error: " + e.message); }
      });
    });
    async function loadAllBookings(containerId) {
      const el = document.getElementById(containerId);
      el.innerHTML = "<p style=\"color:var(--muted);\">Loading...</p>";
      try {
        const res = await fetch("/api/bookings?all=1", { headers: authHeaders() });
        const list = await res.json().catch(() => []);
        if (!res.ok) { el.innerHTML = "<p style=\"color:var(--danger);\">Access error.</p>"; return; }
        var activeList = list.filter(function(b) { return b.status !== "cancelled"; });
        if (!activeList.length) { el.innerHTML = "<p style=\"color:var(--muted);\">No active bookings.</p>"; return; }
        const isCashier = containerId === "cashier-bookings-list";
        el.innerHTML = activeList.map(b => {
          const line = "#" + b.id + " · " + (b.userName || "Client #" + b.userId) + " · session " + b.sessionId + " · " + b.status + " · " + (b.totalPrice || 0).toLocaleString("en-US");
          if (isCashier && b.status !== "cancelled")
            return "<div class=\"ticket-item\" style=\"display:flex;align-items:center;justify-content:space-between;gap:8px;flex-wrap:wrap;\"><span style=\"flex:1;min-width:0;\">" + line + "</span><span style=\"flex-shrink:0;display:flex;gap:6px;\"><button type=\"button\" class=\"cashier-change-seats-btn btn-outline\" data-booking-id=\"" + b.id + "\" data-session-id=\"" + b.sessionId + "\" style=\"padding:4px 10px;font-size:12px;\">Change seats</button><button type=\"button\" class=\"cashier-cancel-btn btn-outline\" data-booking-id=\"" + b.id + "\" style=\"padding:4px 10px;font-size:12px;\">Cancel</button></span></div>";
          return "<div class=\"ticket-item\"><span>" + line + "</span></div>";
        }).join("");
        if (isCashier) {
          el.querySelectorAll(".cashier-cancel-btn").forEach(btn => {
            btn.addEventListener("click", async () => {
              const id = btn.dataset.bookingId;
              if (!id || !confirm("Cancel booking #" + id + "?")) return;
              try {
                const r = await fetch("/api/bookings/" + id, { method: "DELETE", headers: authHeaders() });
                if (r.ok) loadAllBookings("cashier-bookings-list");
                else alert("Cancel error.");
              } catch (e) { alert("Network error."); }
            });
          });
          el.querySelectorAll(".cashier-change-seats-btn").forEach(btn => {
            btn.addEventListener("click", () => openChangeSeatsPanel(btn.dataset.bookingId, btn.dataset.sessionId));
          });
        }
      } catch (e) { el.innerHTML = "<p style=\"color:var(--danger);\">Network error.</p>"; }
    }
    document.getElementById("btn-load-all-bookings").addEventListener("click", () => loadAllBookings("all-bookings-list"));
    document.getElementById("btn-cashier-bookings").addEventListener("click", () => loadAllBookings("cashier-bookings-list"));

    document.getElementById("btn-cashier-free-seats").addEventListener("click", async () => {
      const el = document.getElementById("cashier-free-seats-result");
      const id = document.getElementById("cashier-session-id").value.trim();
      if (!id) { el.innerHTML = "<span style=\"color:var(--muted);\">Enter session ID.</span>"; return; }
      el.innerHTML = "<span style=\"color:var(--muted);\">Loading...</span>";
      try {
        const res = await fetch("/api/sessions/" + id + "/seats");
        const seats = await res.json().catch(() => []);
        if (!res.ok) { el.innerHTML = "<span style=\"color:var(--danger);\">Session not found.</span>"; return; }
        const free = seats.filter(s => !s.booked);
        const booked = seats.filter(s => s.booked);
        el.innerHTML = "<div><strong>Free:</strong> " + free.length + " seats</div><div style=\"margin-top:4px;\"><strong>Booked:</strong> " + booked.length + "</div>" +
          (free.length ? "<div style=\"margin-top:6px;font-size:12px;color:var(--muted);\">Free: row–seat " + free.slice(0, 15).map(s => s.rowNumber + "–" + s.seatNumber).join(", ") + (free.length > 15 ? " …" : "") + "</div>" : "");
      } catch (e) { el.innerHTML = "<span style=\"color:var(--danger);\">Network error.</span>"; }
    });

    async function loadClientData(containerId, userIdInputId) {
      const el = document.getElementById(containerId);
      const id = document.getElementById(userIdInputId).value.trim();
      if (!id) { el.innerHTML = "<span style=\"color:var(--muted);\">Enter user ID.</span>"; return; }
      el.innerHTML = "<span style=\"color:var(--muted);\">Loading...</span>";
      try {
        const res = await fetch("/api/users/" + id, { headers: authHeaders() });
        const data = await res.json().catch(() => ({}));
        if (!res.ok) {
          el.innerHTML = "<span style=\"color:var(--danger);\">" + (data.error || "Error") + "</span>";
          return;
        }
        const u = data.user || {};
        const bookings = data.lastBookings || [];
        let html = "<div><strong>" + (u.name || "—") + "</strong> · " + (u.email || "—") + " · role: " + (u.role || "—") + "</div>";
        html += "<div style=\"margin-top:8px;font-size:12px;color:var(--muted);\">Recent bookings (" + bookings.length + "):</div>";
        if (bookings.length) {
          html += "<div style=\"margin-top:4px;\">" + bookings.map(b => "#" + b.id + " session " + b.sessionId + ", " + b.status + ", " + (b.totalPrice || 0).toLocaleString("en-US")).join("<br/>") + "</div>";
        } else {
          html += "<div style=\"margin-top:4px;color:var(--muted);\">No bookings.</div>";
        }
        el.innerHTML = html;
      } catch (e) {
        el.innerHTML = "<span style=\"color:var(--danger);\">Network error.</span>";
      }
    }
    document.getElementById("btn-load-client-admin").addEventListener("click", () => loadClientData("admin-client-data", "admin-client-id"));
    document.getElementById("btn-load-client-cashier").addEventListener("click", () => loadClientData("cashier-client-data", "cashier-client-id"));

    async function openChangeSeatsPanel(bookingId, sessionId) {
      if (!bookingId || !sessionId) return;
      openRightPanel("Change seats · Booking #" + bookingId, "<p style=\"color:var(--muted);font-size:13px;\">Loading seats...</p>");
      try {
        const res = await fetch("/api/sessions/" + sessionId + "/seats");
        const seats = await res.json().catch(() => []);
        if (!res.ok) { rightPanelContent.innerHTML = "<p style=\"color:var(--danger);\">Session not found.</p>"; return; }
        const rows = {};
        seats.forEach(s => {
          const r = s.rowNumber || 0;
          if (!rows[r]) rows[r] = [];
          rows[r].push(s);
        });
        const rowNums = Object.keys(rows).map(Number).sort((a,b) => a - b);
        let html = "<div class=\"right-panel-content\"><p style=\"font-size:12px;color:var(--muted);margin-bottom:10px;\">Select new seats for this booking (same session).</p><div id=\"rp-seat-grid\" class=\"seat-grid\" style=\"gap:4px;justify-content:flex-start;\"></div><p style=\"font-size:12px;color:var(--muted);margin-top:10px;\"><span id=\"rp-seat-count\">0</span> seat(s) selected</p><button type=\"button\" id=\"rp-btn-apply-seats\">Apply change</button></div>";
        rightPanelContent.innerHTML = html;
        const gridEl = rightPanelContent.querySelector("#rp-seat-grid");
        const countEl = rightPanelContent.querySelector("#rp-seat-count");
        const selectedIds = new Set();
        function updateCount() { countEl.textContent = selectedIds.size; }
        rowNums.forEach(rNum => {
          const rowDiv = document.createElement("div");
          rowDiv.className = "seat-row";
          rowDiv.style.display = "flex";
          rowDiv.style.alignItems = "center";
          rowDiv.style.gap = "4px";
          rowDiv.appendChild(document.createElement("span")).className = "seat-row-label";
          rowDiv.querySelector(".seat-row-label").textContent = rowLabel(rowNums.indexOf(rNum));
          (rows[rNum] || []).sort((a,b) => (a.seatNumber || 0) - (b.seatNumber || 0)).forEach(s => {
            const seatEl = document.createElement("div");
            seatEl.className = "seat";
            if (s.booked) seatEl.classList.add("unavailable");
            seatEl.dataset.seatId = s.id;
            seatEl.innerHTML = "<span></span>";
            seatEl.addEventListener("click", () => {
              if (s.booked) return;
              if (selectedIds.has(String(s.id))) { selectedIds.delete(String(s.id)); seatEl.classList.remove("selected"); }
              else { selectedIds.add(String(s.id)); seatEl.classList.add("selected"); }
              updateCount();
            });
            rowDiv.appendChild(seatEl);
          });
          gridEl.appendChild(rowDiv);
        });
        rightPanelContent.querySelector("#rp-btn-apply-seats").addEventListener("click", async () => {
          if (selectedIds.size === 0) { alert("Select at least one seat."); return; }
          try {
            const r = await fetch("/api/bookings/" + bookingId, { method: "PATCH", headers: authHeaders(), body: JSON.stringify({ seatIds: Array.from(selectedIds).map(Number) }) });
            const data = await r.json().catch(() => ({}));
            if (!r.ok) { alert(data.error || "Error"); return; }
            alert("Seats updated.");
            closeRightPanel();
            loadAllBookings("cashier-bookings-list");
          } catch (e) { alert("Error: " + e.message); }
        });
      } catch (e) {
        rightPanelContent.innerHTML = "<p style=\"color:var(--danger);\">Network error.</p>";
      }
    }

    loadHalls();
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

	// Сеансы: список и по movieId — без авторизации; создание — admin.
	protectedSessions := middleware.RequireRoleForMethods(
		sessionHandler,
		jwtSecret,
		map[string][]string{http.MethodPost: {"admin"}, http.MethodPut: {"admin"}},
	)
	http.Handle("/api/sessions", protectedSessions)
	http.Handle("/api/sessions/", protectedSessions)

	// Бронирования: требуют авторизации (JWT).
	protectedBookings := middleware.RequireAuth(bookingHandler, jwtSecret)
	http.Handle("/api/bookings", protectedBookings)
	http.Handle("/api/bookings/", protectedBookings)

	// Залы (список для UI).
	http.HandleFunc("/api/halls", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		halls, _ := hallRepo.GetAll()
		if halls == nil {
			halls = []*model.Hall{}
		}
		_ = json.NewEncoder(w).Encode(halls)
	})

	// Данные клиента (только admin/cashier).
	http.Handle("/api/users/", middleware.RequireAuth(clientHandler, jwtSecret))

	// Профиль текущего пользователя.
	http.Handle("/api/me", middleware.RequireAuth(profileHandler, jwtSecret))

	// Локальные постеры: папка web/posters раздаётся по /posters/
	postersDir := "web/posters"
	_ = os.MkdirAll(postersDir, 0755)
	http.Handle("/posters/", http.StripPrefix("/posters/", http.FileServer(http.Dir(postersDir))))

	// Загрузка постера (admin): сохраняет файл в web/posters, возвращает URL.
	uploadPoster := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		const maxSize = 10 << 20 // 10 MB
		if err := r.ParseMultipartForm(maxSize); err != nil {
			http.Error(w, `{"error":"invalid form or file too large"}`, http.StatusBadRequest)
			return
		}
		file, _, err := r.FormFile("poster")
		if err != nil {
			http.Error(w, `{"error":"missing file field 'poster'"}`, http.StatusBadRequest)
			return
		}
		defer file.Close()
		origName := r.MultipartForm.File["poster"][0].Filename
		ext := strings.ToLower(filepath.Ext(origName))
		if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".gif" && ext != ".webp" {
			http.Error(w, `{"error":"allowed: jpg, png, gif, webp"}`, http.StatusBadRequest)
			return
		}
		base := strings.TrimSuffix(filepath.Base(origName), ext)
		safe := ""
		for _, c := range base {
			if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_' {
				safe += string(c)
			}
		}
		if safe == "" {
			safe = "poster"
		}
		name := safe + ext
		fpath := filepath.Join(postersDir, name)
		dst, err := os.Create(fpath)
		if err != nil {
			http.Error(w, `{"error":"failed to save file"}`, http.StatusInternalServerError)
			return
		}
		defer dst.Close()
		if _, err := io.Copy(dst, file); err != nil {
			os.Remove(fpath)
			http.Error(w, `{"error":"failed to write file"}`, http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"url": "/posters/" + name})
	})
	http.Handle("/api/upload-poster", middleware.RequireRoleForMethods(uploadPoster, jwtSecret, map[string][]string{http.MethodPost: {"admin"}}))

	// Аутентификация (регистрация создаёт только клиентов).
	http.HandleFunc("/api/auth/register", authHandler.Register)
	http.HandleFunc("/api/auth/login", authHandler.Login)

	fmt.Println("NotFlix – Final (Assignment 4)")
	fmt.Println("Server listening on http://localhost" + port)
	fmt.Println("  GET  /health            – health check")
	fmt.Println("  GET  /api/movies        – list movies")
	fmt.Println("  GET  /api/movies/:id    – get movie")
	fmt.Println("  POST /api/movies        – create movie (admin, JSON)")
	fmt.Println("  GET  /api/sessions      – list sessions (?movieId=)")
	fmt.Println("  GET  /api/sessions/:id/seats – seats for session")
	fmt.Println("  POST /api/sessions      – create session (admin)")
	fmt.Println("  GET  /api/bookings      – my bookings (auth)")
	fmt.Println("  POST /api/bookings      – create booking (auth)")
	fmt.Println("  POST /api/auth/login    – login (email, password)")
	fmt.Println("  POST /api/auth/register – register")
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}
}
