# Task Management API

Welcome to the **Task Management API**! This system enables users to manage tasks efficiently with features like user authentication, task creation, updates, and soft deletion.

## Features
- **User Authentication:** Secure registration and login with JWT.
- **User Activity Log:** User Activity is recorded like creating, updating, deleting task or registering, logging, account activation .
- **Task Management:** Create, read, update, delete (soft delete) tasks.
- **Task Filtering:** Search tasks using parameters such as `status`, `sort_by`, `page`, etc.
- **Caching:** Redis for performance optimization.
- **Logging:** Using `Lagrus` for structured logging.
- **Rate Limiting:** Goroutine-based rate limiter.
- **Server Error Handling:** Env-based maintenance mode.
- **Context Middleware:** Each request has a **5-second timeout** for better resource management.
- **Database Migrations:** Managed using `golang-migrate`.

## Technologies Used
- **GoLang:** Backend development
- **Gin:** HTTP web framework
- **MySQL:** Database management
- **Redis:** Caching mechanism
- **Bcrypt:** Secure password hashing
- **JWT:** Token-based authentication
- **GORM:** ORM for database interactions
- **Lagrus:** Logging system
- **SecureHeader:** Additional security headers

## API Endpoints

### **User Authentication**
- `POST /register` - Register a new user
- `GET /activation_token/:token` - Activate user account
- `POST /login` - Authenticate and receive JWT token

### **Task Management**
- `GET /tasks` - List tasks with filters (`limit`, `page`, `sort_by`, `status`, `sort_order`)
- `GET /tasks/:id` - Get a single task by ID
- `POST /tasks` - Create a new task
- `PUT /tasks/update/:id` - Update a task
- `DELETE /tasks/delete/:id` - Soft delete a task

## Getting Started

### **Prerequisites**
- Install **GoLang** ([Download](https://golang.org/dl/))
- Install **MySQL** ([Download](https://www.mysql.com/download/))
- Install **Redis** ([Install Guide](https://redis.io/docs/latest/operate/oss_and_stack/install/install-redis-on-linux/))
- Install `golang-migrate` for database migrations:
  ```sh
  go install -tags 'mysql' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
  ```

### **Installation**
1. Clone the repository:
   ```sh
   git clone github.com/iamgak/go-task
   ```
2. Navigate to the project directory:
   ```sh
   cd go-task
   ```
3. Install dependencies:
   ```sh
   go mod tidy
   ```
4. Create MySQL database:
   ```sh
   mysql -u root -p -e "CREATE DATABASE go_task;"
   ```
5. Apply database migrations:
   ```sh
   migrate -path=./migrations -database "mysql://root:password@tcp(127.0.0.1:3306)/go_task?parseTime=true" up
   ```
6. Run the server:
   ```sh
   go run cmd/cli
   ```

## Context Middleware (5-Second Timeout)
To prevent long-running requests and manage resources efficiently, a **global middleware** enforces a **5-second timeout** for each API request:
```go
func TimeoutMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
        defer cancel()
        c.Request = c.Request.WithContext(ctx)
        c.Next()
    }
}
```
This middleware is applied globally in `main.go`:
```go
r := gin.Default()
r.Use(TimeoutMiddleware())
```

### **Database Migrations**
To create a new migration:
```sh
migrate create -ext sql -dir ./migrations -seq create_tasks_table
```
To apply migrations:
```sh
migrate -path=./migrations -database "mysql://root:password@tcp(127.0.0.1:3306)/go_task?parseTime=true" up
```
To roll back the last migration:
```sh
migrate -path=./migrations -database "mysql://root:password@tcp(127.0.0.1:3306)/go_task?parseTime=true" down 1
```
If you get a **dirty database error**, reset the version:
```sh
migrate -path=./migrations -database "mysql://root:password@tcp(127.0.0.1:3306)/go_task?parseTime=true" force <version_number>
```

## Usage Examples
- **Register a new user:** `POST https://localhost:8000/register`
- **Login:** `POST https://localhost:8000/login`
- **Fetch tasks:** `GET https://localhost:8000/tasks`
- **Get a task by ID:** `GET https://localhost:8000/tasks/:id`
- **Create a task:** `POST https://localhost:8000/tasks`
- **Update a task:** `PUT https://localhost:8000/tasks/update/:id`
- **Soft delete a task:** `DELETE https://localhost:8000/tasks/delete/:id`

Example requests:
```sh
curl -X GET "localhost:8080/tasks?due_date_after=2024-10-01"
curl -X GET "localhost:8080/tasks?limit=1&page=1&sort_by=id&status=in_progress&due_date_before=2024-10-01&sort_order=asc"
curl -X GET "localhost:8080/activation_token/{verification_token}"
```

## Contributing
Contributions are welcome! Fork the repository and submit a pull request with your changes.

