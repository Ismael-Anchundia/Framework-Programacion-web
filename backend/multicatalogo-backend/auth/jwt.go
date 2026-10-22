// Paquete auth maneja la autenticación y autorización con JWT
package auth

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims representa los claims del JWT
type Claims struct {
	Email      string `json:"email"`
	Rol        string `json:"rol"`
	Sub        int    `json:"sub"` // ID del usuario
	ReferidoID *int   `json:"referido_id,omitempty"`
	jwt.RegisteredClaims
}

// GenerateToken genera un token JWT para un usuario
func GenerateToken(userID int, email, rol string, referidoID *int) (string, error) {
	// Obtener JWT_SECRET del entorno
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "multicatalogo-secret-key-2026-very-secure" // Valor por defecto
	}

	// Obtener tiempo de expiración
	expirationHours := os.Getenv("JWT_EXPIRATION_HOURS")
	if expirationHours == "" {
		expirationHours = "2" // 2 horas por defecto
	}

	hours, err := time.ParseDuration(expirationHours + "h")
	if err != nil || hours <= 0 {
		hours = 2 * time.Hour
	}
	expirationTime := time.Now().Add(hours)
	claims := &Claims{
		Email:      email,
		Rol:        rol,
		Sub:        userID,
		ReferidoID: referidoID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "multicatalogo-backend",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ValidateToken valida un token JWT y retorna los claims
func ValidateToken(tokenString string) (*Claims, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "multicatalogo-secret-key-2026-very-secure"
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("token inválido")
}

// ExtractToken extrae el token del header Authorization
func ExtractToken(authHeader string) (string, error) {
	if authHeader == "" {
		return "", errors.New("header Authorization faltante")
	}

	// El formato debe ser: "Bearer <token>"
	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		return "", errors.New("formato de Authorization inválido")
	}

	return authHeader[7:], nil
}
