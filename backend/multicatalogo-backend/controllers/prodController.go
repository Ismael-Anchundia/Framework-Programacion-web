// Seguimos dentro del paquete controllers, ya que maneja la lógica de otra sección del negocio.
package controllers

import (
	// Importamos Fiber para manejar la respuesta HTTP.
	"github.com/gofiber/fiber/v2"
	// Importamos nuestro paquete de modelos para usar la estructura Producto y APIError.
	"multicatalogo-backend/models"
	// Importamos nuestro paquete repository para acceder a los datos
	"multicatalogo-backend/repository"
	// Importamos strconv para conversión de strings a números
	"strconv"
)

// GetProductos es la función controladora encargada de devolver el catálogo de artículos.
func GetProductos(c *fiber.Ctx) error {
	// Obtenemos el parámetro de búsqueda
	busqueda := c.Query("q", "")

	// Obtenemos los productos desde la base de datos usando el repository
	productos, err := repository.ObtenerProductos(busqueda)
	if err != nil {
		return c.Status(500).JSON(models.APIError{
			Status:  500,
			Message: "Error al obtener productos",
			Details: err.Error(),
		})
	}

	// Devolvemos los productos tal como vienen de la base de datos
	// El frontend espera los campos completos
	return c.JSON(productos)
}

// GetProductoPorID es la función controladora encargada de devolver un producto específico por su ID.
func GetProductoPorID(c *fiber.Ctx) error {
	// Obtenemos el parámetro ID de la URL y lo convertimos a entero
	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)

	// Si la conversión falla (el usuario envió texto en lugar de número), devolvemos un error HTTP 400
	if err != nil {
		return c.Status(400).JSON(models.APIError{
			Status:  400,
			Message: "El parámetro ID debe ser un número válido",
			Details: idParam,
		})
	}

	// Obtenemos el producto desde la base de datos usando el repository
	producto, err := repository.ObtenerProductoPorID(id)
	if err != nil {
		return c.Status(500).JSON(models.APIError{
			Status:  500,
			Message: "Error al obtener producto",
			Details: err.Error(),
		})
	}

	// Si el producto no existe
	if producto == nil {
		return c.Status(404).JSON(models.APIError{
			Status:  404,
			Message: "Producto no encontrado",
			Details: id,
		})
	}

	return c.JSON(producto)
}
