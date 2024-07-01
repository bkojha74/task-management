// database_test.go
// Author: Bipin Kumar Ojha (Freelancer)

package database

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/bkojha74/task-management/helper"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TestMain is the entry point for testing. It sets up the MongoDB connection,
// runs the tests, and then disconnects from MongoDB.
func TestMain(m *testing.M) {
	// Get the current working directory
	currentWorkDirectory, err := os.Getwd()
	if err != nil {
		log.Fatal(err.Error())
	}

	// Load environment variables from the configuration file
	helper.LoadEnv(currentWorkDirectory + "/../config")

	// Initialize MongoDB connection using the URI from environment variables
	Init(helper.GetEnv("MONGO_URI"))

	// Run the tests
	code := m.Run()

	// Disconnect from MongoDB and clean up resources
	Disconnect()

	// Exit with the code from the test run
	os.Exit(code)
}

// TestInsertTask tests the insertion of a task into the MongoDB collection
func TestInsertTask(t *testing.T) {
	// Create a task document
	task := bson.M{
		"userId":      primitive.NewObjectID(),
		"title":       "Test Task",
		"description": "This is an insert test task",
		"allotted_to": "bipin",
		"done_by":     "",
		"status":      "Pending",
		"start_time":  primitive.NewDateTimeFromTime(time.Now()),
		"end_time":    primitive.NewDateTimeFromTime(time.Now().AddDate(0, 0, 7)),
	}

	// Insert the task into the Tasks collection
	result, err := TasksCollection.InsertOne(context.Background(), task)
	assert.NoError(t, err)              // Assert that there is no error
	assert.NotNil(t, result.InsertedID) // Assert that the inserted ID is not nil
}

// TestFindTasks tests finding a task in the MongoDB collection
func TestFindTasks(t *testing.T) {
	// Create a task document
	task := bson.M{
		"userId":      primitive.NewObjectID(),
		"title":       "Find Test Task",
		"description": "This is a find test task",
		"allotted_to": "bipin",
		"done_by":     "",
		"status":      "Pending",
		"start_time":  primitive.NewDateTimeFromTime(time.Now()),
		"end_time":    primitive.NewDateTimeFromTime(time.Now().AddDate(0, 0, 7)),
	}

	// Insert the task into the Tasks collection
	insertResult, err := TasksCollection.InsertOne(context.Background(), task)
	assert.NoError(t, err) // Assert that there is no error

	// Find the inserted task by ID
	var foundTask bson.M
	err = TasksCollection.FindOne(context.Background(), bson.M{"_id": insertResult.InsertedID}).Decode(&foundTask)
	assert.NoError(t, err)                             // Assert that there is no error
	assert.Equal(t, task["title"], foundTask["title"]) // Assert that the titles match
}

// TestFindTaskByID tests finding a task by its ID in the MongoDB collection
func TestFindTaskByID(t *testing.T) {
	// Create a task document
	task := bson.M{
		"userId":      primitive.NewObjectID(),
		"title":       "Find Test Task",
		"description": "This is a find test task",
		"allotted_to": "bipin",
		"done_by":     "",
		"status":      "Pending",
		"start_time":  primitive.NewDateTimeFromTime(time.Now()),
		"end_time":    primitive.NewDateTimeFromTime(time.Now().AddDate(0, 0, 7)),
	}

	// Insert the task into the Tasks collection
	insertResult, err := TasksCollection.InsertOne(context.Background(), task)
	assert.NoError(t, err) // Assert that there is no error

	// Find the inserted task by ID
	var foundTask bson.M
	err = TasksCollection.FindOne(context.Background(), bson.M{"_id": insertResult.InsertedID}).Decode(&foundTask)
	assert.NoError(t, err)                             // Assert that there is no error
	assert.Equal(t, task["title"], foundTask["title"]) // Assert that the titles match
}

