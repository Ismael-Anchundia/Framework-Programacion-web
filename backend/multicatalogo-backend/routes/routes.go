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
func SetupRoutes(app *fiber.App, authController *controllers.AuthController, prodController *controllers.ProdController, redController *controllers.RedController, checkoutController *controllers.CheckoutController) {
	// Definimos una ruta HTTP GET en "/api/health" para verificar el estado del servidor.
	app.Get("/api/health", controllers.Health)

	// Definimos una ruta HTTP POST en "/api/login" y la enlazamos al controlador con DI.
	app.Post("/api/login", authController.Login)

	// Definimos una ruta HTTP GET en "/api/productos" y la enlazamos al controlador con DI.
	app.Get("/api/productos", prodController.GetProductos)

	// Definimos una ruta HTTP GET en "/api/productos/:id" y la enlazamos al controlador con DI.
	app.Get("/api/productos/:id", prodController.GetProductoPorID)

	// Definimos una ruta HTTP GET en "/api/red" y la enlazamos al controlador con DI (protegido con RequireAuth)
	app.Get("/api/red", middleware.RequireAuth, redController.GetRed)

	// Definimos una ruta HTTP POST en "/api/checkout" y la enlazamos al controlador con DI (protegido con RequireAuth)
	app.Post("/api/checkout", middleware.RequireAuth, checkoutController.ProcesarCheckout)
}