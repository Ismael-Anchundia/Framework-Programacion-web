// Declaramos el paquete controllers para agrupar las funciones que manejan la lógica de negocio de las rutas.
package controllers

import (
	// Importamos el framework Fiber para tener acceso al contexto (c *fiber.Ctx) de la petición HTTP.
	"github.com/gofiber/fiber/v2"
	// Importamos nuestro paquete de modelos para poder usar la estructura LoginRequest y APIError.
	"multicatalogo-backend/models"
	// Importamos nuestro paquete repository para acceder a los datos
	"multicatalogo-backend/repository"
	// Importamos strings para validaciones de cadenas
	"strings"
)

// Login es la función controladora que se ejecutará cuando el cliente envíe sus credenciales.
func Login(c *fiber.Ctx) error {
	// Creamos una variable 'req' del tipo LoginRequest (ubicada en nuestro paquete models) para almacenar los datos.
	var req models.LoginRequest

	// Intentamos parsear (transformar) el cuerpo JSON entrante y guardarlo en la variable 'req'.
	if err := c.BodyParser(&req); err != nil {
		// Si ocurre un error al parsear (ej. JSON mal formado), retornamos un estado HTTP 400 (Bad Request) usando la estructura APIError.
		return c.Status(400).JSON(models.APIError{
			Status:  400,
			Message: "Cuerpo de petición inválido",
			Details: err.Error(),
		})
	}

	// Validación explícita: aseguramos que Email y Password no vengan vacíos
	if strings.TrimSpace(req.Email) == "" {
		return c.Status(400).JSON(models.APIError{
			Status:  400,
			Message: "El campo email es requerido",
		})
	}

	if strings.TrimSpace(req.Password) == "" {
		return c.Status(400).JSON(models.APIError{
			Status:  400,
			Message: "El campo password es requerido",
		})
	}

	// Buscamos el usuario en la base de datos usando el repository
	usuario, err := repository.BuscarUsuarioPorCredenciales(req.Email, req.Password)
	if err != nil {
		return c.Status(500).JSON(models.APIError{
			Status:  500,
			Message: "Error al buscar usuario",
			Details: err.Error(),
		})
	}

	// Si el usuario no existe o las credenciales son incorrectas
	if usuario == nil {
		return c.Status(401).JSON(models.APIError{
			Status:  401,
			Message: "Credenciales incorrectas",
		})
	}

	// Si las credenciales son correctas, retornamos el token ficticio y los datos del usuario
	return c.JSON(fiber.Map{
		"token": "fake-jwt-token-123",
		"email": usuario.Email,
		"rol":   usuario.Rol,
	})
}