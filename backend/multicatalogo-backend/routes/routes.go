// Declaramos el paquete routes, cuya única responsabilidad es registrar los endpoints de la API.
package routes

import (
	// Importamos Fiber para poder recibir y configurar la instancia principal de la aplicación.
	"github.com/gofiber/fiber/v2"
	// Importamos nuestro paquete de controladores, donde residen las funciones que se ejecutarán en cada ruta.
	"multicatalogo-backend/controllers"
	// Importamos el middleware de autenticación
	"multicatalogo-backend/middleware"
)

// SetupRoutes es una función que recibe un puntero a la aplicación Fiber (*fiber.App) y los controladores para inyectar las rutas.
func SetupRoutes(app *fiber.App, authController *controllers.AuthController, prodController *controllers.ProdController, redController *controllers.RedController, checkoutController *controllers.CheckoutController, registerController *controllers.RegisterController, adminProductosController *controllers.AdminProductosController, pedidosController *controllers.PedidosController, comisionesController *controllers.ComisionesController, liquidacionesController *controllers.LiquidacionesController, reportesController *controllers.ReportesController, usuariosController *controllers.UsuariosController) {
	// Definimos una ruta HTTP GET en "/api/health" para verificar el estado del servidor.
	app.Get("/api/health", controllers.Health)

	// Definimos una ruta HTTP POST en "/api/login" y la enlazamos al controlador con DI.
	app.Post("/api/login", authController.Login)

	// Definimos una ruta HTTP POST en "/api/register" (pública)
	app.Post("/api/register", registerController.Register)
	app.Get("/api/admin/usuarios", middleware.RequireAuth, middleware.RequireAdmin, usuariosController.Listar)
	app.Patch("/api/admin/usuarios/:id", middleware.RequireAuth, middleware.RequireAdmin, usuariosController.Actualizar)

	// Definimos una ruta HTTP GET en "/api/productos" y la enlazamos al controlador con DI.
	app.Get("/api/productos", prodController.GetProductos)

	// Definimos una ruta HTTP GET en "/api/productos/:id" y la enlazamos al controlador con DI.
	app.Get("/api/productos/:id", prodController.GetProductoPorID)

	// CRUD de productos (admin - protegido con RequireAdmin)
	app.Post("/api/admin/productos", middleware.RequireAuth, middleware.RequireAdmin, adminProductosController.CrearProducto)
	app.Put("/api/admin/productos/:id", middleware.RequireAuth, middleware.RequireAdmin, adminProductosController.ActualizarProducto)
	app.Delete("/api/admin/productos/:id", middleware.RequireAuth, middleware.RequireAdmin, adminProductosController.DesactivarProducto)
	app.Get("/api/admin/productos", middleware.RequireAuth, middleware.RequireAdmin, adminProductosController.ObtenerProductosAdmin)

	// Gestión de pedidos
	app.Get("/api/pedidos/mios", middleware.RequireAuth, pedidosController.ObtenerPedidosMios)
	app.Get("/api/pedidos", middleware.RequireAuth, middleware.RequireAdmin, pedidosController.ObtenerPedidosTodos)
	app.Get("/api/pedidos/:id", middleware.RequireAuth, pedidosController.ObtenerPedidoPorID)
	app.Patch("/api/pedidos/:id/estado", middleware.RequireAuth, middleware.RequireAdmin, pedidosController.ActualizarEstadoPedido)
	app.Post("/api/pedidos/:id/cancelar", middleware.RequireAuth, middleware.RequireAdmin, pedidosController.CancelarPedido)

	// Comisiones
	app.Get("/api/comisiones/mias", middleware.RequireAuth, comisionesController.ObtenerComisionesMias)
	app.Get("/api/comisiones/pendientes", middleware.RequireAuth, middleware.RequireAdmin, comisionesController.ObtenerComisionesPendientes)

	// Liquidaciones
	app.Post("/api/liquidaciones", middleware.RequireAuth, middleware.RequireAdmin, liquidacionesController.CrearLiquidacion)
	app.Get("/api/liquidaciones", middleware.RequireAuth, middleware.RequireAdmin, liquidacionesController.ObtenerLiquidacionesTodas)
	app.Get("/api/liquidaciones/mias", middleware.RequireAuth, liquidacionesController.ObtenerLiquidacionesMias)

	// Reportes
	app.Get("/api/reportes/resumen", middleware.RequireAuth, middleware.RequireAdmin, reportesController.ObtenerReporteResumen)
	app.Get("/api/reportes/top-productos", middleware.RequireAuth, middleware.RequireAdmin, reportesController.ObtenerTopProductos)
	app.Get("/api/reportes/top-referidos", middleware.RequireAuth, middleware.RequireAdmin, reportesController.ObtenerTopReferidos)
	app.Get("/api/reportes/comisiones-pendientes", middleware.RequireAuth, middleware.RequireAdmin, reportesController.ObtenerComisionesPendientes)

	// Definimos una ruta HTTP GET en "/api/red" y la enlazamos al controlador con DI (protegido con RequireAuth)
	app.Get("/api/red", middleware.RequireAuth, redController.GetRed)

	// Definimos una ruta HTTP GET en "/api/red/mia" - red del usuario autenticado
	app.Get("/api/red/mia", middleware.RequireAuth, redController.GetRedMia)

	// Definimos una ruta HTTP POST en "/api/checkout" y la enlazamos al controlador con DI (protegido con RequireAuth)
	app.Post("/api/checkout", middleware.RequireAuth, checkoutController.ProcesarCheckout)
}
