// users.go
// Author: Bipin Kumar Ojha (Freelancer)

package handlers

import (
	"context"
	"time"

	"github.com/bkojha74/task-management/database"
	"github.com/bkojha74/task-management/models"
	"github.com/bkojha74/task-management/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// SignUp handles user registration. It parses the user information from the request body,
// checks if the username already exists, hashes the password, and stores the user in the database.
//
// Parameters:
// - c: Fiber context, which provides methods to interact with the request and response.
//
// Returns:
// - error: An error object if an error occurs during the process.
func SignUp(c *fiber.Ctx) error {
	var user models.User
	if err := c.BodyParser(&user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse JSON"})
	}

	var existingUser models.User
	err := database.UsersCollection.FindOne(context.Background(), bson.M{"username": user.Username}).Decode(&existingUser)
	if err != nil && err != mongo.ErrNoDocuments {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	if existingUser.Username != "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "username already taken"})
	}

	user.Password = utils.HashPassword(user.Password)

	result, err := database.UsersCollection.InsertOne(context.Background(), user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not create user"})
	}

	user.ID = result.InsertedID.(primitive.ObjectID)
	return c.Status(fiber.StatusCreated).JSON(user)
}

// SignIn handles user authentication. It verifies the username and password,
// generates a JWT token if the credentials are valid, and returns the token in the response.
//
// Parameters:
// - jwtSecret: The secret key used to sign the JWT token.
// - tokenExpiryTime: The token's expiration time in seconds.
//
// Returns:
// - fiber.Handler: A Fiber handler function that performs the sign-in process.
func SignIn(jwtSecret string, tokenExpiryTime int) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var user models.User
		if err := c.BodyParser(&user); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse JSON"})
		}

		if user.Username == "" || user.Password == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "username and password should not be blank!"})
		}

		var foundUser models.User
		err := database.UsersCollection.FindOne(context.Background(), bson.M{"username": user.Username}).Decode(&foundUser)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid credentials"})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
		}

		if !utils.CheckPasswordHash(user.Password, foundUser.Password) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid credentials"})
		}

		claims := jwt.MapClaims{
			"userId": foundUser.ID.Hex(),
			"exp":    time.Now().Add(time.Second * time.Duration(tokenExpiryTime)).Unix(),
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

		tokenString, err := token.SignedString([]byte(jwtSecret))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not generate token"})
		}

		return c.JSON(fiber.Map{"token": tokenString})
	}
}

// SignOut handles user sign-out. It returns a simple success message.
//
// Parameters:
// - c: Fiber context, which provides methods to interact with the request and response.
//
// Returns:
// - error: An error object if an error occurs during the process.
func SignOut(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "signed out"})
}
