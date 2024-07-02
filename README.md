# Task Manager

A task manager application built with Go, Fiber, and MongoDB. This application provides a RESTful API for user authentication and task management.

## Author

Bipin Kumar Ojha (Freelancer)

## Features

- User sign-up and sign-in with JWT authentication
- Create, read, update, and delete tasks
- Middleware for JWT authentication

## Technologies

- Go
- Fiber
- MongoDB
- JWT

## Getting Started

### Prerequisites

- Go (1.16 or later)
- MongoDB
- Docker (optional, for running MongoDB in a container)

### Installation

1. Clone the repository:

    ```sh
    git clone https://github.com/yourusername/taskmanager.git
    cd taskmanager
    ```

2. Set up environment variables:

    Create a `.env` file in the `config` directory with the following variables:

    ```env
    MONGO_URI=<your-mongodb-uri>
    TEST_MONGO_URI=<your-test-mongodb-uri>
    JWT_SECRET=<your-jwt-secret>
    APP_PORT=<your-app-port>
    TOKEN_EXPIRY_TIME=<expiry-time-in-second>
    ```

3. Install dependencies:

    ```sh
    go mod tidy
    ```

4. Run the application:

    ```sh
    go run main.go
    ```

### Running Tests

To run the tests, use the following command:

```sh
go test ./... -v
```
### API Endpoints
### 1. User Authentication
**Sign Up**
```
    URL: /signup
    Method: POST
    Body: json
          {
            "username": "testuser",
            "password": "testpassword"
          }

    Responses:
        201 Created: User created successfully
        400 Bad Request: Invalid request data
```
**Sign In**
```
    URL: /signin
    Method: POST
    Body: json
          {
            "username": "testuser",
            "password": "testpassword"
          }

    Responses:
        200 OK: Successful authentication, returns JWT token
        401 Unauthorized: Invalid username or password
```
**Sign Out**
```
    URL: /signout
    Method: POST
    Headers: 
        Authorization: <token>

    Responses:
        200 OK: Successful sign-out
        401 Unauthorized: Invalid or missing token
```
### 2. Task Management
**Create Task**
```
    URL: /tasks
    Method: POST
    Headers:
        Authorization: <token>
    Body:
    json
    {
        "title": "Test Task",
        "description": "This is a test task"
        "allotted_to": "testuser",
        "done_by": "",
        "status": "Pending",
        "start_time": "2024-07-01T00:00:00Z",
        "end_time": "2024-07-02T00:00:00Z"
    }

    Responses:
        201 Created: Task created successfully
        400 Bad Request: Invalid request data
        401 Unauthorized: Invalid or missing token
```
**Get All Tasks**
```
    URL: /tasks
    Method: GET
    Headers:
        Authorization: <token>

    Responses:
        200 OK: Returns a list of tasks
        401 Unauthorized: Invalid or missing token
```
**Get Task by ID**
```
    URL: /tasks/:id
    Method: GET
    Headers:
        Authorization: <token>

    Responses:
        200 OK: Returns the task with the given ID
        401 Unauthorized: Invalid or missing token
        404 Not Found: Task not found
```
**Update Task**
```
    URL: /tasks/:id
    Method: PUT
    Headers:
        Authorization: <token>
    Body:
    json
    {
        "title": "Updated Task",
        "description": "This is an updated task"
    }

    Responses:
        200 OK: Task updated successfully
        400 Bad Request: Invalid request data
        401 Unauthorized: Invalid or missing token
        404 Not Found: Task not found
```
**Delete Task**
```
    URL: /tasks/:id
    Method: DELETE
    Headers:
        Authorization: <token>

    Responses:
        200 OK: Task deleted successfully
        401 Unauthorized: Invalid or missing token
        404 Not Found: Task not found
```
### Project Structure

```
.
├── config
│   ├── .env
├── database
│   ├── database.go
│   ├── database_test.go
├── handlers
│   ├── handlers_test.go
│   ├── tasks.go
│   ├── users.go
├── helper
│   ├── helper.go
├── middleware
│   ├── middleware.go
├── models
│   ├── models.go
├── utils
│   └── utils.go
├── .gitignore
├── go.mod
├── go.sum
├── main.go
├── LICENSE
└── README.md
```

### License

This project is licensed under the MIT License.

```go
This `README.md` includes a detailed list of the API endpoints and their usage, as well as the project structure. Make sure to replace placeholder values like `<your-mongodb-uri>` and `<your-jwt-secret>` with actual values relevant to your setup.
```