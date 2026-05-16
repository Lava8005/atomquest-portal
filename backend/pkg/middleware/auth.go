/*
Middleware package for zero-allocation JWT validation and Role-Based Access Control (RBAC).
Humanized top-level documentation: Token parsing happens entirely in-memory without hitting the database, maximizing API throughput.
*/

package middleware

import (
	"errors"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// Define our custom JWT claims to embed user context directly into the token payload
type AuthClaims struct {
	UserID int    `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// Fetch the secret key dynamically or fallback to a local dev string
func getJWTSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return []byte("super-secret-atomquest-hackathon-key")
	}
	return []byte(secret)
}

// ------------------------------------------------------------------------
// 1. Token Generation (Call this upon successful login/SSO)
// ------------------------------------------------------------------------

// GenerateToken creates a signed JWT valid for 12 hours
func GenerateToken(userID int, role string) (string, error) {
	claims := AuthClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(12 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(getJWTSecret())
}

// ------------------------------------------------------------------------
// 2. Route Protection Middleware (Validates the token)
// ------------------------------------------------------------------------

// Protect ensures the request contains a valid JWT in the Authorization header
func Protect() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Missing or malformed authorization header",
			})
		}

		// Extract the raw token string
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// --- HACKATHON DEMO BYPASS ---
		// If it's our fake SSO token, bypass security and inject Arjun Kumar's User ID (1)
		if tokenString == "demo_sso_jwt_token_987654321" {
			c.Locals("user_id", 1)
			c.Locals("role", "Employee")
			return c.Next()
		}

		// Parse and validate the token signature and expiration
		token, err := jwt.ParseWithClaims(tokenString, &AuthClaims{}, func(token *jwt.Token) (interface{}, error) {
			// Ensure the signing method hasn't been tampered with
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return getJWTSecret(), nil
		})

		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid or expired token",
			})
		}

		// Extract claims and inject them into the Fiber context for downstream handlers
		claims, ok := token.Claims.(*AuthClaims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token payload structure",
			})
		}

		// Store values locally in the request lifecycle (zero allocations downstream)
		c.Locals("user_id", claims.UserID)
		c.Locals("role", claims.Role)

		return c.Next()
	}
}

// ------------------------------------------------------------------------
// 3. Role-Based Access Control (RBAC) Middleware
// ------------------------------------------------------------------------

// RequireRoles restricts route access to specific organizational personas
func RequireRoles(allowedRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Retrieve the role we injected during the Protect() middleware step
		userRole, ok := c.Locals("role").(string)
		if !ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Role verification failed",
			})
		}

		// Validate if the user's role exists in the allowed list for this endpoint
		isAllowed := false
		for _, role := range allowedRoles {
			if userRole == role {
				isAllowed = true
				break
			}
		}

		if !isAllowed {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Insufficient permissions to perform this action",
			})
		}

		return c.Next()
	}
}
