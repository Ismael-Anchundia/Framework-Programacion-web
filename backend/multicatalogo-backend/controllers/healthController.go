// Declaramos el paquete controllers para agrupar las funciones que manejan la lógica de negocio de las rutas.
package controllers

import (
	"multicatalogo-backend/config"

	"github.com/gofiber/fiber/v2"
)

// HealthResponse define la estructura de respuesta del health check
type HealthResponse struct {
	Status    string `json:"status"`
	Message   string `json:"message"`
	Postgres  string `json:"postgres"`
	Redis     string `json:"redis"`
}

// Health es la función controladora que verifica el estado operativo del servidor y sus dependencias.
func Health(c *fiber.Ctx) error {
	response := HealthResponse{
		Status:  "ok",
		Message: "Servidor Go/Fiber operativo",
	}

	// Verificar conexión a PostgreSQL
	if config.DB != nil {
		if err := config.DB.Ping(); err != nil {
			response.Status = "degraded"
			response.Postgres = "down"
			response.Message = "PostgreSQL no disponible"
		} else {
			response.Postgres = "up"
		}
	} else {
		response.Status = "degraded"
		response.Postgres = "not_configured"
	}

	// Verificar conexión a Redis
	if config.RedisClient != nil {
		_, err := config.RedisClient.Ping(c.Context()).Result()
		if err != nil {
			response.Status = "degraded"
			response.Redis = "down"
			response.Message = "Redis no disponible"
		} else {
			response.Redis = "up"
		}
	} else {
		response.Redis = "not_configured"
	}

	// Determinar código HTTP según estado
	statusCode := 200
	if response.Status == "degraded" {
		statusCode = 503
	}

	return c.Status(statusCode).JSON(response)
}
