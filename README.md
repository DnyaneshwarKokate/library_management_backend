# 📚 Library Management System — REST API (Go + Gin + GORM + MySQL)

A high-performance, enterprise-grade Library Management REST API built with **Go (Golang)**, **Gin Web Framework**, **GORM ORM**, and **MySQL**. The application strictly follows a clean **layered architecture** (Controller → Service → Repository) with interface abstraction, panic recovery, structured logging, concurrent worker pools, and database pessimistic row locking.

---

## 🏗️ Project Architecture & Folder Structure

```text
library_management_backend/
├── app/                  # Application container & dependency injection (InitApp)
├── config/               # Configuration loader (.env, Database)
├── constants/            # Role definitions & Borrow status constants
├── controller/           # HTTP Request Handlers & Input Validation (Interface + Struct)
├── database/             # GORM MySQL Connection Initialization & Migration
├── dto/                  # Data Transfer Objects (Request/Response schemas)
├── middleware/           # Auth JWT Middleware, Role RBAC & Request Timeout Middleware
├── model/                # GORM Database Models (User, Book, BorrowRecord)
├── repository/           # Database Access Layer (Interface + Struct + SQL Queries)
├── route/                # Gin Router Setup & Endpoint Registration
├── service/              # Business Logic Layer (Interface + Struct + Logrus Logging)
├── utils/                # Response Builders, JWT Utilities, Password Hashing
└── workers/              # Concurrent Worker Pool for Batch Processing
```

---

## 🚀 Phase-Wise Implementation Breakdown

### 🔹 Phase 1 — Project Setup & Architecture
- Initialized Go Module (`go.mod`) and structured layered architecture.
- Configured Gin HTTP router and GORM MySQL database connection.
- Implemented environment configuration using `.env`.
- Added dependency injection container (`app/app.go`) wiring controllers, services, and repositories.

### 🔹 Phase 2 — Database & Models
- Defined **GORM Database Models**:
  - `User`: `id`, `uuid`, `name`, `email` (unique constraint), `password`, `role` (`ADMIN`, `MEMBER`), `status`.
  - `Book`: `id`, `uuid`, `title`, `author`, `isbn` (unique constraint), `category`, `total_copies`, `available_copies`.
  - `BorrowRecord`: `id`, `uuid`, `user_id`, `book_id`, `borrow_date`, `due_date`, `returned_at`, `status` (`BORROWED`, `RETURNED`, `OVERDUE`).
- Set up auto-migration, GORM soft deletes (`gorm.DeletedAt`), foreign keys, and timestamps (`created_at`, `updated_at`).

### 🔹 Phase 3 — Authentication & Authorization
- **User Registration** (`POST /user/register`): Validates name, email, and password strength (min 6 chars, uppercase, lowercase, number, special char). Hashes passwords using bcrypt.
- **User Login** (`POST /user/login`): Verifies credentials and generates JWT tokens containing `user_id`, `user_uuid`, and `role`.
- **JWT Auth Middleware** (`middleware.AuthMiddleware`): Validates Bearer tokens and automatically populates request headers (`auth_user_id`, `user_id`, `user_type`).
- **Role-Based Access Control** (`middleware.RequireRole`): Enforces role authorization (e.g. restricts book creation and admin stats to `ADMIN` users).

### 🔹 Phase 4 — Book Management API
- **List Books** (`POST /books/list`): Supports search filter, category filter, pagination (`limit`, `offset`), returning `data`, `filter_count`, and `total_count`.
- **Book Details** (`POST /books/details`): Fetches book details by UUID.
- **Create Book** (`POST /books/create`): Admin endpoint to add books (validates unique ISBN).
- **Update Book** (`POST /books/update`): Admin endpoint to modify book metadata and copy numbers.
- **Delete Book** (`POST /books/delete`): Admin endpoint to soft-delete books.

### 🔹 Phase 5 — Borrow Book API
- **Borrow Book** (`POST /borrow/create`):
  - Validates user active status and book availability (`available_copies > 0`).
  - Enforces active borrow limit of **maximum 3 books** per user.
  - Prevents duplicate borrowing of the same book while active.
  - Calculates `due_date` (14 days from borrowing date).
  - Executes ACID database transaction decrementing available copies and creating the record.

### 🔹 Phase 6 — Return Book API
- **Return Book** (`POST /borrow/return` & `POST /borrow/return/:id`):
  - Verifies user ownership of the borrow record.
  - Validates that the record is not already returned.
  - Executes ACID database transaction updating record status to `RETURNED`, setting `returned_at`, and incrementing `available_copies`.

### 🔹 Phase 7 — My Borrowing History API
- **Borrow History** (`POST /borrow/my-borrowings` & `POST /my-borrowings`):
  - Retrieves logged-in user's borrowing history using `auth_user_id` header.
  - Supports filtering by status (`BORROWED`, `RETURNED`, `OVERDUE`) and date range (`from_date`, `to_date`).
  - Supports pagination (`limit`, `offset`) with `filter_count` and `total_count`.

