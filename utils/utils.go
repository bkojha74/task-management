package utils

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) string {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(hashedPassword)
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func JWTMiddleware(secret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get the token from the Authorization header
		tokenString := c.Get("Authorization")
		if tokenString == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing or malformed JWT"})
		}

		// Parse the token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Make sure that the token method conform to "SigningMethodHMAC"
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fiber.NewError(fiber.StatusUnauthorized, "unexpected signing method")
			}
			return []byte(secret), nil
		})

		if err != nil {
			log.Printf("Error parsing JWT: %v", err)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid JWT"})
		}

		// Extract the claims and set them in the context
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			c.Locals("userId", claims["userId"])
			return c.Next()
		} else {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid JWT"})
		}
	}
}
