// Paquete controllers maneja la lógica de negocio de liquidaciones
package controllers

import (
	"multicatalogo-backend/interfaces"
	"multicatalogo-backend/middleware"
	"multicatalogo-backend/models"

	"github.com/gofiber/fiber/v2"
)

// LiquidacionesController es el controlador con dependencias inyectadas
type LiquidacionesController struct {
	repo interfaces.DBRepository
}

// NewLiquidacionesController crea una nueva instancia de LiquidacionesController
func NewLiquidacionesController(repo interfaces.DBRepository) *LiquidacionesController {
	return &LiquidacionesController{
		repo: repo,
	}
}

// CrearLiquidacion maneja el endpoint POST /api/liquidaciones (admin)
func (lc *LiquidacionesController) CrearLiquidacion(c *fiber.Ctx) error {
	var req struct {
		ReferidoID int     `json:"referido_id"`
		Total      float64 `json:"total"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(middleware.SanitizeError(models.APIError{
			Status:  400,
			Message: "Error al parsear la petición",
			Details: err.Error(),
		}))
	}

	if req.ReferidoID == 0 {
		return c.Status(400).JSON(middleware.SanitizeError(models.APIError{
			Status:  400,
			Message: "El referido_id es requerido",
		}))
	}

	err := lc.repo.CrearLiquidacion(req.ReferidoID, req.Total)
	if err != nil {
		return c.Status(500).JSON(middleware.SanitizeError(models.APIError{
			Status:  500,
			Message: "Error al crear liquidación",
			Details: err.Error(),
		}))
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "Liquidación creada exitosamente",
	})
}

// ObtenerLiquidacionesMias maneja el endpoint GET /api/liquidaciones/mias
func (lc *LiquidacionesController) ObtenerLiquidacionesMias(c *fiber.Ctx) error {
	referidoID := c.Locals("referido_id")
	if referidoID == nil {
		return c.Status(401).JSON(middleware.SanitizeError(models.APIError{
			Status:  401,
			Message: "No autorizado: usuario no tiene referido asociado",
		}))
	}

	liquidaciones, err := lc.repo.ObtenerLiquidacionesMias(referidoID.(int))
	if err != nil {
		return c.Status(500).JSON(middleware.SanitizeError(models.APIError{
			Status:  500,
			Message: "Error al obtener liquidaciones",
			Details: err.Error(),
		}))
	}

	return c.JSON(liquidaciones)
}

// ObtenerLiquidacionesTodas maneja el endpoint GET /api/liquidaciones (admin)
func (lc *LiquidacionesController) ObtenerLiquidacionesTodas(c *fiber.Ctx) error {
	liquidaciones, err := lc.repo.ObtenerLiquidacionesTodas()
	if err != nil {
		return c.Status(500).JSON(middleware.SanitizeError(models.APIError{
			Status:  500,
			Message: "Error al obtener liquidaciones",
			Details: err.Error(),
		}))
	}

	return c.JSON(liquidaciones)
}
