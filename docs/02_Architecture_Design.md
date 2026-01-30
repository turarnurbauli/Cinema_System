# Architecture & Design: Cinema System

## 1. Overall System Architecture

### Architectural Approach: Monolithic Architecture

At this stage, the project uses a monolithic architecture where all system components are in a single application. This simplifies development and deployment in the initial stages of the project.

### System Components:

```
┌─────────────────────────────────────────────────┐
│           HTTP Server (REST API)                │
├─────────────────────────────────────────────────┤
│  Handlers Layer (HTTP handlers)                 │
├─────────────────────────────────────────────────┤
│  Service Layer (Business logic)                  │
├─────────────────────────────────────────────────┤
│  Repository Layer (Data access)                 │
├─────────────────────────────────────────────────┤
│  Database (PostgreSQL)                          │
└─────────────────────────────────────────────────┘
```

## 2. System Modules

### 2.1. Movie Management Module (Movies)
**Responsibilities:**
- CRUD operations with movies
- Storage of movie information (title, description, duration, genre, rating)

**Main Entities:**
- Movie

### 2.2. Hall Management Module (Halls)
**Responsibilities:**
- Cinema hall management
- Seat configuration in halls
- Definition of seat types (regular, VIP, etc.)

**Main Entities:**
- Hall
- Seat

### 2.3. Session Management Module (Sessions)
**Responsibilities:**
- Session creation and management
- Linking movie, hall, and showtime
- Managing seat availability

**Main Entities:**
- Session

### 2.4. Booking Module (Bookings)
**Responsibilities:**
- Booking creation
- Booking status management
- Payment processing (simulation)

**Main Entities:**
- Booking
- Ticket

### 2.5. User Module (Users)
**Responsibilities:**
- Authentication and authorization
- User role management
- User profiles

**Main Entities:**
- User
- Role

## 3. Diagrams

**Визуальные диаграммы (Mermaid), которые отображаются как графики в GitHub и VS Code, находятся в [04_Diagrams_Mermaid.md](04_Diagrams_Mermaid.md):** Use-Case, ERD, UML, архитектура слоёв, сценарий бронирования. Ниже — те же диаграммы в текстовом (ASCII) виде.

### 3.1. Use-Case Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                        Cinema System                         │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌──────────────┐                                            │
│  │   Customer   │                                            │
│  └──────┬───────┘                                            │
│         │                                                     │
│         ├── View session schedule                            │
│         ├── Search movies                                    │
│         ├── View movie information                          │
│         ├── Book tickets                                    │
│         └── Purchase tickets                                │
│                                                              │
│  ┌──────────────┐                                            │
│  │ Administrator│                                            │
│  └──────┬───────┘                                            │
│         │                                                     │
│         ├── Manage movies                                    │
│         ├── Manage halls                                     │
│         ├── Manage sessions                                  │
│         ├── View statistics                                  │
│         └── Manage users                                     │
│                                                              │
│  ┌──────────────┐                                            │
│  │   Cashier    │                                            │
│  └──────┬───────┘                                            │
│         │                                                     │
│         ├── Sell tickets                                     │
│         ├── View available seats                            │
│         └── Refund tickets                                  │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

### 3.2. ERD (Entity Relationship Diagram)

```
┌─────────────┐         ┌─────────────┐         ┌─────────────┐
│    User     │         │    Movie    │         │    Hall     │
├─────────────┤         ├─────────────┤         ├─────────────┤
│ id (PK)     │         │ id (PK)     │         │ id (PK)     │
│ email       │         │ title       │         │ name        │
│ password    │         │ description │         │ capacity    │
│ name        │         │ duration    │         │ seat_layout │
│ role_id(FK) │         │ genre       │         └─────────────┘
└──────┬──────┘         │ rating      │                 │
       │                └──────┬───────┘                 │
       │                       │                         │
       │                ┌──────▼───────┐         ┌──────▼───────┐
       │                │   Session     │         │     Seat      │
       │                ├───────────────┤         ├───────────────┤
       │                │ id (PK)       │         │ id (PK)       │
       │                │ movie_id (FK) │         │ hall_id (FK)  │
       │                │ hall_id (FK)  │         │ row_number    │
       │                │ start_time    │         │ seat_number   │
       │                │ price         │         │ seat_type     │
       │                └──────┬────────┘         └───────────────┘
       │                       │
       │                ┌──────▼────────┐
       │                │   Booking     │
       │                ├───────────────┤
       │                │ id (PK)       │
       │                │ user_id (FK)  │
       │                │ session_id(FK)│
       │                │ status        │
       │                │ total_price   │
       │                │ created_at    │
       │                └──────┬────────┘
       │                       │
       │                ┌──────▼────────┐
       │                │    Ticket     │
       │                ├───────────────┤
       │                │ id (PK)      │
       │                │ booking_id(FK)│
       │                │ seat_id (FK)  │
       │                │ price         │
       │                └───────────────┘
       │
┌──────▼──────┐
│    Role     │
├─────────────┤
│ id (PK)     │
│ name        │
│ permissions │
└─────────────┘
```

### 3.3. UML Class Diagram (main classes)

