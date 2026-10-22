// Paquete models define las estructuras relacionadas con el checkout
package models

// CheckoutRequest define la estructura esperada para el checkout
// NOTA: El email y usuario_id se obtienen del token JWT, no del body
type CheckoutRequest struct {
	Direccion    string  `json:"direccion"`
	Ciudad       string  `json:"ciudad"`
	Provincia    string  `json:"provincia"`
	CodigoPostal string  `json:"codigo_postal"`
	MetodoPago   string  `json:"metodo_pago"`
	Items        []Item  `json:"items"`
}

// Item representa un producto en el carrito
type Item struct {
	ProductoID int     `json:"producto_id"`
	Cantidad   int     `json:"cantidad"`
	Precio     float64 `json:"precio"`
	Nombre     string  `json:"nombre"`
}

// CheckoutResponse define la respuesta del checkout
type CheckoutResponse struct {
	PedidoID int     `json:"pedido_id"`
	Total    float64 `json:"total"`
	Estado   string  `json:"estado"`
	Mensaje  string  `json:"mensaje"`
}
