// Paquete controllers maneja la lógica de negocio de autenticación con inyección de dependencias
package controllers

import (
	"multicatalogo-backend/auth"
	"multicatalogo-backend/interfaces"
	"multicatalogo-backend/middleware"
	"multicatalogo-backend/models"

	"github.com/gofiber/fiber/v2"
)

// AuthController es el controlador con dependencias inyectadas
type AuthController struct {
	repo interfaces.DBRepository
}

// NewAuthController crea una nueva instancia de AuthController
func NewAuthController(repo interfaces.DBRepository) *AuthController {
	return &AuthController{
		repo: repo,
	}
}

// Login maneja el endpoint POST /api/login
func (ac *AuthController) Login(c *fiber.Ctx) error {
	var req models.LoginRequest

	// Parsear el cuerpo de la petición
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(middleware.SanitizeError(models.APIError{
			Status:  400,
			Message: "Error al parsear la petición",
			Details: err.Error(),
		}))
	}

	// Validación explícita de campos vacíos
	if req.Email == "" {
		return c.Status(400).JSON(middleware.SanitizeError(models.APIError{
			Status:  400,
			Message: "El campo email es requerido",
		}))
	}

	if req.Password == "" {
		return c.Status(400).JSON(middleware.SanitizeError(models.APIError{
			Status:  400,
			Message: "El campo password es requerido",
		}))
	}

	// Buscar usuario en el repositorio (verifica password con bcrypt)
	usuario, err := ac.repo.BuscarUsuarioPorCredenciales(req.Email, req.Password)
	if err != nil {
		if err.Error() == "cuenta desactivada" {
			return c.Status(403).JSON(middleware.SanitizeError(models.APIError{
				Status:  403,
				Message: "Cuenta desactivada",
			}))
		}
		return c.Status(500).JSON(middleware.SanitizeError(models.APIError{
			Status:  500,
			Message: "Error al buscar usuario",
			Details: err.Error(),
		}))
	}

	if usuario == nil {
		return c.Status(401).JSON(middleware.SanitizeError(models.APIError{
			Status:  401,
			Message: "Credenciales inválidas",
		}))
	}

	// Generar token JWT real
	token, err := auth.GenerateToken(usuario.ID, usuario.Email, usuario.Rol, usuario.ReferidoID)
	if err != nil {
		return c.Status(500).JSON(middleware.SanitizeError(models.APIError{
			Status:  500,
			Message: "Error al generar token",
			Details: err.Error(),
		}))
	}

	// Retornar respuesta con email, rol y token JWT real
	return c.JSON(fiber.Map{
		"token":       token,
		"email":       usuario.Email,
		"rol":         usuario.Rol,
		"referido_id": usuario.ReferidoID,
	})
}
