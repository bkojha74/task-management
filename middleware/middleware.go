// middleware.go
// Author: Bipin Kumar Ojha (Freelancer)

package middleware

import (
	"github.com/gofiber/fiber/v2"
	jwtware "github.com/gofiber/jwt/v3"
	"github.com/golang-jwt/jwt/v4"
)

// Protected creates a middleware handler that protects routes using JWT authentication.
// It checks the Authorization header for a valid JWT token, extracts the user ID from the token,
// and sets it in the request context for further use in the application.
// If the token is invalid or not present, it returns a 401 Unauthorized response.
//
// Parameters:
// - jwtSecret: The secret key used to sign the JWT token.
//
// Returns:
// - fiber.Handler: The Fiber middleware handler for JWT authentication.
func Protected(jwtSecret string) fiber.Handler {
	return jwtware.New(jwtware.Config{
		SigningKey:  []byte(jwtSecret),      // The secret key for signing the JWT token.
		ContextKey:  "user",                 // The key used to store the JWT token in the request context.
		TokenLookup: "header:Authorization", // The location of the JWT token in the request (Authorization header).

		// SuccessHandler is called when the JWT token is successfully validated.
		// It extracts the user ID from the token and sets it in the request context.
		SuccessHandler: func(c *fiber.Ctx) error {
			user := c.Locals("user").(*jwt.Token)         // Retrieve the JWT token from the request context.
			claims := user.Claims.(jwt.MapClaims)         // Extract the claims from the token.
			c.Locals("userId", claims["userId"].(string)) // Set the user ID in the request context.
			return c.Next()                               // Proceed to the next middleware/handler.
		},

		// ErrorHandler is called when the JWT token is invalid or not present.
		// It returns a 401 Unauthorized response with an error message.
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			if err != nil {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
			}
			return c.Next() // Proceed to the next middleware/handler.
		},
	})
}
