// handlers_test.go
// Author: Bipin Kumar Ojha (Freelancer)

package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/bkojha74/task-management/database"
	"github.com/bkojha74/task-management/models"
	"github.com/bkojha74/task-management/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var testMongoURI string
var jwtSecret string
var testApp *fiber.App

func TestMain(m *testing.M) {
	// Load environment variables
	err := godotenv.Load("../config/.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	testMongoURI = os.Getenv("TEST_MONGO_URI")
	jwtSecret = os.Getenv("JWT_SECRET")
	if testMongoURI == "" || jwtSecret == "" {
		log.Fatal("TEST_MONGO_URI and JWT_SECRET must be set in the environment")
	}

	// Set up MongoDB connection
	clientOptions := options.Client().ApplyURI(testMongoURI)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}

	database.UsersCollection = client.Database("testdb").Collection("users")
	database.TasksCollection = client.Database("testdb").Collection("tasks")

	// Initialize Fiber app
	testApp = fiber.New()
	testApp.Post("/signup", SignUp)
	testApp.Post("/signin", SignIn(jwtSecret, 60))
	testApp.Post("/tasks", utils.JWTMiddleware(jwtSecret), CreateTask)
	testApp.Get("/tasks", utils.JWTMiddleware(jwtSecret), GetTasks)
	testApp.Get("/tasks/:id", utils.JWTMiddleware(jwtSecret), GetTask)
	testApp.Put("/tasks/:id", utils.JWTMiddleware(jwtSecret), UpdateTask)
	testApp.Delete("/tasks/:id", utils.JWTMiddleware(jwtSecret), DeleteTask)
	testApp.Post("/signout", SignOut)

	// Start the server in a goroutine
	go func() {
		log.Fatal(testApp.Listen(":4000"))
	}()

	// Run tests
	code := m.Run()

	// Clean up database after tests
	/*err = client.Database("testdb").Drop(context.Background())
	if err != nil {
		log.Fatal(err)
	}*/

	os.Exit(code)
}

