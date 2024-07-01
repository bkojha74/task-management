// tasks.go
// Author: Bipin Kumar Ojha (Freelancer)

package handlers

import (
	"context"
	"time"

	"github.com/bkojha74/task-management/database"
	"github.com/bkojha74/task-management/models"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// CreateTask handles the creation of a new task. It validates the allotted user,
// sets the task's initial status, and inserts the task into the database.
//
// Parameters:
// - c: Fiber context, which provides methods to interact with the request and response.
//
// Returns:
// - error: An error object if an error occurs during the process.
func CreateTask(c *fiber.Ctx) error {
	userId := c.Locals("userId").(string)

	var task models.Task
	if err := c.BodyParser(&task); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}

	// Validate allottedTo field
	var user models.User
	err := database.UsersCollection.FindOne(context.Background(), bson.M{"username": task.AllottedTo}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Allotted user does not exist"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error checking allotted user"})
	}

	task.ID = primitive.NewObjectID()
	task.UserID, _ = primitive.ObjectIDFromHex(userId)
	task.StartDate = primitive.NewDateTimeFromTime(time.Now())
	task.Status = "Pending"

	_, err = database.TasksCollection.InsertOne(context.Background(), task)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not create task"})
	}

	return c.Status(fiber.StatusCreated).JSON(task)
}

// GetTasks retrieves all tasks associated with the logged-in user from the database.
//
// Parameters:
// - c: Fiber context, which provides methods to interact with the request and response.
//
// Returns:
// - error: An error object if an error occurs during the process.
func GetTasks(c *fiber.Ctx) error {
	userId := c.Locals("userId").(string)

	// Convert userId to ObjectID
	userObjectId, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	var tasks []models.Task
	filter := bson.M{"userId": userObjectId}

	cursor, err := database.TasksCollection.Find(context.Background(), filter)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "No tasks found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error fetching tasks"})
	}

	if err = cursor.All(context.Background(), &tasks); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error decoding tasks"})
	}

	return c.Status(fiber.StatusOK).JSON(tasks)
}

// GetTask retrieves a specific task by its ID and the logged-in user ID from the database.
//
// Parameters:
// - c: Fiber context, which provides methods to interact with the request and response.
//
// Returns:
// - error: An error object if an error occurs during the process.
func GetTask(c *fiber.Ctx) error {
	userId := c.Locals("userId").(string)
	taskId := c.Params("id")

	taskIdHex, err := primitive.ObjectIDFromHex(taskId)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid task ID"})
	}

	userIdHex, _ := primitive.ObjectIDFromHex(userId)
	var task models.Task
	err = database.TasksCollection.FindOne(context.Background(), bson.M{"_id": taskIdHex, "userId": userIdHex}).Decode(&task)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Task not found"})
	}

	return c.JSON(task)
}

// UpdateTask updates a specific task by its ID and the logged-in user ID in the database.
//
// Parameters:
// - c: Fiber context, which provides methods to interact with the request and response.
//
// Returns:
// - error: An error object if an error occurs during the process.
func UpdateTask(c *fiber.Ctx) error {
	userId := c.Locals("userId").(string)
	taskId := c.Params("id")

	taskIdHex, err := primitive.ObjectIDFromHex(taskId)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid task ID"})
	}

	userIdHex, _ := primitive.ObjectIDFromHex(userId)
	var task models.Task
	if err := c.BodyParser(&task); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}

	task.UserID = userIdHex
	task.ID = taskIdHex

	result, err := database.TasksCollection.UpdateOne(context.Background(), bson.M{"_id": taskIdHex, "userId": userIdHex}, bson.M{"$set": task})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not update task"})
	}

	if result.MatchedCount == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Task not found"})
	}

	return c.JSON(task)
}

// DeleteTask deletes a specific task by its ID and the logged-in user ID from the database.
//
// Parameters:
// - c: Fiber context, which provides methods to interact with the request and response.
//
// Returns:
// - error: An error object if an error occurs during the process.
func DeleteTask(c *fiber.Ctx) error {
	userId := c.Locals("userId").(string)
	taskId := c.Params("id")

	taskIdHex, err := primitive.ObjectIDFromHex(taskId)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid task ID"})
	}

	userIdHex, _ := primitive.ObjectIDFromHex(userId)

	filter := bson.M{"_id": taskIdHex, "userId": userIdHex}
	result, err := database.TasksCollection.DeleteOne(context.Background(), filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not delete task"})
	}

	if result.DeletedCount == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Task not found"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}
