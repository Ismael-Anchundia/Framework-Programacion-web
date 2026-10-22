// Paquete controllers maneja la lógica de negocio de productos para admin
package controllers

import (
	"multicatalogo-backend/interfaces"
	"multicatalogo-backend/middleware"
	"multicatalogo-backend/models"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// AdminProductosController es el controlador con dependencias inyectadas
type AdminProductosController struct {
	repo interfaces.DBRepository
}

// NewAdminProductosController crea una nueva instancia de AdminProductosController
func NewAdminProductosController(repo interfaces.DBRepository) *AdminProductosController {
	return &AdminProductosController{
		repo: repo,
	}
}

// CrearProducto maneja el endpoint POST /api/admin/productos (admin)
func (apc *AdminProductosController) CrearProducto(c *fiber.Ctx) error {
	var producto models.Producto

	if err := c.BodyParser(&producto); err != nil {
		return c.Status(400).JSON(middleware.SanitizeError(models.APIError{
			Status:  400,
			Message: "Error al parsear la petición",
			Details: err.Error(),
		}))
	}

	// Validaciones básicas
	if producto.Nombre == "" {
		return c.Status(400).JSON(middleware.SanitizeError(models.APIError{
			Status:  400,
			Message: "El nombre es requerido",
		}))
	}

	if producto.Precio <= 0 {
		return c.Status(400).JSON(middleware.SanitizeError(models.APIError{
			Status:  400,
			Message: "El precio debe ser mayor que cero",
		}))
	}
	if producto.Stock < 0 {
		return c.Status(400).JSON(middleware.SanitizeError(models.APIError{Status: 400, Message: "El stock no puede ser negativo"}))
	}

	err := apc.repo.CrearProducto(&producto)
	if err != nil {
		return c.Status(500).JSON(middleware.SanitizeError(models.APIError{
			Status:  500,
			Message: "Error al crear producto",
			Details: err.Error(),
		}))
	}

	return c.Status(201).JSON(producto)
}

// ActualizarProducto maneja el endpoint PUT /api/admin/productos/:id (admin)
func (apc *AdminProductosController) ActualizarProducto(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.Status(400).JSON(middleware.SanitizeError(models.APIError{
			Status:  400,
			Message: "El ID debe ser un número válido",
			Details: err.Error(),
		}))
	}

	var producto models.Producto
	if err := c.BodyParser(&producto); err != nil {
		return c.Status(400).JSON(middleware.SanitizeError(models.APIError{
			Status:  400,
			Message: "Error al parsear la petición",
			Details: err.Error(),
		}))
	}
	if producto.Nombre == "" || producto.Precio <= 0 || producto.Stock < 0 {
		return c.Status(400).JSON(middleware.SanitizeError(models.APIError{Status: 400, Message: "Nombre, precio mayor que cero y stock no negativo son requeridos"}))
	}

	err = apc.repo.ActualizarProducto(id, &producto)
	if err != nil {
		return c.Status(500).JSON(middleware.SanitizeError(models.APIError{
			Status:  500,
			Message: "Error al actualizar producto",
			Details: err.Error(),
		}))
	}

	return c.JSON(fiber.Map{
		"message": "Producto actualizado exitosamente",
	})
}

// DesactivarProducto maneja el endpoint DELETE /api/admin/productos/:id (admin)
func (apc *AdminProductosController) DesactivarProducto(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.Status(400).JSON(middleware.SanitizeError(models.APIError{
			Status:  400,
			Message: "El ID debe ser un número válido",
			Details: err.Error(),
		}))
	}

	err = apc.repo.DesactivarProducto(id)
	if err != nil {
		return c.Status(500).JSON(middleware.SanitizeError(models.APIError{
			Status:  500,
			Message: "Error al desactivar producto",
			Details: err.Error(),
		}))
	}

	return c.JSON(fiber.Map{
		"message": "Producto desactivado exitosamente",
	})
}

// ObtenerProductosAdmin maneja el endpoint GET /api/admin/productos (admin)
func (apc *AdminProductosController) ObtenerProductosAdmin(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	productos, total, err := apc.repo.ObtenerProductosAdminPaginados(c.Query("q"), page, limit)
	if err != nil {
		return c.Status(500).JSON(middleware.SanitizeError(models.APIError{
			Status:  500,
			Message: "Error al obtener productos",
			Details: err.Error(),
		}))
	}

	return c.JSON(fiber.Map{"data": productos, "page": page, "limit": limit, "total": total})
}
