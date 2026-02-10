# Assignment 4 – Milestone 2 (Core System Implementation)

**Deadline:** 23:59 01.02.2026  
**Submit:** ZIP `Assignment4_TeamName.zip` → Moodle  
**Run:** `go run .`

---

## Goal

Working core system based on Assignment 3 design: backend, data models matching ERD, 3–5 core features, basic persistence. No need to implement everything.

---

## Grading Rubric (Checklist)

### 1. Backend Application (30%)

- [x] HTTP server using `net/http` (or approved framework)
- [x] At least **3 working endpoints** (e.g. list, get by id, create)
- [x] **JSON** input and output

### 2. Data Model & Storage (25%)

- [x] Data structures **match ERD** from Assignment 3 (Movie: id, title, description, duration, genre, rating; etc.)
- [x] **CRUD** for at least **one core entity** (e.g. Movie)
- [x] **Safe concurrent access** (e.g. mutex, no crashes under concurrent requests)

### 3. Concurrency (15%)

- [x] Use at least **one goroutine**
- [x] Examples: background worker, async processing, channel-based logic

### 4. Git Workflow (15%)

- [x] Feature branches per team member (`alkhan-almas`, `nurbauli-turar`)
- [x] At least **2 commits per member**
- [x] Meaningful commit messages

### 5. Demo & Explanation (15%)

- [x] Demonstrate running system (`go run .`, then call endpoints e.g. via Postman)
- [x] Show implemented features
- [x] Explain how implementation follows Assignment 3 design (layers, ERD)

---

## What to Implement (Minimum)

| Requirement              | Example for Cinema System                          |
|--------------------------|-----------------------------------------------------|
| 3–5 core features        | List movies, get movie by ID, create movie, (update, delete) |
| Basic persistence         | In-memory store (slice + map) or database           |
| 3+ endpoints             | `GET /api/movies`, `GET /api/movies/:id`, `POST /api/movies` |
| CRUD for one entity      | Movie: Create, Read (one + list), Update, Delete    |
| One goroutine            | Background logger, cleanup task, or channel worker |

---

## What Is NOT Required (Yet)

- Full feature set
- Authentication / authorization
- Full UI
- Complete error handling
- Performance optimization

---

## Architecture (Must Follow Assignment 3)

- **Monolith:** one running backend
- **Layers:** Handlers → Service → Repository → Storage (in-memory or DB)
- **ERD:** Movie (id, title, description, duration, genre, rating); other entities as needed for 3–5 use cases

---

## Suggested Endpoints for Cinema System

1. `GET  /api/movies`        — list all movies (JSON)
2. `GET  /api/movies/:id`    — get one movie by ID (JSON)
3. `POST /api/movies`        — create movie (JSON body → JSON response)
4. `PUT  /api/movies/:id`    — update movie (optional)
5. `DELETE /api/movies/:id`  — delete movie (optional)

Health/readiness (optional): `GET /health` → `{"status":"ok"}`

---

## Demo Script (for Defense)

1. Run `go run .` — server starts (e.g. port 8080).
2. Postman/curl:
   - `GET http://localhost:8080/api/movies` → `[]` or list of movies
   - `POST http://localhost:8080/api/movies` with JSON body → 201 + created movie
   - `GET http://localhost:8080/api/movies/1` → one movie JSON
3. Show code: handlers → service → repository; model matches ERD; one goroutine (e.g. background task).
4. Show Git: branches, 2+ commits per member.

---

## Notes

- Implementation must follow the **architecture approved in Assignment 3**.
- Major design changes must be **justified during defense**.
- Code quality is less important than **correctness and clarity** at this stage.
