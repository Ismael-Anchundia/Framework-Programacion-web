// src/components/Checkout.tsx
// Flujo de checkout simulado (Tema 5):
// - Formulario de envío con validación básica
// - Resumen del pedido en tiempo real
// - Simula el procesamiento del pago y redirige a /confirmacion
import { useState, type FormEvent } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useCart } from '../context/CartContext';
import { api } from '../services/api';
import { useAuth } from '../context/AuthContext';
import { useOnlineStatus } from '../hooks/useOnlineStatus';

interface DatosEnvio {
  nombre: string;
  email: string;
  direccion: string;
  ciudad: string;
  provincia: string;
  codigoPostal: string;
  telefono: string;
  metodoPago: string;
}

const METODOS_PAGO = ["Tarjeta de crédito", "Transferencia bancaria", "Efectivo contra entrega"];

const Checkout = () => {
  const { cart, totalPrice, clearCart } = useCart();
  const { user } = useAuth();
  const online = useOnlineStatus();
  const navigate = useNavigate();

  const [datos, setDatos] = useState<DatosEnvio>({
    nombre: "",
    email: "",
    direccion: "",
    ciudad: "",
    provincia: "",
    codigoPostal: "",
    telefono: "",
    metodoPago: METODOS_PAGO[0],
  });
  const [procesando, setProcesando] = useState<boolean>(false);
  const [error, setError] = useState<string>("");

  // Si el carrito está vacío, mostramos un aviso en lugar del formulario
  if (cart.length === 0) {
    return (
      <div className="bg-white p-10 rounded-lg border border-slate-200 text-center">
        <p className="text-slate-600 text-lg mb-4">No tienes productos en el carrito.</p>
        <Link to="/catalogo" className="text-indigo-600 font-semibold hover:underline">
          ← Ir al catálogo
        </Link>
      </div>
    );
  }

  const handleChange = (campo: keyof DatosEnvio, valor: string) => {
    setDatos((prev) => ({ ...prev, [campo]: valor }));
  };

  const handleSubmit = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (!online) {
      setError("Necesitas conexión a internet para confirmar el pedido. Tu carrito se conserva.");
      return;
    }
    setProcesando(true);
    setError("");
    try {
      const response = await api<{ pedido_id: number; total: number; estado: string; mensaje: string }>("/api/checkout", {
        method: "POST",
        body: JSON.stringify({
          direccion: datos.direccion,
          ciudad: datos.ciudad,
          provincia: datos.provincia,
          codigo_postal: datos.codigoPostal,
          metodo_pago: datos.metodoPago,
          items: cart.map(item => ({ producto_id: item.id, cantidad: item.cantidad, precio: item.precio })),
        }),
      });
      const pedido = {
        numero: `MC-${response.pedido_id}`,
        pedido_id: response.pedido_id,
        items: cart.map(item => ({ nombre: item.nombre, cantidad: item.cantidad, precio: item.precio })),
        total: response.total,
        estado: response.estado,
        cliente: { nombre: datos.nombre || user?.email || "Cliente", email: user?.email || "", ciudad: datos.ciudad },
        fecha: new Date().toLocaleString("es-EC"),
      };
      clearCart();
      navigate("/confirmacion", { state: { pedido } });
    } catch (err) {
      setError(err instanceof Error ? err.message : "No se pudo registrar el pedido.");
    } finally {
      setProcesando(false);
    }
  };

  return (
    <div>
      <h1 className="text-2xl font-bold text-slate-800 mb-6">Finalizar Compra</h1>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* ===== FORMULARIO DE ENVÍO ===== */}
        <form onSubmit={handleSubmit} className="lg:col-span-2 space-y-5">
          {error && <p role="alert" className="rounded-lg bg-red-50 p-3 text-sm text-red-700">{error}</p>}
          <div className="bg-white p-6 rounded-lg border border-slate-200 shadow-sm">
            <h2 className="text-lg font-bold text-slate-800 mb-4">Datos de Envío</h2>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-slate-700 mb-1">
                  Nombre completo *
                </label>
                <input
                  type="text"
                  required
                  value={datos.nombre}
                  onChange={(e) => handleChange("nombre", e.target.value)}
                  className="w-full px-4 py-2.5 rounded-lg border border-slate-300 focus:ring-2 focus:ring-indigo-600 outline-none transition"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-slate-700 mb-1">Provincia *</label>
                <input type="text" required value={datos.provincia} onChange={(e) => handleChange("provincia", e.target.value)} className="w-full px-4 py-2.5 rounded-lg border border-slate-300" />
              </div>
              <div>
                <label className="block text-sm font-medium text-slate-700 mb-1">Código postal *</label>
                <input type="text" required value={datos.codigoPostal} onChange={(e) => handleChange("codigoPostal", e.target.value)} className="w-full px-4 py-2.5 rounded-lg border border-slate-300" />
              </div>
              <div>
                <label className="block text-sm font-medium text-slate-700 mb-1">
                  Correo electrónico *
                </label>
                <input
                  type="email"
                  required
                  value={datos.email}
                  onChange={(e) => handleChange("email", e.target.value)}
                  className="w-full px-4 py-2.5 rounded-lg border border-slate-300 focus:ring-2 focus:ring-indigo-600 outline-none transition"
                />
              </div>
              <div className="md:col-span-2">
                <label className="block text-sm font-medium text-slate-700 mb-1">
                  Dirección *
                </label>
                <input
                  type="text"
                  required
                  value={datos.direccion}
                  onChange={(e) => handleChange("direccion", e.target.value)}
                  className="w-full px-4 py-2.5 rounded-lg border border-slate-300 focus:ring-2 focus:ring-indigo-600 outline-none transition"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-slate-700 mb-1">
                  Ciudad *
                </label>
                <input
                  type="text"
                  required
                  value={datos.ciudad}
                  onChange={(e) => handleChange("ciudad", e.target.value)}
                  className="w-full px-4 py-2.5 rounded-lg border border-slate-300 focus:ring-2 focus:ring-indigo-600 outline-none transition"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-slate-700 mb-1">
                  Teléfono *
                </label>
                <input
                  type="tel"
                  required
                  value={datos.telefono}
                  onChange={(e) => handleChange("telefono", e.target.value)}
                  className="w-full px-4 py-2.5 rounded-lg border border-slate-300 focus:ring-2 focus:ring-indigo-600 outline-none transition"
                />
              </div>
            </div>
          </div>

          <div className="bg-white p-6 rounded-lg border border-slate-200 shadow-sm">
            <h2 className="text-lg font-bold text-slate-800 mb-4">Método de Pago</h2>
            <div className="space-y-3">
              {METODOS_PAGO.map((metodo) => (
                <label
                  key={metodo}
                  className="flex items-center gap-3 p-3 rounded-lg border border-slate-200 cursor-pointer hover:border-indigo-300 transition"
                >
                  <input
                    type="radio"
                    name="metodoPago"
                    value={metodo}
                    checked={datos.metodoPago === metodo}
                    onChange={(e) => handleChange("metodoPago", e.target.value)}
                    className="accent-indigo-600"
                  />
                  <span className="text-sm text-slate-700">{metodo}</span>
                </label>
              ))}
            </div>
          </div>

          <button
            type="submit"
            disabled={procesando || !online}
            className="w-full bg-indigo-600 text-white py-3 rounded-lg font-semibold hover:bg-indigo-700 transition disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {procesando ? "Procesando pedido..." : online ? "Confirmar Pedido" : "Conéctate para comprar"}
          </button>
        </form>

        {/* ===== RESUMEN DEL PEDIDO ===== */}
        <div className="bg-white p-6 rounded-lg border border-slate-200 shadow-sm h-fit">
          <h2 className="text-lg font-bold text-slate-800 mb-4">Resumen del Pedido</h2>
          <div className="space-y-3 border-b border-slate-100 pb-4 mb-4">
            {cart.map((item) => (
              <div key={item.id} className="flex justify-between text-sm">
                <span className="text-slate-600">
                  {item.nombre} × {item.cantidad}
                </span>
                <span className="font-semibold text-slate-800">
                  ${(item.precio * item.cantidad).toFixed(2)}
                </span>
              </div>
            ))}
          </div>
          <div className="flex justify-between items-center">
            <span className="text-slate-800 font-bold">Total</span>
            <span className="text-2xl font-bold text-indigo-600">
              ${totalPrice.toFixed(2)}
            </span>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Checkout;
