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
	CrearUsuario(usuario *UsuarioDB, referidoID *int) error
	ObtenerUsuarios() ([]UsuarioDB, error)
	ObtenerUsuariosPaginados(page, limit int) ([]UsuarioDB, int, error)
	ActualizarUsuarioEstado(usuarioID int, activo bool) error
	ActualizarUsuario(usuarioID int, activo *bool, rol *string) error

	// Operaciones de productos
	ObtenerProductos(busqueda string) ([]models.Producto, error)
	ObtenerProductosPaginados(busqueda, categoria string, page, limit int) ([]models.Producto, int, error)
	ObtenerProductoPorID(id int) (*models.Producto, error)
	CrearProducto(producto *models.Producto) error
	ActualizarProducto(id int, producto *models.Producto) error
	DesactivarProducto(id int) error
	ObtenerProductosAdmin(busqueda string) ([]models.Producto, error)
	ObtenerProductosAdminPaginados(busqueda string, page, limit int) ([]models.Producto, int, error)

	// Operaciones de red multinivel
	ObtenerRed(email string) (*ReferidoDB, error)
	ObtenerRedPorReferidoID(referidoID int) (*ReferidoDB, error)
	ObtenerReferidoPorCodigo(codigo string) (*ReferidoDB, error)

	// Operaciones de pedidos (transacciones ACID)
	CrearPedido(pedido *PedidoDB, items []PedidoItemDB) error
	ObtenerPedidosMios(usuarioID int, page, limit int) ([]PedidoDB, int, error)
	ObtenerPedidosTodos(page, limit int, estado string) ([]PedidoDB, int, error)
	ObtenerPedidoPorID(id int) (*PedidoDB, error)
	ActualizarEstadoPedido(id int, estado string) error
	CancelarPedido(id int) error

	// Operaciones de comisiones
	CrearComisiones(pedidoID int, usuarioID int, total float64) error
	ObtenerComisionesMias(referidoID int) ([]ComisionDB, error)
	ObtenerComisionesPendientes() ([]ComisionDB, error)
	ActualizarComisiones(total float64, usuarioID int) error // Método legacy para compatibilidad

	// Operaciones de liquidaciones
	CrearLiquidacion(referidoID int, total float64) error
	ObtenerLiquidacionesMias(referidoID int) ([]LiquidacionDB, error)
	ObtenerLiquidacionesTodas() ([]LiquidacionDB, error)

	// Operaciones de reportes
	ObtenerReporteResumen(desde, hasta string) (map[string]interface{}, error)
	ObtenerTopProductos(limit int) ([]map[string]interface{}, error)
	ObtenerTopReferidos(limit int) ([]map[string]interface{}, error)
}

// CacheRepository define la interfaz para operaciones de caché
type CacheRepository interface {
	Get(key string) (string, error)
	Set(key string, value string, ttlSeconds int) error
	Delete(key string) error
}

// UsuarioDB representa un usuario en la base de datos
type UsuarioDB struct {
	ID         int    `json:"id"`
	Email      string `json:"email"`
	Password   string `json:"-"`
	Rol        string `json:"rol"`
	Nombre     string `json:"nombre"`
	Activo     bool   `json:"activo"`
	ReferidoID *int   `json:"referido_id,omitempty"`
	CreadoEn   string `json:"creado_en,omitempty"`
}

// ReferidoDB representa un referido en la red multinivel
type ReferidoDB struct {
	ID       int          `json:"id"`
	Nombre   string       `json:"nombre"`
	Nivel    int          `json:"nivel"`
	Ventas   float64      `json:"ventas"`
	ParentID *int         `json:"parent_id,omitempty"`
	Codigo   string       `json:"codigo,omitempty"`
	Hijos    []ReferidoDB `json:"hijos,omitempty"`
}

// PedidoDB representa un pedido en la base de datos
type PedidoDB struct {
	ID           int            `json:"id"`
	UsuarioID    int            `json:"usuario_id"`
	Total        float64        `json:"total"`
	Estado       string         `json:"estado"`
	Direccion    string         `json:"direccion"`
	MetodoPago   string         `json:"metodo_pago"`
	Ciudad       string         `json:"ciudad"`
	Provincia    string         `json:"provincia"`
	CodigoPostal string         `json:"codigo_postal"`
	CreadoEn     string         `json:"creado_en,omitempty"`
	Items        []PedidoItemDB `json:"items,omitempty"`
}

// PedidoItemDB representa un ítem de pedido en la base de datos
type PedidoItemDB struct {
	ID         int     `json:"id"`
	PedidoID   int     `json:"pedido_id"`
	ProductoID int     `json:"producto_id"`
	Cantidad   int     `json:"cantidad"`
	Precio     float64 `json:"precio"`
	Subtotal   float64 `json:"subtotal"`
	Nombre     string  `json:"nombre,omitempty"`
}

// ComisionDB representa una comisión en el ledger
type ComisionDB struct {
	ID             int     `json:"id"`
	PedidoID       int     `json:"pedido_id"`
	ReferidoID     int     `json:"referido_id"`
	Nivel          int     `json:"nivel"`
	Porcentaje     float64 `json:"porcentaje"`
	Monto          float64 `json:"monto"`
	Estado         string  `json:"estado"`
	LiquidacionID  *int    `json:"liquidacion_id,omitempty"`
	CreadoEn       string  `json:"creado_en,omitempty"`
	ReferidoNombre string  `json:"referido_nombre,omitempty"`
}

// LiquidacionDB representa una liquidación
type LiquidacionDB struct {
	ID         int     `json:"id"`
	ReferidoID int     `json:"referido_id"`
	Total      float64 `json:"total"`
	Estado     string  `json:"estado"`
	CreadoEn   string  `json:"creado_en,omitempty"`
	PagadaEn   string  `json:"pagada_en,omitempty"`
}
