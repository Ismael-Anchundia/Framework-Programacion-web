// Paquete controllers maneja la lógica de negocio de pedidos
package controllers

import (
	"multicatalogo-backend/interfaces"
	"multicatalogo-backend/middleware"
	"multicatalogo-backend/models"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// PedidosController es el controlador con dependencias inyectadas
type PedidosController struct {
	repo interfaces.DBRepository
}

// NewPedidosController crea una nueva instancia de PedidosController
func NewPedidosController(repo interfaces.DBRepository) *PedidosController {
	return &PedidosController{
		repo: repo,
	}
}

// ObtenerPedidosMios maneja el endpoint GET /api/pedidos/mios (cliente)
func (pc *PedidosController) ObtenerPedidosMios(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(401).JSON(middleware.SanitizeError(models.APIError{
			Status:  401,
			Message: "No autorizado",
		}))
	}

	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	pedidos, total, err := pc.repo.ObtenerPedidosMios(userID.(int), page, limit)
	if err != nil {
		return c.Status(500).JSON(middleware.SanitizeError(models.APIError{
			Status:  500,
			Message: "Error al obtener pedidos",
			Details: err.Error(),
		}))
	}

	return c.JSON(fiber.Map{
		"data":  pedidos,
		"page":  page,
		"limit": limit,
		"total": total,
	})
}

// ObtenerPedidosTodos maneja el endpoint GET /api/pedidos (admin)
func (pc *PedidosController) ObtenerPedidosTodos(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	estado := c.Query("estado")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	pedidos, total, err := pc.repo.ObtenerPedidosTodos(page, limit, estado)
	if err != nil {
		return c.Status(500).JSON(middleware.SanitizeError(models.APIError{
			Status:  500,
			Message: "Error al obtener pedidos",
			Details: err.Error(),
		}))
	}

	return c.JSON(fiber.Map{
		"data":  pedidos,
		"page":  page,
		"limit": limit,
		"total": total,
	})
}

// ObtenerPedidoPorID maneja el endpoint GET /api/pedidos/:id
func (pc *PedidosController) ObtenerPedidoPorID(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.Status(400).JSON(middleware.SanitizeError(models.APIError{
			Status:  400,
			Message: "El ID debe ser un número válido",
			Details: err.Error(),
		}))
	}

	pedido, err := pc.repo.ObtenerPedidoPorID(id)
	if err != nil {
		return c.Status(500).JSON(middleware.SanitizeError(models.APIError{
			Status:  500,
			Message: "Error al obtener pedido",
			Details: err.Error(),
		}))
	}

	if pedido == nil {
		return c.Status(404).JSON(middleware.SanitizeError(models.APIError{
			Status:  404,
			Message: "Pedido no encontrado",
		}))
	}
	if c.Locals("rol") != "admin" && pedido.UsuarioID != c.Locals("user_id").(int) {
		return c.Status(404).JSON(middleware.SanitizeError(models.APIError{Status: 404, Message: "Pedido no encontrado"}))
	}

	return c.JSON(pedido)
}

// ActualizarEstadoPedido maneja el endpoint PATCH /api/pedidos/:id/estado (admin)
func (pc *PedidosController) ActualizarEstadoPedido(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.Status(400).JSON(middleware.SanitizeError(models.APIError{
			Status:  400,
			Message: "El ID debe ser un número válido",
			Details: err.Error(),
		}))
	}

	var req struct {
		Estado string `json:"estado"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(middleware.SanitizeError(models.APIError{
			Status:  400,
			Message: "Error al parsear la petición",
			Details: err.Error(),
		}))
	}

	if req.Estado == "" {
		return c.Status(400).JSON(middleware.SanitizeError(models.APIError{
			Status:  400,
			Message: "El estado es requerido",
		}))
	}
	if req.Estado == "cancelado" {
		err = pc.repo.CancelarPedido(id)
		if err != nil {
			return c.Status(400).JSON(middleware.SanitizeError(models.APIError{Status: 400, Message: err.Error()}))
		}
		return c.JSON(fiber.Map{"message": "Pedido cancelado y stock devuelto"})
	}

	err = pc.repo.ActualizarEstadoPedido(id, req.Estado)
	if err != nil {
		return c.Status(400).JSON(middleware.SanitizeError(models.APIError{
			Status:  400,
			Message: err.Error(),
		}))
	}

	return c.JSON(fiber.Map{
		"message": "Estado actualizado exitosamente",
	})
}

// CancelarPedido maneja el endpoint POST /api/pedidos/:id/cancelar (admin)
func (pc *PedidosController) CancelarPedido(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.Status(400).JSON(middleware.SanitizeError(models.APIError{
			Status:  400,
			Message: "El ID debe ser un número válido",
			Details: err.Error(),
		}))
	}

	err = pc.repo.CancelarPedido(id)
	if err != nil {
		return c.Status(500).JSON(middleware.SanitizeError(models.APIError{
			Status:  500,
			Message: "Error al cancelar pedido",
			Details: err.Error(),
		}))
	}

	return c.JSON(fiber.Map{
		"message": "Pedido cancelado exitosamente",
	})
}
