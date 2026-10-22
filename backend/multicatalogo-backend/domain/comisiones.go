// Paquete domain contiene funciones puras para lógica de dominio (testeables sin I/O)
package domain

// Comision representa una comisión calculada
type Comision struct {
	Nivel      int     `json:"nivel"`
	Porcentaje float64 `json:"porcentaje"`
	Monto      float64 `json:"monto"`
}

// RepartirComisiones calcula las comisiones multinivel (10%, 5%, 2%)
// Función pura sin I/O - testeable
func RepartirComisiones(total float64) []Comision {
	return []Comision{
		{Nivel: 1, Porcentaje: 10.0, Monto: total * 0.10},
		{Nivel: 2, Porcentaje: 5.0, Monto: total * 0.05},
		{Nivel: 3, Porcentaje: 2.0, Monto: total * 0.02},
	}
}
