// Paquete controllers maneja la lógica de negocio del checkout
package controllers

import (
	"log"
	"multicatalogo-backend/cache"
	"multicatalogo-backend/interfaces"
	"multicatalogo-backend/middleware"
	"multicatalogo-backend/models"

	"github.com/gofiber/fiber/v2"
)

// CheckoutController es el controlador con dependencias inyectadas
type CheckoutController struct {
	repo  interfaces.DBRepository
	cache *cache.RedisCache
}

// NewCheckoutController crea una nueva instancia de CheckoutController
func NewCheckoutController(repo interfaces.DBRepository, cache *cache.RedisCache) *CheckoutController {
	return &CheckoutController{
		repo:  repo,
		cache: cache,
	}
}

// ProcesarCheckout maneja el endpoint POST /api/checkout
func (cc *CheckoutController) ProcesarCheckout(c *fiber.Ctx) error {
	// Parsear el cuerpo de la petición
	var req models.CheckoutRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(middleware.SanitizeError(models.APIError{
			Status:  400,
			Message: "Error al parsear la petición",
			Details: err.Error(),
		}))
	}

	// Obtener email del token JWT (NO del body del request)
	email := c.Locals("email")
	if email == nil {
		return c.Status(401).JSON(middleware.SanitizeError(models.APIError{
			Status:  401,
			Message: "No autorizado: sesión no iniciada",
		}))
	}

	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(401).JSON(middleware.SanitizeError(models.APIError{
			Status:  401,
			Message: "No autorizado: usuario no identificado",
		}))
	}

	// Validaciones básicas
	if len(req.Items) == 0 {
		return c.Status(400).JSON(middleware.SanitizeError(models.APIError{
			Status:  400,
			Message: "El carrito no puede estar vacío",
		}))
	}

	// Calcular total del pedido
	var total float64
	itemsDB := make([]interfaces.PedidoItemDB, len(req.Items))
	for i, item := range req.Items {
		subtotal := float64(item.Cantidad) * item.Precio
		total += subtotal

		itemsDB[i] = interfaces.PedidoItemDB{
			ProductoID: item.ProductoID,
			Cantidad:   item.Cantidad,
			Precio:     item.Precio,
			Subtotal:   subtotal,
		}
	}

	// Crear pedido con transacción ACID
	pedidoDB := &interfaces.PedidoDB{
		UsuarioID:    userID.(int),
		Total:        total,
		Estado:       "pendiente",
		Direccion:    req.Direccion,
		MetodoPago:   req.MetodoPago,
		Ciudad:       req.Ciudad,
		Provincia:    req.Provincia,
		CodigoPostal: req.CodigoPostal,
	}

	err := cc.repo.CrearPedido(pedidoDB, itemsDB)
	if err != nil {
		return c.Status(500).JSON(middleware.SanitizeError(models.APIError{
			Status:  500,
			Message: "Error al crear pedido",
			Details: err.Error(),
		}))
	}

	// Crear comisiones en el ledger (síncrono para garantizar consistencia)
	err = cc.repo.CrearComisiones(pedidoDB.ID, userID.(int), total)
	if err != nil {
		log.Printf("Error al crear comisiones: %v", err)
		// No retornamos error porque el pedido ya se creó
	}

	// Invalidar caché de la red del usuario
	cacheKey := "red:tree:" + email.(string)
	go func() {
		err = cc.cache.Delete(cacheKey)
		if err != nil {
			log.Printf("Error al invalidar caché: %v", err)
		}
	}()

	// Retornar respuesta inmediata al cliente
	return c.Status(201).JSON(models.CheckoutResponse{
		PedidoID: pedidoDB.ID,
		Total:    total,
		Estado:   "pendiente",
		Mensaje:  "Pedido creado exitosamente",
	})
}
