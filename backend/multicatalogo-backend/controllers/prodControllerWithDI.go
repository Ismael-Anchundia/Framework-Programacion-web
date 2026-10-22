// Paquete controllers maneja la lógica de negocio de productos con inyección de dependencias
package controllers

import (
	"multicatalogo-backend/interfaces"
	"multicatalogo-backend/middleware"
	"multicatalogo-backend/models"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// ProdController es el controlador con dependencias inyectadas
type ProdController struct {
	repo interfaces.DBRepository
}

// NewProdController crea una nueva instancia de ProdController
func NewProdController(repo interfaces.DBRepository) *ProdController {
	return &ProdController{
		repo: repo,
	}
}

// GetProductos maneja el endpoint GET /api/productos
func (pc *ProdController) GetProductos(c *fiber.Ctx) error {
	busqueda := c.Query("q", "")
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	productos, total, err := pc.repo.ObtenerProductosPaginados(busqueda, c.Query("categoria"), page, limit)
	if err != nil {
		return c.Status(500).JSON(middleware.SanitizeError(models.APIError{
			Status:  500,
			Message: "Error al obtener productos",
			Details: err.Error(),
		}))
	}

	return c.JSON(fiber.Map{"data": productos, "page": page, "limit": limit, "total": total})
}

// GetProductoPorID maneja el endpoint GET /api/productos/:id
func (pc *ProdController) GetProductoPorID(c *fiber.Ctx) error {
	// Convertir el parámetro ID de string a int
	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.Status(400).JSON(middleware.SanitizeError(models.APIError{
			Status:  400,
			Message: "El ID debe ser un número válido",
			Details: err.Error(),
		}))
	}

	// Buscar producto por ID
	producto, err := pc.repo.ObtenerProductoPorID(id)
	if err != nil {
		return c.Status(500).JSON(middleware.SanitizeError(models.APIError{
			Status:  500,
			Message: "Error al buscar producto",
			Details: err.Error(),
		}))
	}

	if producto == nil {
		return c.Status(404).JSON(middleware.SanitizeError(models.APIError{
			Status:  404,
			Message: "Producto no encontrado",
		}))
	}

	return c.JSON(producto)
}
