// src/components/Catalogo.tsx
// Catálogo interactivo con búsqueda y filtros (Tema 5):
// - useMemo para filtrar la lista sin recalcular en cada render
// - Estado de carga y estado vacío
// - Tarjetas enlazadas al detalle /producto/:id
import { useEffect, useMemo, useState } from 'react';
import { Link, useSearchParams } from 'react-router-dom';
import { useCart } from '../context/CartContext';
import { getProductos } from '../services/productosService';
import { categorias, type Producto } from '../data/productos';

const normalizarCategoria = (value: string) => value.trim().toLocaleLowerCase('es');

const Catalogo = () => {
  const { addToCart } = useCart();
  const [productos, setProductos] = useState<Producto[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [error, setError] = useState('');
  const pageSize = 12;

  // Lectura del query string (?categoria=...) enviado desde el Storefront
  const [searchParams, setSearchParams] = useSearchParams();

  // Estado de los filtros
  const [busqueda, setBusqueda] = useState<string>("");
  const [categoria, setCategoria] = useState<string>(searchParams.get("categoria") ?? "Todas");

  useEffect(() => {
    let vigente = true;
    setLoading(true);
    getProductos(busqueda, page, pageSize, categoria)
      .then((data) => {
        if (vigente) { setProductos(data?.data ?? []); setTotal(data?.total ?? 0); setError(''); }
      })
      .catch((err) => { if (vigente) { setProductos([]); setError(err instanceof Error ? err.message : 'No se pudo cargar el catálogo.'); } })
      .finally(() => { if (vigente) setLoading(false); });
    return () => { vigente = false; };
  }, [busqueda, page, categoria]);

  // Filtrado con useMemo: solo se recalcula cuando cambian busqueda/categoria/productos
  const productosFiltrados = useMemo(() => {
    return productos.filter((prod) => {
      const coincideCategoria = categoria === "Todas"
        || normalizarCategoria(prod.categoria) === normalizarCategoria(categoria);
      return coincideCategoria;
    });
  }, [productos, categoria]);
  const categoriasDisponibles = useMemo(() => {
    const unicas = new Map<string, string>();
    [...categorias, ...productos.map((p) => p.categoria).filter(Boolean)].forEach((cat) => {
      const etiqueta = cat.trim();
      const clave = normalizarCategoria(etiqueta);
      if (clave && !unicas.has(clave)) unicas.set(clave, etiqueta);
    });
    return [...unicas.values()].sort((a, b) => a.localeCompare(b, 'es'));
  }, [productos]);

  const handleCategoriaChange = (value: string) => {
    setCategoria(value);
    setPage(1);
    // Sincronizamos el query string para que la URL sea compartible
    if (value === "Todas") {
      setSearchParams({});
    } else {
      setSearchParams({ categoria: value });
    }
  };

  if (loading) {
    return <div className="text-center p-8">Cargando catálogo...</div>;
  }

  return (
    <div>
      <h1 className="text-2xl font-bold text-slate-800 mb-2">Catálogo de Productos</h1>
      <p className="text-slate-500 mb-6">
        {total} producto(s) en el catálogo
      </p>
      {error && <p role="alert" className="mb-4 rounded-lg bg-red-50 p-3 text-sm text-red-700">{error}</p>}

      {/* ===== BARRA DE BÚSQUEDA Y FILTROS ===== */}
      <div className="flex flex-col md:flex-row gap-4 mb-8">
        <input
          type="text"
          value={busqueda}
          onChange={(e) => { setBusqueda(e.target.value); setPage(1); }}
          placeholder="Buscar por nombre o descripción..."
          className="flex-1 px-4 py-3 rounded-lg border border-slate-300 focus:ring-2 focus:ring-indigo-600 focus:border-indigo-600 outline-none transition"
        />
        <select
          value={categoria}
          onChange={(e) => handleCategoriaChange(e.target.value)}
          className="px-4 py-3 rounded-lg border border-slate-300 bg-white focus:ring-2 focus:ring-indigo-600 focus:border-indigo-600 outline-none transition"
        >
          <option value="Todas">Todas las categorías</option>
          {categoriasDisponibles.map((cat) => (
            <option key={cat} value={cat}>
              {cat}
            </option>
          ))}
        </select>
      </div>

      {/* ===== GRILLA DE PRODUCTOS ===== */}
      {productosFiltrados.length === 0 ? (
        <div className="bg-white p-10 rounded-lg border border-slate-200 text-center">
          <p className="text-slate-500">
            No hay productos que coincidan con tu búsqueda.
          </p>
          <button
            onClick={() => {
              setBusqueda("");
              setCategoria("Todas");
              setSearchParams({});
            }}
            className="mt-4 text-indigo-600 font-semibold hover:underline"
          >
            Limpiar filtros
          </button>
        </div>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
          {productosFiltrados.map((prod) => (
            <div
              key={prod.id}
              className="bg-white rounded-lg overflow-hidden border border-slate-200 shadow-sm flex flex-col"
            >
              <Link to={`/producto/${prod.id}`}>
                <img
                  src={prod.img}
                  alt={prod.nombre}
                  loading="lazy"
                  onError={(event) => { event.currentTarget.onerror = null; event.currentTarget.src = '/product-placeholder.svg'; }}
                  className="w-full h-40 object-cover hover:opacity-90 transition"
                />
              </Link>
              <div className="p-4 flex flex-col flex-1">
                <p className="text-xs uppercase tracking-wide text-indigo-500 font-semibold">
                  {prod.categoria}
                </p>
                <Link to={`/producto/${prod.id}`}>
                  <h3 className="font-semibold text-slate-700 hover:text-indigo-600 transition">
                    {prod.nombre}
                  </h3>
                </Link>
                <p className="text-indigo-600 font-bold mt-2 mb-4">
                  ${prod.precio.toFixed(2)}
                </p>

                  <button
                  onClick={() => addToCart(prod)}
                    disabled={prod.stock === 0}
                  className="mt-auto w-full bg-slate-900 text-white py-2 rounded text-sm hover:bg-indigo-600 transition"
                >
                  {prod.stock === 0 ? "Agotado" : "Añadir al Carrito"}
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
      <div className="mt-8 flex items-center justify-center gap-4">
        <button disabled={page <= 1} onClick={() => setPage(p => p - 1)} className="rounded-lg border border-slate-300 px-4 py-2 text-sm disabled:opacity-40">Anterior</button>
        <span className="text-sm text-slate-600">Página {page} de {Math.max(1, Math.ceil(total / pageSize))}</span>
        <button disabled={page * pageSize >= total} onClick={() => setPage(p => p + 1)} className="rounded-lg border border-slate-300 px-4 py-2 text-sm disabled:opacity-40">Siguiente</button>
      </div>
    </div>
  );
};

export default Catalogo;