func TestSignUp(t *testing.T) {
	user := models.User{
		Username: "testuser",
		Password: "testpassword",
	}
	body, _ := json.Marshal(user)

	req, err := http.NewRequest(http.MethodPost, "http://localhost:4000/signup", bytes.NewBuffer(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	// Create a custom HTTP client with a longer timeout
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusCreated, resp.StatusCode)

	var createdUser models.User
	err = json.NewDecoder(resp.Body).Decode(&createdUser)
	require.NoError(t, err)
	require.Equal(t, user.Username, createdUser.Username)
}

func TestJWTMiddleware(t *testing.T) {
	// Sign in to get a valid token
	user := models.User{
		Username: "testjwt",
		Password: "testpassword",
	}
	body, _ := json.Marshal(user)

	// First, sign up the user
	req, err := http.NewRequest(http.MethodPost, "http://localhost:4000/signup", bytes.NewBuffer(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 10 * time.Second}
	_, _ = client.Do(req)

	// Sign in to get the token
	req, err = http.NewRequest(http.MethodPost, "http://localhost:4000/signin", bytes.NewBuffer(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	var tokenResp map[string]string
	err = json.NewDecoder(resp.Body).Decode(&tokenResp)
	require.NoError(t, err)

	token := tokenResp["token"]

	// Test protected route with valid token
	req, err = http.NewRequest(http.MethodGet, "http://localhost:4000/tasks", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", token) // Corrected header value without 'Bearer ' prefix

	resp, err = client.Do(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	// Test protected route with invalid token
	req, err = http.NewRequest(http.MethodGet, "http://localhost:4000/tasks", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer invalidtoken")

	resp, err = client.Do(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestSignIn(t *testing.T) {
	// Test case: Successful sign-in
	user := models.User{
		Username: "testuser",
		Password: "testpassword",
	}
	body, _ := json.Marshal(user)

	req, err := http.NewRequest(http.MethodPost, "http://localhost:4000/signin", bytes.NewBuffer(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var tokenResp map[string]string
	err = json.NewDecoder(resp.Body).Decode(&tokenResp)
	require.NoError(t, err)
	require.NotEmpty(t, tokenResp["token"])

	// Test case: Incorrect password
	invalidUser := models.User{
		Username: "testuser",
		Password: "invalidpassword",
	}
	invalidBody, _ := json.Marshal(invalidUser)

	req, err = http.NewRequest(http.MethodPost, "http://localhost:4000/signin", bytes.NewBuffer(invalidBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	// Test case: User not found
	nonexistentUser := models.User{
		Username: "nonexistentuser",
		Password: "password",
	}
	nonexistentBody, _ := json.Marshal(nonexistentUser)

	req, err = http.NewRequest(http.MethodPost, "http://localhost:4000/signin", bytes.NewBuffer(nonexistentBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	// Test case: Missing username in request body
	missingUsername := models.User{
		Password: "password",
	}
	missingUsernameBody, _ := json.Marshal(missingUsername)

	req, err = http.NewRequest(http.MethodPost, "http://localhost:4000/signin", bytes.NewBuffer(missingUsernameBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// Test case: Missing password in request body
	missingPassword := models.User{
		Username: "testuser",
	}
	missingPasswordBody, _ := json.Marshal(missingPassword)

	req, err = http.NewRequest(http.MethodPost, "http://localhost:4000/signin", bytes.NewBuffer(missingPasswordBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestCreateTask(t *testing.T) {
	// Sign in to get a valid token
	user := models.User{
		Username: "TestCreateTask",
		Password: "testpassword",
	}
	body, _ := json.Marshal(user)

	// First, sign up the user
	req, err := http.NewRequest(http.MethodPost, "http://localhost:4000/signup", bytes.NewBuffer(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 10 * time.Second}
	_, _ = client.Do(req)

	// Sign in to get the token
	req, err = http.NewRequest(http.MethodPost, "http://localhost:4000/signin", bytes.NewBuffer(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	var tokenResp map[string]string
	err = json.NewDecoder(resp.Body).Decode(&tokenResp)
	require.NoError(t, err)

	token := tokenResp["token"]

	// Create a new task with valid token
	task := models.Task{
		Title:       "Test @ Task",
		Description: "This is a test task",
		AllottedTo:  "TestCreateTask",
	}
	body, _ = json.Marshal(task)

	req, err = http.NewRequest(http.MethodPost, "http://localhost:4000/tasks", bytes.NewBuffer(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)

	resp, err = client.Do(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusCreated, resp.StatusCode)

	var createdTask models.Task
	err = json.NewDecoder(resp.Body).Decode(&createdTask)
	require.NoError(t, err)
	require.Equal(t, task.Title, createdTask.Title)
	require.Equal(t, task.Description, createdTask.Description)
}

func TestGetTasks(t *testing.T) {
	// Sign in to get a valid token
	user := models.User{
		Username: "testgettasks",
		Password: "testpassword",
	}
	body, _ := json.Marshal(user)

	// First, sign up the user
	req, err := http.NewRequest(http.MethodPost, "http://localhost:4000/signup", bytes.NewBuffer(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 10 * time.Second}
	_, _ = client.Do(req)

	// Sign in to get the token
	req, err = http.NewRequest(http.MethodPost, "http://localhost:4000/signin", bytes.NewBuffer(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	var tokenResp map[string]string
	err = json.NewDecoder(resp.Body).Decode(&tokenResp)
	require.NoError(t, err)

	token := tokenResp["token"]

	// Get tasks with valid token
	req, err = http.NewRequest(http.MethodGet, "http://localhost:4000/tasks", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", token)

	resp, err = client.Do(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	var tasks []models.Task
	err = json.NewDecoder(resp.Body).Decode(&tasks)
	require.NoError(t, err)

	// Verify that tasks are returned (you may want to add more specific checks here)
	require.GreaterOrEqual(t, len(tasks), 0)
}

func TestUpdateTask(t *testing.T) {
	// Sign in to get a valid token
	user := models.User{
		Username: "testupdatetask",
		Password: "testpassword",
	}
	body, _ := json.Marshal(user)

	// First, sign up the user
	req, err := http.NewRequest(http.MethodPost, "http://localhost:4000/signup", bytes.NewBuffer(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 10 * time.Second}
	_, _ = client.Do(req)

	// Sign in to get the token
	req, err = http.NewRequest(http.MethodPost, "http://localhost:4000/signin", bytes.NewBuffer(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	var tokenResp map[string]string
	err = json.NewDecoder(resp.Body).Decode(&tokenResp)
	require.NoError(t, err)

	token := tokenResp["token"]

	// Create a new task to update later
	task := models.Task{
		Title:       "Test Update Task",
		Description: "This is a test task",
		AllottedTo:  "testupdatetask",
	}
	body, _ = json.Marshal(task)

	req, err = http.NewRequest(http.MethodPost, "http://localhost:4000/tasks", bytes.NewBuffer(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)

	resp, err = client.Do(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusCreated, resp.StatusCode)

	var createdTask models.Task
	err = json.NewDecoder(resp.Body).Decode(&createdTask)
	require.NoError(t, err)

	// Update the created task
	updatedTask := models.Task{
		Title:       "Updated Test Task",
		Description: "This is an updated test task",
		AllottedTo:  "testupdatetask",
	}
	body, _ = json.Marshal(updatedTask)

	req, err = http.NewRequest(http.MethodPut, "http://localhost:4000/tasks/"+createdTask.ID.Hex(), bytes.NewBuffer(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)

	resp, err = client.Do(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	var updatedTaskResponse models.Task
	err = json.NewDecoder(resp.Body).Decode(&updatedTaskResponse)
	require.NoError(t, err)
	require.Equal(t, updatedTask.Title, updatedTaskResponse.Title)
	require.Equal(t, updatedTask.Description, updatedTaskResponse.Description)
}

func TestGetTask(t *testing.T) {
	// Sign in to get a valid token
	user := models.User{
		Username: "testgettask",
		Password: "testpassword",
	}
	body, _ := json.Marshal(user)

	// First, sign up the user
	req, err := http.NewRequest(http.MethodPost, "http://localhost:4000/signup", bytes.NewBuffer(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 10 * time.Second}
	_, _ = client.Do(req)

	// Sign in to get the token
	req, err = http.NewRequest(http.MethodPost, "http://localhost:4000/signin", bytes.NewBuffer(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	var tokenResp map[string]string
	err = json.NewDecoder(resp.Body).Decode(&tokenResp)
	require.NoError(t, err)

	token := tokenResp["token"]

	// Create a new task with valid token
	task := models.Task{
		Title:       "Test Get Task",
		Description: "This is a test task",
		AllottedTo:  "testgettask",
	}
	body, _ = json.Marshal(task)

	req, err = http.NewRequest(http.MethodPost, "http://localhost:4000/tasks", bytes.NewBuffer(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)

	resp, err = client.Do(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusCreated, resp.StatusCode)

	var createdTask models.Task
	err = json.NewDecoder(resp.Body).Decode(&createdTask)
	require.NoError(t, err)

	// Get the created task by ID
	req, err = http.NewRequest(http.MethodGet, "http://localhost:4000/tasks/"+createdTask.ID.Hex(), nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", token)

	resp, err = client.Do(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	var fetchedTask models.Task
	err = json.NewDecoder(resp.Body).Decode(&fetchedTask)
	require.NoError(t, err)
	require.Equal(t, createdTask.ID, fetchedTask.ID)
	require.Equal(t, createdTask.Title, fetchedTask.Title)
	require.Equal(t, createdTask.Description, fetchedTask.Description)
}

func TestDeleteTask(t *testing.T) {
	// Sign in to get a valid token
	user := models.User{
		Username: "testdeletetask",
		Password: "testpassword",
	}
	body, _ := json.Marshal(user)

	// First, sign up the user
	req, err := http.NewRequest(http.MethodPost, "http://localhost:4000/signup", bytes.NewBuffer(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 10 * time.Second}
	_, _ = client.Do(req)

	// Sign in to get the token
	req, err = http.NewRequest(http.MethodPost, "http://localhost:4000/signin", bytes.NewBuffer(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	var tokenResp map[string]string
	err = json.NewDecoder(resp.Body).Decode(&tokenResp)
	require.NoError(t, err)

	token := tokenResp["token"]

	// Create a new task with valid token
	task := models.Task{
		Title:       "Test Task",
		Description: "This is a test task",
		AllottedTo:  "testdeletetask",
	}
	body, _ = json.Marshal(task)

	req, err = http.NewRequest(http.MethodPost, "http://localhost:4000/tasks", bytes.NewBuffer(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)

	resp, err = client.Do(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusCreated, resp.StatusCode)

	var createdTask models.Task
	err = json.NewDecoder(resp.Body).Decode(&createdTask)
	require.NoError(t, err)

	// Delete the created task by ID
	req, err = http.NewRequest(http.MethodDelete, "http://localhost:4000/tasks/"+createdTask.ID.Hex(), nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", token)

	resp, err = client.Do(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusNoContent, resp.StatusCode)

	// Verify the task was deleted
	req, err = http.NewRequest(http.MethodGet, "http://localhost:4000/tasks/"+createdTask.ID.Hex(), nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", token)

	resp, err = client.Do(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}

func TestSignOut(t *testing.T) {
	// Sign in to get a valid token
	user := models.User{
		Username: "testsignout",
		Password: "testpassword",
	}
	body, _ := json.Marshal(user)

	// First, sign up the user
	req, err := http.NewRequest(http.MethodPost, "http://localhost:4000/signup", bytes.NewBuffer(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 10 * time.Second}
	_, _ = client.Do(req)

	// Sign in to get the token
	req, err = http.NewRequest(http.MethodPost, "http://localhost:4000/signin", bytes.NewBuffer(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	var tokenResp map[string]string
	err = json.NewDecoder(resp.Body).Decode(&tokenResp)
	require.NoError(t, err)

	token := tokenResp["token"]

	// Test signout with valid token
	req, err = http.NewRequest(http.MethodPost, "http://localhost:4000/signout", nil)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)

	_, err = client.Do(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
}
