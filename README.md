# 📚 Enterprise Library Management System REST API

[![Go Version](https://img.shields.io/badge/Go-1.25%2B-00ADD8?style=for-the-badge&logo=go)](https://golang.org/)
[![Gin Framework](https://img.shields.io/badge/Framework-Gin_v1.12-008080?style=for-the-badge&logo=gin)](https://gin-gonic.com/)
[![GORM ORM](https://img.shields.io/badge/ORM-GORM_v1.31-29BEB0?style=for-the-badge)](https://gorm.io/)
[![Database](https://img.shields.io/badge/Database-MySQL_8.0-4479A1?style=for-the-badge&logo=mysql)](https://www.mysql.com/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg?style=for-the-badge)](LICENSE)

A high-performance, enterprise-grade RESTful API for modern library operations built using **Go (Golang)**, **Gin Web Framework**, **GORM**, and **MySQL**. Engineered following **Clean Layered Architecture** (Controller → Service → Repository), featuring **pessimistic row locking** (`SELECT ... FOR UPDATE`) to prevent race conditions during book checkout, and a **concurrent worker pool** for automated batch background processing.

---

## ✨ Key Features

- 🏛️ **Clean Layered Architecture**: Decoupled design with interface abstractions across Controllers, Services, and Repositories for maximum testability and maintainability.
- 🔒 **ACID Transactions & Race-Free Checkout**: Utilizes GORM database transactions paired with MySQL pessimistic row locking (`SELECT ... FOR UPDATE`) to guarantee atomic copy count updates without over-booking.
- ⚡ **Concurrent Worker Pool**: Automated batch processing of overdue borrow records using 5 concurrent worker goroutines with Go channels and atomic counters.
- 🔐 **Authentication & RBAC**: Secure password hashing with `bcrypt`, JWT authentication (`golang-jwt/jwt`), and Role-Based Access Control (`ADMIN` vs `MEMBER`).
- ⏱️ **Resilience Middleware**: Built-in 10-second global request timeout middleware to prevent thread starvation, alongside custom CORS configuration.
- 📊 **Real-Time Admin Analytics**: Comprehensive dashboard metrics tracking active borrowings, overdue counts, total users, available books, and completed returns.
- 📝 **Structured Logging & Input Validation**: Production-grade logging using `Logrus` and strict input schema validation via `go-playground/validator/v10`.

---

## 🏗️ Architecture & System Design

### Layered Architecture Flow

```mermaid
graph TD
    Client[HTTP Client / Frontend] -->|JSON Requests| Router[Gin Router & Middleware]
    Router -->|Auth & Timeout Check| Controller[Controller Layer]
    Controller -->|DTO Validation| Service[Service Business Logic]
    Service -->|ACID Tx & Row Locking| Repository[Repository Data Access Layer]
    Repository -->|SQL Queries| DB[(MySQL Database)]

    subgraph Background Automation
        Admin[Admin / Cron Job] -->|Trigger Batch| WorkerPool[Concurrent Worker Pool - 5 Workers]
        WorkerPool -->|Process Overdue Records| Service
    end
```

### Pessimistic Concurrency Control Flow

```mermaid
sequenceDiagram
    autonumber
    actor User1 as User A (Checkout)
    actor User2 as User B (Checkout)
    participant DB as MySQL DB (Book ID: 42, Copies: 1)

    User1->>DB: BEGIN TRANSACTION & SELECT FOR UPDATE (Lock Row)
    User2->>DB: BEGIN TRANSACTION & SELECT FOR UPDATE (Blocked)
    Note over DB: User A holds row lock. User B waits in queue.
    DB-->>User1: Book Locked (Copies: 1)
    User1->>DB: Decrement Copies (1 -> 0) & Create Borrow Record
    User1->>DB: COMMIT TRANSACTION (Release Lock)
    DB-->>User2: Lock Acquired (Copies: 0)
    User2-->>User2: Check Copies (0 available)
    User2->>DB: ROLLBACK TRANSACTION
    User2-->>User2: Return 400 Bad Request ("No copies available")
```

---

## 🛠️ Phase-Wise Implementation Roadmap

### 🔹 Phase 1 — Project Setup
- [x] Create a new Go project module (`go.mod`).
- [x] Configure **Gin Web Framework** router engine.
- [x] Configure **GORM ORM** and MySQL driver initialization.
- [x] Establish `.env` environment configuration management using `godotenv`.
- [x] Create modular layered project structure:
  - `app`: Dependency Injection Container & Initializer.
  - `config`: Environment loader & DB connections.
  - `constants`: Roles (`ADMIN`, `MEMBER`) & Borrow Statuses (`BORROWED`, `RETURNED`, `OVERDUE`).
  - `controller`: HTTP Request Handlers & Input Binding.
  - `database`: GORM MySQL Connection Setup & Auto-Migration.
  - `dto`: Data Transfer Objects (Request/Response structs).
  - `middleware`: JWT Authentication, RBAC, and 10s Request Timeout.
  - `model`: Database Entities (`User`, `Book`, `BorrowRecord`).
  - `repository`: Data Access Layer & GORM Queries.
  - `route`: Gin Router Setup & Route Group Registration.
  - `service`: Business Logic Layer & ACID Transactions.
  - `utils`: JWT utilities, Password Hashing, Validation & Structured Logging.
  - `workers`: Goroutine Worker Pool for Overdue Records Batch Processing.
- [x] Add application configuration parsing.
- [x] Add database connection handling and auto-migration.

---

### 🔹 Phase 2 — Database & Models
- [x] **Create `users` table/model**:
  - Fields: `id`, `uuid`, `name`, `email` (unique constraint), `password`, `role` (`ADMIN`, `MEMBER`), `status`, `created_at`, `updated_at`, `deleted_at`.
- [x] **Create `books` table/model**:
  - Fields: `id`, `uuid`, `title`, `author`, `isbn` (unique constraint), `category`, `total_copies`, `available_copies`, `created_at`, `updated_at`, `deleted_at`.
- [x] **Create `borrow_records` table/model**:
  - Fields: `id`, `uuid`, `user_id`, `book_id`, `borrow_date`, `due_date`, `returned_at`, `status` (`BORROWED`, `RETURNED`, `OVERDUE`), `created_at`, `updated_at`, `deleted_at`.
- [x] **Define Model Relationships**:
  - User → Borrow Records (`1:N` foreign key relationship `user_id`).
  - Book → Borrow Records (`1:N` foreign key relationship `book_id`).
- [x] Add appropriate Primary Keys (`id`, `uuid`) and Foreign Keys.
- [x] Add unique constraint for user `email`.
- [x] Add unique constraint for book `isbn`.
- [x] Include automatic GORM timestamps (`created_at`, `updated_at`, `deleted_at`).

---

### 🔹 Phase 3 — Authentication & Authorization
- [x] **User Registration API** (`POST /user/register`):
  - Validate `name` presence.
  - Validate `email` format.
  - Validate `password` complexity (min 6 chars, uppercase, lowercase, number, special char).
  - Check and reject duplicate email registrations.
  - Hash password using `bcrypt` before database storage.
- [x] **User Login API** (`POST /user/login`):
  - Validate email and password inputs.
  - Verify password hash matching.
  - Generate signed JWT token containing `user_id`, `user_uuid`, and `role`.
- [x] **JWT Authentication Middleware** (`middleware.AuthMiddleware`):
  - Parse and validate Bearer JWT token from `Authorization` header.
  - Extract user ID, UUID, and role, populating request headers (`auth_user_id`, `user_id`, `user_type`).
- [x] **Role-Based Access Control Middleware** (`middleware.RequireRole`):
  - Enforce role permissions restricting specific endpoints to `ADMIN` or `MEMBER`.

---

### 🔹 Phase 4 — Book Management
- [x] **Create Book** (`POST /books/create` | `/api/v1/books`):
  - Allow access to `ADMIN` only.
  - Validate `title`, `author`, `isbn`, `category`, and `total_copies`.
  - Reject duplicate ISBN entries.
  - Automatically initialize `available_copies = total_copies`.
- [x] **Get Books** (`POST /books/list` | `/api/v1/books`):
  - Implement pagination (`limit`, `offset`).
  - Implement search filter by `title` and `author`.
  - Implement filtering by `category`.
  - Return pagination metadata (`filter_count`, `total_count`).
- [x] **Get Book** (`POST /books/details` | `/api/v1/books/:id`):
  - Fetch book details by UUID or ID.
  - Return proper `404 Not Found` when book does not exist.
- [x] **Update Book** (`POST /books/update` | `PUT /api/v1/books/:id`):
  - Allow access to `ADMIN` only.
  - Validate update request payload.
  - Recalculate `available_copies` dynamically when `total_copies` is modified.
- [x] **Delete Book** (`POST /books/delete` | `DELETE /api/v1/books/:id`):
  - Allow access to `ADMIN` only.
  - Soft-delete book record (`gorm.DeletedAt`) while preventing deletion if active borrow records exist.

---

### 🔹 Phase 5 — Borrow Book
- [x] **Borrow Book API** (`POST /borrow/create` | `/api/v1/books/:id/borrow`):
  - Verify authenticated user token.
  - Verify target book existence.
  - Check copy availability (`available_copies > 0`).
  - Prevent duplicate borrowing of the same book while an active record exists.
  - Enforce active borrow limit of **maximum 3 books** per user.
  - Set `due_date` automatically to **14 days** from borrowing date.
  - Execute GORM ACID database transaction:
    - Insert new `BorrowRecord` with status `BORROWED`.
    - Decrement `available_copies` by 1.

---

### 🔹 Phase 6 — Return Book
- [x] **Return Book API** (`POST /borrow/return` & `/borrow/return/:id` | `/api/v1/borrow-records/:id/return`):
  - Verify borrow record existence.
  - Verify logged-in user ownership of the borrowing record.
  - Prevent returning an already returned book (`status == RETURNED`).
  - Execute GORM ACID database transaction:
    - Update status to `RETURNED`.
    - Record exact timestamp in `returned_at`.
    - Increment `available_copies` by 1.

---

### 🔹 Phase 7 — My Borrowing History
- [x] **Borrow History API** (`POST /borrow/my-borrowings` | `/api/v1/my-borrowings`):
  - Retrieve borrowing history exclusively for the logged-in user using `auth_user_id`.
  - Support pagination (`limit`, `offset`).
  - Support status filter (`BORROWED`, `RETURNED`, `OVERDUE`).
  - Support date range filtering (`from_date`, `to_date`).
  - Return associated book metadata embedded alongside borrowing details.

---

### 🔹 Phase 8 — Admin Dashboard
- [x] **Admin Analytics API** (`POST /admin/dashboard` | `/api/v1/admin/dashboard`):
  - Restricted to `ADMIN` role.
  - Returns real-time aggregate metrics:
    - `total_books`: Total non-deleted books.
    - `total_users`: Total registered users.
    - `total_available_books`: Sum of all available copies across catalog.
    - `active_borrowings`: Count of books currently in `BORROWED` status.
    - `overdue_books`: Count of overdue borrowings (`due_date < NOW()`).
    - `completed_borrowings`: Total count of returned records (`RETURNED`).

---

### 🔹 Phase 9 — Concurrency & Locking
- [x] **Concurrent Borrowing Control**:
  - Handles multi-user simultaneous checkout attempts for popular titles.
  - Employs **MySQL Pessimistic Row Locking** (`SELECT ... FOR UPDATE` via `clause.Locking{Strength: "UPDATE"}`).
  - Guarantees that when `available_copies = 1`, exactly **1 user succeeds** while concurrent attempts fail gracefully with `400 Bad Request`.
  - Prevents race conditions and guarantees `available_copies` **never becomes negative**.

```text
Example Concurrency Execution (Available copies = 1):

   User A → Borrow  ──►  [Lock Row & Decrement (1 -> 0)]  ──►  ✅ Success (200 OK)
   User B → Borrow  ──►  [Wait for Lock -> Copies = 0]    ──►  ❌ Failed (400 Bad Request)
   User C → Borrow  ──►  [Wait for Lock -> Copies = 0]    ──►  ❌ Failed (400 Bad Request)
```

---

### 🔹 Phase 10 — Worker Pool Batch Processing
- [x] **Batch Overdue Worker** (`POST /admin/process-overdue` | `/api/v1/admin/process-overdue`):
  - Queries active borrow records where `due_date < NOW()` and `status = 'BORROWED'`.
  - Processes batch updates asynchronously using a **Worker Pool pattern**:
    - Spawns a **maximum of 5 concurrent worker goroutines**.
    - Uses Go **channels** (`chan model.BorrowRecord`) for job queue distribution.
    - Coordinates execution using `sync.WaitGroup`.
    - Tracks progress using thread-safe `sync/atomic` counters (`processed_count`, `failed_count`).
    - Ensures zero data races during record updates.

```text
Batch Execution Architecture:

  1000 Overdue Records  ──►  Job Queue Channel (chan model.BorrowRecord)
                                       │
            ┌──────────────┬───────────┼───────────┬──────────────┐
            ▼              ▼           ▼           ▼              ▼
        Worker 1       Worker 2    Worker 3    Worker 4       Worker 5
            │              │           │           │              │
            └──────────────┴───────────┼───────────┴──────────────┘
                                       ▼
                   Atomic Counters & WaitGroup Sync
                                       │
                                       ▼
                  HTTP JSON Summary Response to Admin
```

---

## 📁 Project Directory Structure

```text
library_management_backend/
├── app/                  # Application container & Dependency Injection (InitApp)
│   └── app.go
├── config/               # Database connection initialization & env config loader
│   └── config.go
├── constants/            # Role definitions & Borrow status constants
│   └── constants.go
├── controller/           # HTTP Request Handlers & DTO binding
│   ├── auth_controller.go
│   ├── book_controller.go
│   ├── borrow_controller.go
│   └── dashboard_controller.go
├── database/             # GORM MySQL driver config & schema auto-migration
│   └── database.go
├── dto/                  # Data Transfer Objects (Request/Response schemas)
│   ├── auth_dto.go
│   ├── book_dto.go
│   ├── borrow_dto.go
│   └── dashboard_dto.go
├── middleware/           # Auth JWT Middleware, Role RBAC & Request Timeout Middleware
│   ├── auth_middleware.go
│   ├── role_middleware.go
│   └── timeout_middleware.go
├── model/                # GORM Database Models (User, Book, BorrowRecord)
│   ├── book.go
│   ├── borrow_record.go
│   └── user.go
├── repository/           # Database Access Layer (SQL Queries & GORM locks)
│   ├── book_repository.go
│   ├── borrow_repository.go
│   └── user_repository.go
├── route/                # Gin Router Setup & Endpoint Registration
│   └── route.go
├── service/              # Business Logic Layer (ACID Transactions & Validations)
│   ├── auth_service.go
│   ├── book_service.go
│   ├── borrow_service.go
│   └── dashboard_service.go
├── utils/                # Response Builders, JWT Utilities, Password Hashing & Logger
│   ├── jwt_utils.go
│   ├── logger.go
│   ├── password_utils.go
│   ├── response_util.go
│   └── validator_util.go
└── workers/              # Concurrent Worker Pool for Batch Processing
    └── overdue_worker.go
```

---

## 🧰 Tech Stack & Dependencies

| Category | Technology / Library | Description |
|---|---|---|
| **Language** | [Go (Golang 1.25+)](https://golang.org/) | High-performance compiled language |
| **Web Framework** | [Gin v1.12.0](https://github.com/gin-gonic/gin) | Ultra-fast HTTP web framework |
| **ORM** | [GORM v1.31.2](https://gorm.io/) | Developer-friendly ORM library for Go |
| **Database** | [MySQL 8.0+](https://www.mysql.com/) | Relational Database Management System |
| **Authentication** | [golang-jwt/jwt/v5](https://github.com/golang-jwt/jwt) | JSON Web Token implementation |
| **Password Hashing** | [x/crypto/bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt) | Secure bcrypt password hashing |
| **Logging** | [Logrus v1.9.4](https://github.com/sirupsen/logrus) | Structured logger for Go |
| **Validation** | [Validator v10](https://github.com/go-playground/validator) | Struct & field validation |
| **Identifiers** | [Google UUID v1.6.0](https://github.com/google/uuid) | Universally Unique Identifiers generation |

---

## ⚙️ Environment Variables

Create a `.env` file in the root directory:

```env
# Server Configuration
PORT=8080

# Database Configuration
DB_HOST=127.0.0.1
DB_PORT=3306
DB_USER=root
DB_PASSWORD=your_mysql_password
DB_NAME=library_db

# Security & JWT Configuration
JWT_SECRET=super_secret_jwt_key_library_app
```

---

## 🚀 Quick Start & Local Setup

### Prerequisites

- **Go 1.25** or higher installed.
- **MySQL 8.0+** running locally or via Docker.

### 1. Clone & Navigate
```bash
git clone https://github.com/your-username/library_management_backend.git
cd library_management_backend
```

### 2. Configure Environment
Copy and customize `.env`:
```bash
cp .env.example .env
```

### 3. Initialize MySQL Database
Ensure your MySQL instance is active and create the database schema:
```sql
CREATE DATABASE IF NOT EXISTS library_db;
```

### 4. Install Dependencies
```bash
go mod download
```

### 5. Start Application
```bash
# Run server directly
go run .

# Run with custom env file
go run . -env custom.env
```
The server automatically runs GORM migrations and listens on `http://localhost:8080`.

---

## 📡 API Endpoints Reference

### 🔓 Public Endpoints

| Method | Endpoint | Description | Request Body |
|---|---|---|---|
| `POST` | `/health-check` | Server health check status | None |
| `POST` | `/user/register` | Register new user (`MEMBER` / `ADMIN`) | `RegisterUserRequest` |
| `POST` | `/user/login` | Authenticate user & receive JWT token | `LoginUserRequest` |
| `POST` | `/books/list` | Filter & paginate catalog | `GetBooksListRequest` |
| `POST` | `/books/details` | Retrieve book details by UUID | `GetBookDetailsRequest` |

### 🔒 Member Endpoints (Requires Auth Bearer Token)

| Method | Endpoint | Description | Request Headers |
|---|---|---|---|
| `POST` | `/borrow/create` | Borrow a book (Checks 3 max limit) | `Authorization`, `auth_user_id` |
| `POST` | `/borrow/return` | Return a borrowed book | `Authorization`, `auth_user_id` |
| `POST` | `/borrow/my-borrowings` | View personal borrowing history | `Authorization`, `auth_user_id` |
| `POST` | `/my-borrowings` | Alternative endpoint for borrowing history | `Authorization`, `auth_user_id` |

### 🛡️ Admin Endpoints (Requires Admin Bearer Token)

| Method | Endpoint | Description | Allowed Role |
|---|---|---|---|
| `POST` | `/books/create` | Add new book to catalog | `ADMIN` |
| `POST` | `/books/update` | Update existing book info | `ADMIN` |
| `POST` | `/books/delete` | Soft delete book record | `ADMIN` |
| `POST` | `/admin/dashboard` | Fetch system analytics stats | `ADMIN` |
| `POST` | `/admin/process-overdue` | Trigger background worker batch update | `ADMIN` |

---

## 💻 Sample cURL Requests

### 1. Register User
```bash
curl -X POST http://localhost:8080/user/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "password": "Password123!",
    "role": "MEMBER"
  }'
```

### 2. User Login
```bash
curl -X POST http://localhost:8080/user/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "Password123!"
  }'
```

### 3. Create Book (Admin Only)
```bash
curl -X POST http://localhost:8080/books/create \
  -H "Authorization: Bearer <ADMIN_JWT_TOKEN>" \
  -H "auth_user_id: 1" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Clean Architecture in Go",
    "author": "Robert C. Martin",
    "isbn": "978-0134494166",
    "category": "Software Engineering",
    "total_copies": 5
  }'
```

### 4. Borrow Book
```bash
curl -X POST http://localhost:8080/borrow/create \
  -H "Authorization: Bearer <MEMBER_JWT_TOKEN>" \
  -H "auth_user_id: 2" \
  -H "Content-Type: application/json" \
  -d '{
    "book_uuid": "<BOOK_UUID>"
  }'
```

### 5. Return Book
```bash
curl -X POST http://localhost:8080/borrow/return \
  -H "Authorization: Bearer <MEMBER_JWT_TOKEN>" \
  -H "auth_user_id: 2" \
  -H "Content-Type: application/json" \
  -d '{
    "borrow_record_uuid": "<BORROW_RECORD_UUID>"
  }'
```

### 6. Trigger Overdue Worker Batch Processing
```bash
curl -X POST http://localhost:8080/admin/process-overdue \
  -H "Authorization: Bearer <ADMIN_JWT_TOKEN>" \
  -H "auth_user_id: 1"
```

---

## ⚡ High-Throughput Batch Worker Pool

The project features an in-memory concurrent background worker engine (`workers/overdue_worker.go`) designed for efficient batch updates:

- **Fan-Out Architecture**: Spawns **5 worker goroutines** listening on a Go channel populated with overdue borrow records (`due_date < NOW() && status = 'BORROWED'`).
- **Atomic Operations**: Utilizes `sync/atomic` primitives (`AddInt64`) to compute total processed vs failed records safely across goroutines without lock contention.
- **Synchronization**: Coordinates completion using `sync.WaitGroup` to guarantee non-blocking execution while delivering exact execution stats in the admin API response.

---

## 🧪 Build & Test Verification

```bash
# Compile and build binary
go build ./...

# Run code format verification
go fmt ./...

# Run static analysis
go vet ./...
```

---

## 📄 License

This project is open-source software licensed under the **MIT License**.
