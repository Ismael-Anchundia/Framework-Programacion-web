// Paquete controllers maneja la lógica de negocio de comisiones
package controllers

import (
	"multicatalogo-backend/interfaces"
	"multicatalogo-backend/middleware"
	"multicatalogo-backend/models"

	"github.com/gofiber/fiber/v2"
)

// ComisionesController es el controlador con dependencias inyectadas
type ComisionesController struct {
	repo interfaces.DBRepository
}

// NewComisionesController crea una nueva instancia de ComisionesController
func NewComisionesController(repo interfaces.DBRepository) *ComisionesController {
	return &ComisionesController{
		repo: repo,
	}
}

// ObtenerComisionesMias maneja el endpoint GET /api/comisiones/mias
func (cc *ComisionesController) ObtenerComisionesMias(c *fiber.Ctx) error {
	referidoID := c.Locals("referido_id")
	if referidoID == nil {
		return c.Status(401).JSON(middleware.SanitizeError(models.APIError{
			Status:  401,
			Message: "No autorizado: usuario no tiene referido asociado",
		}))
	}

	comisiones, err := cc.repo.ObtenerComisionesMias(referidoID.(int))
	if err != nil {
		return c.Status(500).JSON(middleware.SanitizeError(models.APIError{
			Status:  500,
			Message: "Error al obtener comisiones",
			Details: err.Error(),
		}))
	}

	// Calcular totales
	var totalPendiente, totalPagada float64
	for _, comision := range comisiones {
		if comision.Estado == "pendiente" {
			totalPendiente += comision.Monto
		} else if comision.Estado == "pagada" {
			totalPagada += comision.Monto
		}
	}

	return c.JSON(fiber.Map{
		"comisiones":       comisiones,
		"total_pendiente":  totalPendiente,
		"total_pagada":     totalPagada,
	})
}

// ObtenerComisionesPendientes maneja el endpoint GET /api/comisiones/pendientes (admin)
func (cc *ComisionesController) ObtenerComisionesPendientes(c *fiber.Ctx) error {
	comisiones, err := cc.repo.ObtenerComisionesPendientes()
	if err != nil {
		return c.Status(500).JSON(middleware.SanitizeError(models.APIError{
			Status:  500,
			Message: "Error al obtener comisiones pendientes",
			Details: err.Error(),
		}))
	}

	return c.JSON(comisiones)
}
