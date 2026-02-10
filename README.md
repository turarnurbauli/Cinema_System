# Cinema System – Assignment 3 & 4

Project for Advanced Programming 1 (Assignment 3: Design; Assignment 4: Core Implementation).

**GitHub:** [https://github.com/turarnurbauli/ADP-3](https://github.com/turarnurbauli/ADP-3)

## Team
- **Alkhan Almas** (Se-2425) — [GitHub @AlmasAlkhan](https://github.com/AlmasAlkhan)
- **Tuar Nurbauli** (Se-2425) — [GitHub @turarnurbauli](https://github.com/turarnurbauli)

## Running the Project

```bash
go run .
```

Server starts at **http://localhost:8080**.

### Assignment 4 – API (Postman / curl)

| Method | URL | Description |
|--------|-----|-------------|
| GET | /health | Health check (JSON) |
| GET | /api/movies | List all movies |
| GET | /api/movies/:id | Get movie by ID |
| POST | /api/movies | Create movie (JSON body) |
| PUT | /api/movies/:id | Update movie |
| DELETE | /api/movies/:id | Delete movie |

Example – create movie:
```bash
curl -X POST http://localhost:8080/api/movies -H "Content-Type: application/json" -d "{\"title\":\"Inception\",\"description\":\"Sci-fi\",\"duration\":148,\"genre\":\"Sci-Fi\",\"rating\":8.8}"
```

## Project Structure

```
cinema-system/
├── docs/             # Documentation (Assignments 3 & 4)
│   ├── 01_Project_Proposal.md
│   ├── 02_Architecture_Design.md
│   ├── 03_Project_Plan.md
│   ├── 04_Diagrams_Mermaid.md
│   └── 05_Assignment4_Requirements.md
├── model/            # Domain models (ERD)
├── repository/       # In-memory storage (safe concurrency)
├── service/          # Business logic
├── handler/          # HTTP handlers (JSON)
├── main.go           # Server + goroutine
└── go.mod
```

## Documentation

- **[Project Proposal](docs/01_Project_Proposal.md)** – relevance, competitors, users
- **[Architecture & Design](docs/02_Architecture_Design.md)** – monolith, ERD, UML
- **[Project Plan](docs/03_Project_Plan.md)** – Gantt weeks 7–10
- **[Diagrams (Mermaid)](docs/04_Diagrams_Mermaid.md)** – Use-Case, ERD, UML
- **[Assignment 4 Requirements](docs/05_Assignment4_Requirements.md)** – checklist & demo script
- **[DEFENSE.md](DEFENSE.md)** – defense preparation

## Assignment 3 ✅

✅ Proposal, Architecture, Plan, Repository (50/50)

✅ Project Proposal (30%)  
✅ Architecture & Diagrams (35%)  
✅ Project Plan (20%)  
✅ Repository Setup (15%) – Git repository with branches and commits

## Assignment 4 (Milestone 2) ✅

✅ HTTP server (`net/http`), 3+ endpoints, JSON  
✅ Data model `Movie` (ERD), CRUD, MongoDB Atlas storage  
✅ One goroutine (background heartbeat logger)  
✅ Git: feature branches, 2+ commits per member  
