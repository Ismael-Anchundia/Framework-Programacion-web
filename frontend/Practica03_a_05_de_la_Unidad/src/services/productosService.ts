import { api } from './api';
import type { Producto } from '../data/productos';

export type Pagina<T> = { data: T[]; page: number; limit: number; total: number };

// Older records or newly created products may omit `galeria` in the API
// response. Keep the frontend model consistent so detail views can always
// iterate over an array.
const normalizarProducto = (producto: Producto): Producto => ({
  ...producto,
  galeria: Array.isArray(producto.galeria) ? producto.galeria : [],
});

export const getProductos = async (busqueda = '', page = 1, limit = 12, categoria = ''): Promise<Pagina<Producto>> => {
  const params = new URLSearchParams({ page: String(page), limit: String(limit) });
  if (busqueda.trim()) params.set('q', busqueda.trim());
  if (categoria && categoria !== 'Todas') params.set('categoria', categoria);
  const pagina = await api<Pagina<Producto>>(`/api/productos?${params.toString()}`, {}, { auth: false });
  return { ...pagina, data: pagina.data.map(normalizarProducto) };
};

export const getProductoById = async (id: number): Promise<Producto | undefined> => {
  const producto = await api<Producto>(`/api/productos/${id}`, {}, { auth: false });
  return producto ? normalizarProducto(producto) : undefined;
};