### 🔹 Phase 8 — Admin Dashboard API
- **Admin Dashboard** (`POST /admin/dashboard`):
  - Admin-only analytics endpoint returning real-time metrics:
    - `total_books`: Total non-deleted books.
    - `total_users`: Total registered users.
    - `total_available_books`: Sum of all available copies.
    - `active_borrowings`: Count of books currently `BORROWED`.
    - `overdue_books`: Count of overdue borrowings (`due_date < NOW()`).
    - `completed_borrowings`: Count of returned books (`RETURNED`).

### 🔹 Phase 9 — Concurrency Control
- **Database Row Locking (`SELECT ... FOR UPDATE`)**:
  - Implemented GORM pessimistic locking clause (`clause.Locking{Strength: "UPDATE"}`) inside the `BorrowBook` ACID transaction.
  - Ensures that when multiple users try to borrow the last available copy (`available_copies = 1`) at the exact same millisecond, MySQL queues the requests. Exactly **1 user succeeds** while all other concurrent requests receive `400 Bad Request: no copies available`.
  - Guarantees `available_copies` can **never drop below 0**.

### 🔹 Phase 10 — Worker Pool Batch Processing
- **Process Overdue Worker** (`POST /admin/process-overdue`):
  - Admin batch processing endpoint for updating overdue records.
  - Queries active borrowings where `due_date < NOW()`.
  - Processes batch records using a **Worker Pool with 5 concurrent goroutines**, Go channels (`chan model.BorrowRecord`), `sync.WaitGroup`, and `sync/atomic` counters.
  - Returns `total_overdue_found`, `processed_count`, and `failed_count`.

---

## 🛠️ Middleware Features

1. **Header Extraction**:
   - Header `auth_user_id`: Primary uint user ID.
   - Header `user_id`: User UUID string.
   - Header `user_type`: Role (`ADMIN` / `MEMBER`).
2. **Request Timeout Middleware** (`middleware.TimeoutMiddleware`):
   - Enforces a **10-second request timeout** globally across all endpoints.
   - Aborts slow requests and returns `504 Gateway Timeout`.

---

## 📋 API Endpoints Reference & cURL Examples

### 1. Health Check
```bash
curl -X POST http://localhost:8080/health-check
```

### 2. User Authentication
```bash
# Register User
curl -X POST http://localhost:8080/user/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "password": "Password123!",
    "role": "MEMBER"
  }'

# Login User
curl -X POST http://localhost:8080/user/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "Password123!"
  }'
```

### 3. Book Management
```bash
# List Books
curl -X POST http://localhost:8080/books/list \
  -H "Content-Type: application/json" \
  -d '{"limit": 10, "offset": 0, "search": "Golang"}'

# Create Book (Admin)
curl -X POST http://localhost:8080/books/create \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <ADMIN_TOKEN>" \
  -H "auth_user_id: 1" \
  -d '{
    "title": "Clean Architecture in Go",
    "author": "Robert C. Martin",
    "isbn": "978-0134494166",
    "category": "Technology",
    "total_copies": 5
  }'
```

### 4. Borrowing & Returning
```bash
# Borrow Book
curl -X POST http://localhost:8080/borrow/create \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <TOKEN>" \
  -H "auth_user_id: 2" \
  -d '{"book_uuid": "<BOOK_UUID>"}'

# Return Book
curl -X POST http://localhost:8080/borrow/return \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <TOKEN>" \
  -H "auth_user_id: 2" \
  -d '{"borrow_record_uuid": "<RECORD_UUID>"}'

# My Borrowings History
curl -X POST http://localhost:8080/borrow/my-borrowings \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <TOKEN>" \
  -H "auth_user_id: 2" \
  -d '{"limit": 10, "status": "BORROWED"}'
```

### 5. Admin Utilities
```bash
# Admin Dashboard Stats
curl -X POST http://localhost:8080/admin/dashboard \
  -H "Authorization: Bearer <ADMIN_TOKEN>" \
  -H "auth_user_id: 1"

# Process Overdue Records (Worker Pool)
curl -X POST http://localhost:8080/admin/process-overdue \
  -H "Authorization: Bearer <ADMIN_TOKEN>" \
  -H "auth_user_id: 1"
```

---

## 💻 How to Run Locally

1. **Configure Environment Variables** (`.env`):
   ```env
   PORT=8080
   DB_HOST=127.0.0.1
   DB_PORT=3306
   DB_USER=root
   DB_PASSWORD=your_password
   DB_NAME=library_db
   JWT_SECRET=super_secret_jwt_key_library_app
   ```

2. **Run Server**:
   ```bash
   go run .
   ```

3. **Verify Build**:
   ```bash
   go build ./...
   ```
