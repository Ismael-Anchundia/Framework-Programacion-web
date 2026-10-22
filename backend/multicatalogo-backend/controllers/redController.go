// Declaramos el paquete controllers para agrupar las funciones que manejan la lógica de negocio de las rutas.
package controllers

import (
	// Importamos el framework Fiber para tener acceso al contexto (c *fiber.Ctx) de la petición HTTP.
	"github.com/gofiber/fiber/v2"
	// Importamos nuestro paquete de modelos para usar la estructura APIError.
	"multicatalogo-backend/models"
	// Importamos nuestro paquete repository para acceder a los datos
	"multicatalogo-backend/repository"
)

// GetRed es la función controladora que devuelve la red multinivel de referidos
func GetRed(c *fiber.Ctx) error {
	// Obtenemos el email del usuario autenticado del contexto o de query params
	// Por ahora usamos un email por defecto para el usuario "demo"
	email := c.Query("email")
	if email == "" {
		email = "demo@multicatalogo.com" // Email por defecto para demostración
	}

	// Obtenemos la red desde la base de datos usando el repository
	red, err := repository.ObtenerRed(email)
	if err != nil {
		return c.Status(500).JSON(models.APIError{
			Status:  500,
			Message: "Error al obtener red de referidos",
			Details: err.Error(),
		})
	}

	// Si no hay datos de red
	if red == nil {
		return c.Status(404).JSON(models.APIError{
			Status:  404,
			Message: "Red de referidos no encontrada",
		})
	}

	return c.JSON(red)
}