```
┌─────────────────────────────────────────────────────────────┐
│                        Models                                │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌──────────────┐      ┌──────────────┐      ┌─────────────┐│
│  │    User      │      │    Movie     │      │    Hall     ││
│  ├──────────────┤      ├──────────────┤      ├─────────────┤│
│  │ +ID: int     │      │ +ID: int     │      │ +ID: int    ││
│  │ +Email: str  │      │ +Title: str  │      │ +Name: str  ││
│  │ +Password:str│      │ +Description:│      │ +Capacity:int│
│  │ +Name: str   │      │ +Duration: int│      └─────────────┘│
│  │ +RoleID: int │      │ +Genre: str  │                      │
│  └──────┬───────┘      │ +Rating: float│                     │
│         │              └──────┬───────┘                      │
│         │                     │                              │
│  ┌──────▼───────┐      ┌──────▼───────┐      ┌─────────────┐│
│  │    Role      │      │   Session    │      │     Seat    ││
│  ├──────────────┤      ├──────────────┤      ├─────────────┤│
│  │ +ID: int     │      │ +ID: int     │      │ +ID: int    ││
│  │ +Name: str   │      │ +MovieID: int│      │ +HallID: int││
│  │ +Permissions:│      │ +HallID: int │      │ +RowNum: int││
│  └──────────────┘      │ +StartTime:  │      │ +SeatNum:int││
│                        │ +Price: float│      │ +Type: str  ││
│                        └──────┬───────┘      └─────────────┘│
│                               │                             │
│                        ┌──────▼───────┐                     │
│                        │   Booking    │                     │
│                        ├──────────────┤                     │
│                        │ +ID: int     │                     │
│                        │ +UserID: int │                     │
│                        │ +SessionID:int│                    │
│                        │ +Status: str │                     │
│                        │ +TotalPrice: │                     │
│                        └──────┬───────┘                     │
│                               │                             │
│                        ┌──────▼───────┐                     │
│                        │    Ticket    │                     │
│                        ├──────────────┤                     │
│                        │ +ID: int     │                     │
│                        │ +BookingID:  │                     │
│                        │ +SeatID: int │                     │
│                        │ +Price: float│                     │
│                        └──────────────┘                     │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                    Service Layer                             │
├─────────────────────────────────────────────────────────────┤
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │ MovieService │  │SessionService│  │BookingService│      │
│  ├──────────────┤  ├──────────────┤  ├──────────────┤      │
│  │ +Create()    │  │ +Create()    │  │ +Create()    │      │
│  │ +GetByID()   │  │ +GetByID()   │  │ +GetByID()   │      │
│  │ +Update()    │  │ +GetAll()    │  │ +Cancel()    │      │
│  │ +Delete()    │  │ +GetAvailable│  │ +Confirm()   │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                  Repository Layer                            │
├─────────────────────────────────────────────────────────────┤
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │ MovieRepo    │  │SessionRepo   │  │BookingRepo   │      │
│  ├──────────────┤  ├──────────────┤  ├──────────────┤      │
│  │ +Create()    │  │ +Create()    │  │ +Create()    │      │
│  │ +GetByID()   │  │ +GetByID()   │  │ +GetByID()   │      │
│  │ +Update()    │  │ +GetAll()    │  │ +Update()    │      │
│  │ +Delete()    │  │ +GetByHall() │  │ +GetByUser() │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└─────────────────────────────────────────────────────────────┘
```

## 4. Data Flow

### Example: Ticket Booking Process

```
1. Client → HTTP Request → Handler (BookingHandler)
2. Handler → Service (BookingService)
3. Service → Repository (BookingRepository)
4. Repository → Database (PostgreSQL)
5. Database → Repository → Service
6. Service → Handler → HTTP Response → Client
```

### Detailed Flow:

```
[Client] 
  │ POST /api/bookings
  ▼
[BookingHandler]
  │ Validate request
  │ Extract user from session
  ▼
[BookingService]
  │ Validate session availability
  │ Check seat availability
  │ Calculate total price
  │ Create booking
  ▼
[BookingRepository]
  │ Begin transaction
  │ Insert booking
  │ Insert tickets
  │ Update seat availability
  │ Commit transaction
  ▼
[Database]
  │ Store data
  ▼
[Response back through layers]
```

## 5. Module Responsibilities

### Handler Layer
- HTTP request validation
- Input data parsing
- HTTP response formation
- HTTP error handling

### Service Layer
- Application business logic
- Business rule validation
- Coordination between repositories
- Transactional logic

### Repository Layer
- Data access abstraction
- SQL queries
- Data mapping from DB to models
- DB-level transaction management

### Model Layer
- Data structure definition
- Model-level data validation

## 6. Technical Details

### API Endpoints (planned):

**Movies:**
- `GET /api/movies` - list of movies
- `GET /api/movies/:id` - movie information
- `POST /api/movies` - create movie (admin)
- `PUT /api/movies/:id` - update movie (admin)
- `DELETE /api/movies/:id` - delete movie (admin)

**Sessions:**
- `GET /api/sessions` - list of sessions
- `GET /api/sessions/:id` - session information
- `GET /api/sessions/:id/seats` - available seats
- `POST /api/sessions` - create session (admin)

**Bookings:**
- `POST /api/bookings` - create booking
- `GET /api/bookings/:id` - booking information
- `DELETE /api/bookings/:id` - cancel booking

### Database:
- PostgreSQL for data storage
- Use of migrations for schema management
- Indexes on frequently used fields (email, session_id, etc.)
