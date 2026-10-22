// Paquete controllers maneja la lógica de negocio de reportes
package controllers

import (
	"multicatalogo-backend/interfaces"
	"multicatalogo-backend/middleware"
	"multicatalogo-backend/models"
	"time"

	"github.com/gofiber/fiber/v2"
)

// ReportesController es el controlador con dependencias inyectadas
type ReportesController struct {
	repo interfaces.DBRepository
}

// NewReportesController crea una nueva instancia de ReportesController
func NewReportesController(repo interfaces.DBRepository) *ReportesController {
	return &ReportesController{
		repo: repo,
	}
}

// ObtenerReporteResumen maneja el endpoint GET /api/reportes/resumen
func (rc *ReportesController) ObtenerReporteResumen(c *fiber.Ctx) error {
	desde := c.Query("desde")
	hasta := c.Query("hasta")
	for _, date := range []string{desde, hasta} {
		if date != "" {
			if _, err := time.Parse("2006-01-02", date); err != nil {
				return c.Status(400).JSON(models.APIError{Status: 400, Message: "Las fechas deben tener formato YYYY-MM-DD"})
			}
		}
	}
	if desde != "" && hasta != "" && desde > hasta {
		return c.Status(400).JSON(models.APIError{Status: 400, Message: "La fecha inicial no puede ser posterior a la final"})
	}

	reporte, err := rc.repo.ObtenerReporteResumen(desde, hasta)
	if err != nil {
		return c.Status(500).JSON(middleware.SanitizeError(models.APIError{
			Status:  500,
			Message: "Error al obtener reporte de resumen",
			Details: err.Error(),
		}))
	}

	return c.JSON(reporte)
}

// ObtenerTopProductos maneja el endpoint GET /api/reportes/top-productos
func (rc *ReportesController) ObtenerTopProductos(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 10)
	if limit < 1 || limit > 100 {
		limit = 10
	}

	productos, err := rc.repo.ObtenerTopProductos(limit)
	if err != nil {
		return c.Status(500).JSON(middleware.SanitizeError(models.APIError{
			Status:  500,
			Message: "Error al obtener top productos",
			Details: err.Error(),
		}))
	}

	return c.JSON(productos)
}

// ObtenerTopReferidos maneja el endpoint GET /api/reportes/top-referidos
func (rc *ReportesController) ObtenerTopReferidos(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 10)
	if limit < 1 || limit > 100 {
		limit = 10
	}

	referidos, err := rc.repo.ObtenerTopReferidos(limit)
	if err != nil {
		return c.Status(500).JSON(middleware.SanitizeError(models.APIError{
			Status:  500,
			Message: "Error al obtener top referidos",
			Details: err.Error(),
		}))
	}

	return c.JSON(referidos)
}

// ObtenerComisionesPendientes maneja el endpoint GET /api/reportes/comisiones-pendientes
func (rc *ReportesController) ObtenerComisionesPendientes(c *fiber.Ctx) error {
	comisiones, err := rc.repo.ObtenerComisionesPendientes()
	if err != nil {
		return c.Status(500).JSON(middleware.SanitizeError(models.APIError{
			Status:  500,
			Message: "Error al obtener comisiones pendientes",
			Details: err.Error(),
		}))
	}

	// Calcular total pendiente
	var totalPendiente float64
	for _, comision := range comisiones {
		totalPendiente += comision.Monto
	}

	return c.JSON(fiber.Map{
		"comisiones":      comisiones,
		"total_pendiente": totalPendiente,
	})
}
