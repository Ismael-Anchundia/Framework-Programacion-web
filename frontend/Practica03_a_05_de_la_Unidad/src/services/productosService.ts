import { api } from './api';
import type { Producto } from '../data/productos';

export type Pagina<T> = { data: T[]; page: number; limit: number; total: number };

export const getProductos = (busqueda = '', page = 1, limit = 12, categoria = ''): Promise<Pagina<Producto>> => {
  const params = new URLSearchParams({ page: String(page), limit: String(limit) });
  if (busqueda.trim()) params.set('q', busqueda.trim());
  if (categoria && categoria !== 'Todas') params.set('categoria', categoria);
  return api<Pagina<Producto>>(`/api/productos?${params.toString()}`, {}, { auth: false });
};

export const getProductoById = (id: number): Promise<Producto | undefined> =>
  api<Producto>(`/api/productos/${id}`, {}, { auth: false });
