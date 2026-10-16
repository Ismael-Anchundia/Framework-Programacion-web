// Paquete interfaces define los contratos para inyección de dependencias
package interfaces

import (
	"multicatalogo-backend/models"
)

// DBRepository define la interfaz para operaciones de base de datos
type DBRepository interface {
	// Operaciones de usuarios
	BuscarUsuarioPorCredenciales(email, password string) (*UsuarioDB, error)
	ObtenerUsuarioPorEmail(email string) (*UsuarioDB, error)

	// Operaciones de productos
	ObtenerProductos() ([]models.Producto, error)
	ObtenerProductoPorID(id int) (*models.Producto, error)

	// Operaciones de red multinivel
	ObtenerRed(email string) (*ReferidoDB, error)

	// Operaciones de pedidos (transacciones ACID)
	CrearPedido(pedido *PedidoDB, items []PedidoItemDB) error
	ActualizarComisiones(total float64, usuarioID int) error
}

// CacheRepository define la interfaz para operaciones de caché
type CacheRepository interface {
	Get(key string) (string, error)
	Set(key string, value string, ttlSeconds int) error
	Delete(key string) error
}

// UsuarioDB representa un usuario en la base de datos
type UsuarioDB struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Rol      string `json:"rol"`
}

// ReferidoDB representa un referido en la red multinivel
type ReferidoDB struct {
	ID       int          `json:"id"`
	Nombre   string       `json:"nombre"`
	Nivel    int          `json:"nivel"`
	Ventas   float64      `json:"ventas"`
	ParentID *int         `json:"parent_id,omitempty"`
	Hijos    []ReferidoDB `json:"hijos,omitempty"`
}

// PedidoDB representa un pedido en la base de datos
type PedidoDB struct {
	ID          int     `json:"id"`
	UsuarioID   int     `json:"usuario_id"`
	Total       float64 `json:"total"`
	Estado      string  `json:"estado"`
	Direccion   string  `json:"direccion"`
	MetodoPago  string  `json:"metodo_pago"`
	Ciudad      string  `json:"ciudad"`
	Provincia   string  `json:"provincia"`
	CodigoPostal string `json:"codigo_postal"`
}

// PedidoItemDB representa un ítem de pedido en la base de datos
type PedidoItemDB struct {
	ID        int     `json:"id"`
	PedidoID  int     `json:"pedido_id"`
	ProductoID int    `json:"producto_id"`
	Cantidad  int     `json:"cantidad"`
	Precio    float64 `json:"precio"`
	Subtotal  float64 `json:"subtotal"`
}
