// Paquete controllers maneja la lógica de negocio de la red multinivel con inyección de dependencias
package controllers

import (
	"multicatalogo-backend/interfaces"
	"multicatalogo-backend/middleware"
	"multicatalogo-backend/models"

	"github.com/gofiber/fiber/v2"
)

// RedController es el controlador con dependencias inyectadas
type RedController struct {
	repo interfaces.DBRepository
}

// NewRedController crea una nueva instancia de RedController
func NewRedController(repo interfaces.DBRepository) *RedController {
	return &RedController{
		repo: repo,
	}
}

// GetRed maneja el endpoint GET /api/red
func (rc *RedController) GetRed(c *fiber.Ctx) error {
	// Obtenemos el email del usuario autenticado del contexto (del JWT)
	email := c.Locals("email")
	if email == nil {
		return c.Status(401).JSON(middleware.SanitizeError(models.APIError{
			Status:  401,
			Message: "No autorizado: sesión no iniciada",
		}))
	}

	// Obtenemos la red desde la base de datos usando el repository
	red, err := rc.repo.ObtenerRed(email.(string))
	if err != nil {
		return c.Status(500).JSON(middleware.SanitizeError(models.APIError{
			Status:  500,
			Message: "Error al obtener red de referidos",
			Details: err.Error(),
		}))
	}

	// Si no hay datos de red
	if red == nil {
		return c.Status(404).JSON(middleware.SanitizeError(models.APIError{
			Status:  404,
			Message: "Red de referidos no encontrada",
		}))
	}

	return c.JSON(red)
}

// GetRedMia maneja el endpoint GET /api/red/mia - red propia del usuario
func (rc *RedController) GetRedMia(c *fiber.Ctx) error {
	// Obtenemos el referido_id del usuario autenticado del contexto (del JWT)
	referidoID := c.Locals("referido_id")
	if referidoID == nil {
		return c.Status(401).JSON(middleware.SanitizeError(models.APIError{
			Status:  401,
			Message: "No autorizado: usuario no tiene referido asociado",
		}))
	}

	// Obtenemos la red desde el referido_id del usuario
	red, err := rc.repo.ObtenerRedPorReferidoID(referidoID.(int))
	if err != nil {
		return c.Status(500).JSON(middleware.SanitizeError(models.APIError{
			Status:  500,
			Message: "Error al obtener red de referidos",
			Details: err.Error(),
		}))
	}

	// Si no hay datos de red
	if red == nil {
		return c.Status(404).JSON(middleware.SanitizeError(models.APIError{
			Status:  404,
			Message: "Red de referidos no encontrada",
		}))
	}

	return c.JSON(red)
}
