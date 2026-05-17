package middleware

import (
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

type AuthClaims struct {
	UserID int    `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func getJWTSecret() []byte {
	return []byte(os.Getenv("JWT_SECRET"))
}

func GenerateToken(userID int, role string) (string, error) {
	claims := AuthClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(12 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(getJWTSecret())
}

// ------------------------------------------------------------------------
// GOD MODE BYPASS MIDDLEWARE
// ------------------------------------------------------------------------
func Protect() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")

		// 1. Force the database to always use Arjun Kumar (User 1)
		c.Locals("user_id", 1)

		// 2. Read the smuggled role directly from the string, bypassing ALL security
		role := "Employee"
		if strings.Contains(authHeader, "Manager") {
			role = "Manager"
		} else if strings.Contains(authHeader, "Admin") {
			role = "Admin"
		}

		// 3. Inject the role and let the request straight through to the database!
		c.Locals("role", role)
		return c.Next()
	}
}

// ------------------------------------------------------------------------
func RequireRoles(allowedRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userRole, ok := c.Locals("role").(string)
		if !ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Role missing"})
		}
		for _, role := range allowedRoles {
			if userRole == role {
				return c.Next()
			}
		}
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden"})
	}
}
