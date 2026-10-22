// Declaramos el paquete models para agrupar las estructuras de datos de la aplicación.
package models

// LoginRequest define la estructura esperada para el cuerpo de la petición (JSON) al iniciar sesión.
type LoginRequest struct {
	// Email representa el correo del usuario; la etiqueta `json:"email"` indica cómo se mapea el JSON a esta variable.
	Email    string `json:"email"`
	// Password representa la contraseña del usuario; se extrae del campo "password" del JSON entrante.
	Password string `json:"password"`
}

// Producto define la estructura de los datos de un artículo en nuestro catálogo.
type Producto struct {
	// ID es el identificador único numérico del producto.
	ID          int      `json:"id"`
	// Nombre es la descripción en texto del producto.
	Nombre      string   `json:"nombre"`
	// Descripción es la descripción detallada del producto.
	Descripcion string   `json:"descripcion"`
	// Precio es el costo del producto, almacenado como un número con decimales (float64).
	Precio      float64  `json:"precio"`
	// Categoría es la categoría a la que pertenece el producto.
	Categoria   string   `json:"categoria"`
	// Img es la URL o ruta que apunta a la fotografía o imagen del producto.
	Img         string   `json:"img"`
	// Galería es un array de URLs de imágenes adicionales.
	Galeria     []string `json:"galeria,omitempty"`
	// Stock es la cantidad disponible del producto.
	Stock       int      `json:"stock"`
	// Activo indica si el producto está activo (borrado lógico).
	Activo      bool     `json:"activo"`
}

// APIError estandariza los mensajes de error HTTP con una estructura consistente.
type APIError struct {
	Status  int         `json:"status"`  // Código de estado HTTP
	Message string      `json:"message"` // Mensaje descriptivo del error
	Details interface{} `json:"details,omitempty"` // Detalles adicionales opcionales
}