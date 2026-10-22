package controllers

import (
	"database/sql"
	"errors"
	"strconv"

	"multicatalogo-backend/interfaces"
	"multicatalogo-backend/middleware"
	"multicatalogo-backend/models"

	"github.com/gofiber/fiber/v2"
)

type UsuariosController struct{ repo interfaces.DBRepository }

func NewUsuariosController(repo interfaces.DBRepository) *UsuariosController {
	return &UsuariosController{repo: repo}
}

func (uc *UsuariosController) Listar(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	usuarios, total, err := uc.repo.ObtenerUsuariosPaginados(page, limit)
	if err != nil {
		return c.Status(500).JSON(middleware.SanitizeError(models.APIError{Status: 500, Message: "Error al obtener usuarios", Details: err.Error()}))
	}
	return c.JSON(fiber.Map{"data": usuarios, "page": page, "limit": limit, "total": total})
}

func (uc *UsuariosController) Actualizar(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id < 1 {
		return c.Status(400).JSON(models.APIError{Status: 400, Message: "El ID debe ser un número válido"})
	}
	var req struct {
		Activo *bool   `json:"activo"`
		Rol    *string `json:"rol"`
	}
	if err := c.BodyParser(&req); err != nil || (req.Activo == nil && req.Rol == nil) {
		return c.Status(400).JSON(models.APIError{Status: 400, Message: "Indica activo o rol para actualizar"})
	}
	if req.Rol != nil && *req.Rol != "admin" && *req.Rol != "cliente" {
		return c.Status(400).JSON(models.APIError{Status: 400, Message: "Rol inválido"})
	}
	if currentID, ok := c.Locals("user_id").(int); ok && currentID == id && ((req.Activo != nil && !*req.Activo) || (req.Rol != nil && *req.Rol != "admin")) {
		return c.Status(400).JSON(models.APIError{Status: 400, Message: "No puedes desactivar ni cambiar tu propio rol"})
	}
	if err := uc.repo.ActualizarUsuario(id, req.Activo, req.Rol); err != nil {
		status := 500
		message := "Error al actualizar usuario"
		if errors.Is(err, sql.ErrNoRows) {
			status, message = 404, "Usuario no encontrado"
		}
		return c.Status(status).JSON(middleware.SanitizeError(models.APIError{Status: status, Message: message, Details: err.Error()}))
	}
	return c.JSON(fiber.Map{"message": "Usuario actualizado"})
}