// TestFindAllTasks tests finding all tasks in the MongoDB collection for a specific user
func TestFindAllTasks(t *testing.T) {
	// Create a user ID to associate with multiple tasks
	userID := primitive.NewObjectID()
	// Create multiple task documents
	tasks := []interface{}{
		bson.M{
			"userId":      userID,
			"title":       "Find All Task 1",
			"description": "This is the first find all test task",
			"allotted_to": "bipin0",
			"done_by":     "",
			"status":      "Pending",
			"start_time":  primitive.NewDateTimeFromTime(time.Now()),
			"end_time":    primitive.NewDateTimeFromTime(time.Now().AddDate(0, 0, 7)),
		},
		bson.M{
			"userId":      userID,
			"title":       "Find All Task 2",
			"description": "This is the second find all test task",
			"allotted_to": "bipin0",
			"done_by":     "",
			"status":      "Pending",
			"start_time":  primitive.NewDateTimeFromTime(time.Now()),
			"end_time":    primitive.NewDateTimeFromTime(time.Now().AddDate(0, 0, 7)),
		},
	}

	// Insert multiple tasks into the Tasks collection
	_, err := TasksCollection.InsertMany(context.Background(), tasks)
	assert.NoError(t, err) // Assert that there is no error

	// Find all tasks for the specified user ID
	cursor, err := TasksCollection.Find(context.Background(), bson.M{"userId": userID})
	assert.NoError(t, err) // Assert that there is no error

	// Decode the found tasks into a slice
	var foundTasks []bson.M
	err = cursor.All(context.Background(), &foundTasks)
	assert.NoError(t, err)              // Assert that there is no error
	assert.Equal(t, 2, len(foundTasks)) // Assert that two tasks were found
}

// TestUpdateTaskByID tests updating a task by its ID in the MongoDB collection
func TestUpdateTaskByID(t *testing.T) {
	// Create a task document
	task := bson.M{
		"userId":      primitive.NewObjectID(),
		"title":       "Update Test Task",
		"description": "This is an update test task",
		"allotted_to": "bipin1",
		"done_by":     "",
		"status":      "Pending",
		"start_time":  primitive.NewDateTimeFromTime(time.Now()),
		"end_time":    primitive.NewDateTimeFromTime(time.Now().AddDate(0, 0, 7)),
	}

	// Insert the task into the Tasks collection
	insertResult, err := TasksCollection.InsertOne(context.Background(), task)
	assert.NoError(t, err) // Assert that there is no error

	// Update the title of the inserted task
	updatedTitle := "Updated Task Title"
	updateResult, err := TasksCollection.UpdateOne(context.Background(), bson.M{"_id": insertResult.InsertedID}, bson.M{"$set": bson.M{"title": updatedTitle}})
	assert.NoError(t, err)                                // Assert that there is no error
	assert.Equal(t, int64(1), updateResult.ModifiedCount) // Assert that one document was modified

	// Find the updated task by ID
	var updatedTask bson.M
	err = TasksCollection.FindOne(context.Background(), bson.M{"_id": insertResult.InsertedID}).Decode(&updatedTask)
	assert.NoError(t, err)                              // Assert that there is no error
	assert.Equal(t, updatedTitle, updatedTask["title"]) // Assert that the title was updated correctly
}

// TestDeleteTaskByID tests deleting a task by its ID in the MongoDB collection
func TestDeleteTaskByID(t *testing.T) {
	// Create a task document
	task := bson.M{
		"userId":      primitive.NewObjectID(),
		"title":       "Delete Test Task",
		"description": "This is a delete test task",
		"allotted_to": "bipin",
		"done_by":     "",
		"status":      "Pending",
		"start_time":  primitive.NewDateTimeFromTime(time.Now()),
		"end_time":    primitive.NewDateTimeFromTime(time.Now().AddDate(0, 0, 7)),
	}

	// Insert the task into the Tasks collection
	insertResult, err := TasksCollection.InsertOne(context.Background(), task)
	assert.NoError(t, err) // Assert that there is no error

	// Delete the inserted task by ID
	deleteResult, err := TasksCollection.DeleteOne(context.Background(), bson.M{"_id": insertResult.InsertedID})
	assert.NoError(t, err)                               // Assert that there is no error
	assert.Equal(t, int64(1), deleteResult.DeletedCount) // Assert that one document was deleted
}
