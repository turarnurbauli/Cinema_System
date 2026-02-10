# Диаграммы (Mermaid)

Диаграммы ниже отображаются как графики в **GitHub**, **VS Code** (расширение Mermaid) и на [mermaid.live](https://mermaid.live).  
Для экспорта в PNG: скопируйте код блока на [mermaid.live](https://mermaid.live) → Export → PNG.

---

## 1. Use-Case Diagram

```mermaid
flowchart LR
    subgraph Cinema System
        UC1[View session schedule]
        UC2[Search movies]
        UC3[View movie info]
        UC4[Book tickets]
        UC5[Purchase tickets]
        UC6[Manage movies]
        UC7[Manage halls]
        UC8[Manage sessions]
        UC9[View statistics]
        UC10[Manage users]
        UC11[Sell tickets]
        UC12[View available seats]
        UC13[Refund tickets]
    end

    Customer((Customer))
    Admin((Administrator))
    Cashier((Cashier))

    Customer --> UC1
    Customer --> UC2
    Customer --> UC3
    Customer --> UC4
    Customer --> UC5

    Admin --> UC6
    Admin --> UC7
    Admin --> UC8
    Admin --> UC9
    Admin --> UC10

    Cashier --> UC11
    Cashier --> UC12
    Cashier --> UC13
```

---

## 2. ERD (Entity Relationship Diagram)

```mermaid
erDiagram
    User ||--o{ Booking : makes
    User }o--|| Role : has
    Role {
        int id PK
        string name
        string permissions
    }
    User {
        int id PK
        string email
        string password
        string name
        int role_id FK
    }
    Movie {
        int id PK
        string title
        string description
        int duration
        string genre
        float rating
    }
    Hall {
        int id PK
        string name
        int capacity
        string seat_layout
    }
    Session }o--|| Movie : "movie"
    Session }o--|| Hall : "hall"
    Session {
        int id PK
        int movie_id FK
        int hall_id FK
        datetime start_time
        float price
    }
    Seat }o--|| Hall : "in"
    Seat {
        int id PK
        int hall_id FK
        int row_number
        int seat_number
        string seat_type
    }
    Booking }o--|| Session : "for"
    Booking {
        int id PK
        int user_id FK
        int session_id FK
        string status
        float total_price
        datetime created_at
    }
    Ticket }o--|| Booking : "part of"
    Ticket }o--|| Seat : "seat"
    Ticket {
        int id PK
        int booking_id FK
        int seat_id FK
        float price
    }
```

---

## 3. UML Class Diagram (Models + Services + Repositories)

```mermaid
classDiagram
    class User {
        +int ID
        +string Email
        +string Password
        +string Name
        +int RoleID
    }
    class Role {
        +int ID
        +string Name
        +string Permissions
    }
    class Movie {
        +int ID
        +string Title
        +string Description
        +int Duration
        +string Genre
        +float Rating
    }
    class Hall {
        +int ID
        +string Name
        +int Capacity
    }
    class Seat {
        +int ID
        +int HallID
        +int RowNumber
        +int SeatNumber
        +string Type
    }
    class Session {
        +int ID
        +int MovieID
        +int HallID
        +datetime StartTime
        +float Price
    }
    class Booking {
        +int ID
        +int UserID
        +int SessionID
        +string Status
        +float TotalPrice
    }
    class Ticket {
        +int ID
        +int BookingID
        +int SeatID
        +float Price
    }
    User --> Role : role_id
    Booking --> User : user_id
    Booking --> Session : session_id
    Ticket --> Booking : booking_id
    Ticket --> Seat : seat_id
    Seat --> Hall : hall_id
    Session --> Movie : movie_id
    Session --> Hall : hall_id
```

---

## 4. System Architecture (Layers)

```mermaid
flowchart TB
    subgraph Client
        HTTP[HTTP Request]
    end
    subgraph Monolith["Cinema System (Monolith)"]
        H[Handlers Layer]
        S[Service Layer]
        R[Repository Layer]
    end
    DB[(PostgreSQL)]

    HTTP --> H
    H --> S
    S --> R
    R --> DB
    DB --> R
    R --> S
    S --> H
    H --> HTTP
```

---

## 5. Data Flow: Ticket Booking

```mermaid
sequenceDiagram
    participant C as Client
    participant H as BookingHandler
    participant S as BookingService
    participant R as BookingRepository
    participant DB as Database

    C->>H: POST /api/bookings
    H->>H: Validate request
    H->>S: CreateBooking(...)
    S->>S: Check session & seats
    S->>R: Create booking + tickets
    R->>DB: BEGIN; INSERT; COMMIT
    DB-->>R: OK
    R-->>S: Booking
    S-->>H: Booking
    H-->>C: 201 Created
```
