// Paquete controllers maneja la lógica de negocio de registro
package controllers

import (
	"multicatalogo-backend/interfaces"
	"multicatalogo-backend/middleware"
	"multicatalogo-backend/models"

	"github.com/gofiber/fiber/v2"
)

// RegisterController es el controlador con dependencias inyectadas
type RegisterController struct {
	repo interfaces.DBRepository
}

// NewRegisterController crea una nueva instancia de RegisterController
func NewRegisterController(repo interfaces.DBRepository) *RegisterController {
	return &RegisterController{
		repo: repo,
	}
}

// RegisterRequest define la estructura para registro
type RegisterRequest struct {
	Nombre   string `json:"nombre"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Ref      string `json:"ref,omitempty"` // Código de referido opcional
}

// Register maneja el endpoint POST /api/register
func (rc *RegisterController) Register(c *fiber.Ctx) error {
	var req RegisterRequest

	// Parsear el cuerpo de la petición
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(middleware.SanitizeError(models.APIError{
			Status:  400,
			Message: "Error al parsear la petición",
			Details: err.Error(),
		}))
	}

	// Validaciones básicas
	if req.Nombre == "" {
		return c.Status(400).JSON(middleware.SanitizeError(models.APIError{
			Status:  400,
			Message: "El nombre es requerido",
		}))
	}

	if req.Email == "" {
		return c.Status(400).JSON(middleware.SanitizeError(models.APIError{
			Status:  400,
			Message: "El email es requerido",
		}))
	}

	if req.Password == "" {
		return c.Status(400).JSON(middleware.SanitizeError(models.APIError{
			Status:  400,
			Message: "El password es requerido",
		}))
	}

	// Verificar si el código de referido existe (si se proporciona)
	var referidoID *int
	if req.Ref != "" {
		referido, err := rc.repo.ObtenerReferidoPorCodigo(req.Ref)
		if err != nil {
			return c.Status(500).JSON(middleware.SanitizeError(models.APIError{
				Status:  500,
				Message: "Error al verificar código de referido",
				Details: err.Error(),
			}))
		}
		if referido == nil {
			return c.Status(400).JSON(middleware.SanitizeError(models.APIError{
				Status:  400,
				Message: "Código de referido inválido",
			}))
		}
		referidoID = &referido.ID
	}

	// Crear usuario con rol cliente por defecto
	usuario := &interfaces.UsuarioDB{
		Email:      req.Email,
		Password:   req.Password, // El repository lo hasheará con bcrypt
		Rol:        "cliente",
		Nombre:     req.Nombre,
		Activo:     true,
		ReferidoID: referidoID,
	}

	err := rc.repo.CrearUsuario(usuario, referidoID)
	if err != nil {
		// Verificar si es error de email duplicado
		if err.Error() == "email ya existe" {
			return c.Status(409).JSON(middleware.SanitizeError(models.APIError{
				Status:  409,
				Message: "El email ya está registrado",
			}))
		}
		return c.Status(500).JSON(middleware.SanitizeError(models.APIError{
			Status:  500,
			Message: "Error al crear usuario",
			Details: err.Error(),
		}))
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "Usuario creado exitosamente",
		"email":   req.Email,
	})
}
