import { useEffect, useMemo, useState, type FormEvent } from 'react';
import { api } from '../services/api';
import type { Producto } from '../data/productos';

const button = 'rounded-lg px-4 py-2 text-sm font-semibold transition disabled:opacity-50';
const primary = `${button} bg-indigo-600 text-white hover:bg-indigo-700`;
const secondary = `${button} border border-slate-300 bg-white text-slate-700 hover:bg-slate-50`;
const money = (value: number) => `$${Number(value || 0).toFixed(2)}`;

export function AdminProductos() {
  const [items, setItems] = useState<Producto[]>([]);
  const [editing, setEditing] = useState<Producto | null>(null);
  const [error, setError] = useState('');
  const [query, setQuery] = useState('');
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const blank: Producto = { id: 0, nombre: '', descripcion: '', precio: 0, categoria: '', img: '', galeria: [], stock: 0, activo: true };
  const load = () => api<Pagina<Producto>>(`/api/admin/productos?q=${encodeURIComponent(query)}&page=${page}&limit=10`).then(data => { setItems(data?.data ?? []); setTotal(data?.total ?? 0); }).catch(e => setError(e.message));
  useEffect(() => { void load(); }, [query, page]);
  const save = async (event: FormEvent) => {
    event.preventDefault(); if (!editing) return;
    setError('');
    try {
      const payload = { ...editing, precio: Number(editing.precio), stock: Number(editing.stock) };
      await api(editing.id ? `/api/admin/productos/${editing.id}` : '/api/admin/productos', { method: editing.id ? 'PUT' : 'POST', body: JSON.stringify(payload) });
      setEditing(null); await load();
    } catch (e) { setError(e instanceof Error ? e.message : 'No se pudo guardar el producto.'); }
  };
  const changeActive = async (product: Producto) => {
    try {
      if (product.activo) await api(`/api/admin/productos/${product.id}`, { method: 'DELETE' });
      else await api(`/api/admin/productos/${product.id}`, { method: 'PUT', body: JSON.stringify({ ...product, activo: true }) });
      await load();
    } catch (e) { setError(e instanceof Error ? e.message : 'No se pudo actualizar el producto.'); }
  };
  return <section className="space-y-5">
    <header className="flex flex-wrap items-end justify-between gap-3"><div><h1 className="text-2xl font-bold text-slate-900">Productos</h1><p className="mt-1 text-sm text-slate-500">Administra el catálogo y las existencias.</p></div><button className={primary} onClick={() => setEditing({ ...blank })}>Nuevo producto</button></header>
    {error && <p role="alert" className="rounded-lg bg-red-50 p-3 text-sm text-red-700">{error}</p>}
    <input value={query} onChange={e => {setQuery(e.target.value);setPage(1);}} placeholder="Buscar productos…" className="w-full max-w-sm rounded-lg border border-slate-300 bg-white px-3 py-2" />
    <div className="overflow-x-auto rounded-xl border border-slate-200 bg-white"><table className="w-full min-w-[680px] text-left text-sm"><thead className="bg-slate-50 text-slate-500"><tr><th className="p-3">Producto</th><th className="p-3">Categoría</th><th className="p-3">Precio</th><th className="p-3">Stock</th><th className="p-3">Estado</th><th className="p-3">Acciones</th></tr></thead><tbody>{items.map(p => <tr key={p.id} className="border-t border-slate-100"><td className="p-3"><div className="flex items-center gap-3"><img src={p.img} alt="" loading="lazy" onError={event => { event.currentTarget.onerror = null; event.currentTarget.src = '/product-placeholder.svg'; }} className="h-10 w-10 rounded object-cover" /><span className="font-semibold text-slate-800">{p.nombre}</span></div></td><td className="p-3">{p.categoria}</td><td className="p-3">{money(p.precio)}</td><td className="p-3">{p.stock ?? 0}</td><td className="p-3">{p.activo ? 'Activo' : 'Inactivo'}</td><td className="p-3"><div className="flex gap-2"><button className={secondary} onClick={() => setEditing({ ...p })}>Editar</button><button className={secondary} onClick={() => void changeActive(p)}>{p.activo ? 'Desactivar' : 'Activar'}</button></div></td></tr>)}</tbody></table>{items.length === 0 && <p className="p-8 text-center text-slate-500">No hay productos para mostrar.</p>}</div>
    <div className="flex items-center justify-between text-sm text-slate-600"><span>{total} productos</span><div className="flex items-center gap-2"><button className={secondary} disabled={page<=1} onClick={()=>setPage(p=>p-1)}>Anterior</button><span>Página {page} de {Math.max(1,Math.ceil(total/10))}</span><button className={secondary} disabled={page*10>=total} onClick={()=>setPage(p=>p+1)}>Siguiente</button></div></div>
    {editing && <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/50 p-4"><form onSubmit={save} className="max-h-[90vh] w-full max-w-xl space-y-4 overflow-y-auto rounded-xl bg-white p-6 shadow-xl"><div className="flex justify-between"><h2 className="text-xl font-bold">{editing.id ? 'Editar producto' : 'Nuevo producto'}</h2><button type="button" onClick={() => setEditing(null)} aria-label="Cerrar">✕</button></div><div className="grid gap-3 sm:grid-cols-2">{([['nombre','Nombre'],['categoria','Categoría'],['precio','Precio'],['stock','Stock'],['img','URL de imagen']] as const).map(([key,label]) => <label key={key} className="text-sm font-medium">{label}<input required={key !== 'img'} type={key === 'precio' || key === 'stock' ? 'number' : 'text'} min={key === 'precio' ? '0.01' : key === 'stock' ? '0' : undefined} step={key === 'precio' ? '0.01' : undefined} value={String(editing[key] ?? '')} onChange={e => setEditing({ ...editing, [key]: key === 'precio' || key === 'stock' ? Number(e.target.value) : e.target.value })} className="mt-1 w-full rounded-lg border border-slate-300 px-3 py-2" /></label>)}</div><label className="block text-sm font-medium">Descripción<textarea value={editing.descripcion} onChange={e => setEditing({ ...editing, descripcion: e.target.value })} rows={3} className="mt-1 w-full rounded-lg border border-slate-300 px-3 py-2" /></label>{editing.id > 0 && <label className="flex items-center gap-2 text-sm"><input type="checkbox" checked={Boolean(editing.activo)} onChange={e => setEditing({ ...editing, activo: e.target.checked })} />Producto activo</label>}<div className="flex justify-end gap-2"><button type="button" className={secondary} onClick={() => setEditing(null)}>Cancelar</button><button className={primary}>Guardar</button></div></form></div>}
  </section>;
}

type Pedido = { id: number; usuario_id: number; total: number; estado: string; ciudad: string; creado_en?: string; items?: { id: number; nombre: string; cantidad: number; precio: number; subtotal: number }[] };
type Pagina<T> = { data: T[]; page: number; limit: number; total: number };

function PedidoRows({ admin }: { admin: boolean }) {
  const [page, setPage] = useState(1); const [estado, setEstado] = useState('');
  const [result, setResult] = useState<Pagina<Pedido>>({ data: [], page: 1, limit: 10, total: 0 });
  const [selected, setSelected] = useState<Pedido | null>(null); const [error, setError] = useState('');
  const load = () => api<Pagina<Pedido>>(`${admin ? '/api/pedidos' : '/api/pedidos/mios'}?page=${page}&limit=10${admin && estado ? `&estado=${encodeURIComponent(estado)}` : ''}`).then(data => setResult({...data,data:data.data??[]})).catch(e => setError(e.message));
  useEffect(() => { void load(); }, [page, estado, admin]);
  const details = async (id: number) => { try { setSelected(await api<Pedido>(`/api/pedidos/${id}`)); } catch (e) { setError(e instanceof Error ? e.message : 'No se pudo cargar el pedido.'); } };
  const update = async (p: Pedido, next: string) => {
    try {
      if (next === 'cancelado') await api(`/api/pedidos/${p.id}/cancelar`, { method: 'POST' });
      else await api(`/api/pedidos/${p.id}/estado`, { method: 'PATCH', body: JSON.stringify({ estado: next }) });
      await load();
    } catch (e) { setError(e instanceof Error ? e.message : 'No se pudo actualizar el pedido.'); }
  };
  return <div className="space-y-4">{error && <p role="alert" className="rounded-lg bg-red-50 p-3 text-sm text-red-700">{error}</p>}{admin && <select value={estado} onChange={e => { setEstado(e.target.value); setPage(1); }} className="rounded-lg border border-slate-300 bg-white px-3 py-2"><option value="">Todos los estados</option>{['pendiente','preparando','enviado','entregado','cancelado'].map(s => <option key={s}>{s}</option>)}</select>}<div className="overflow-x-auto rounded-xl border border-slate-200 bg-white"><table className="w-full min-w-[650px] text-left text-sm"><thead className="bg-slate-50 text-slate-500"><tr><th className="p-3">Pedido</th>{admin && <th className="p-3">Usuario</th>}<th className="p-3">Fecha</th><th className="p-3">Estado</th><th className="p-3">Total</th><th className="p-3">Detalle</th>{admin && <th className="p-3">Gestión</th>}</tr></thead><tbody>{result.data.map(p => <tr key={p.id} className="border-t border-slate-100"><td className="p-3">#{p.id}</td>{admin && <td className="p-3">{p.usuario_id}</td>}<td className="p-3">{p.creado_en ?? '—'}</td><td className="p-3 capitalize">{p.estado}</td><td className="p-3">{money(p.total)}</td><td className="p-3"><button className="font-semibold text-indigo-600" onClick={() => void details(p.id)}>Ver</button></td>{admin && <td className="p-3"><select value={p.estado} disabled={['cancelado','entregado'].includes(p.estado)} onChange={e => void update(p,e.target.value)} className="rounded border border-slate-300 p-1"><option value={p.estado}>{p.estado}</option>{(p.estado === 'pendiente' ? ['preparando','cancelado'] : p.estado === 'preparando' ? ['enviado','cancelado'] : p.estado === 'enviado' ? ['entregado'] : []).map(s => <option key={s}>{s}</option>)}</select></td>}</tr>)}</tbody></table>{result.data.length === 0 && <p className="p-8 text-center text-slate-500">Todavía no hay pedidos.</p>}</div><div className="flex items-center justify-between text-sm text-slate-600"><span>{result.total} pedidos</span><div className="flex gap-2"><button className={secondary} disabled={page <= 1} onClick={() => setPage(v => v-1)}>Anterior</button><span className="px-2 py-2">{page}</span><button className={secondary} disabled={page*10 >= result.total} onClick={() => setPage(v => v+1)}>Siguiente</button></div></div>{selected && <div className="rounded-xl border border-slate-200 bg-white p-5"><div className="flex justify-between"><h2 className="font-bold">Detalle del pedido #{selected.id}</h2><button onClick={() => setSelected(null)}>Cerrar</button></div><p className="mt-2 text-sm text-slate-500">{selected.estado} · {selected.ciudad} · Total {money(selected.total)}</p><ul className="mt-4 divide-y">{selected.items?.map(item => <li key={item.id} className="flex justify-between py-2 text-sm"><span>{item.nombre} × {item.cantidad}</span><span>{money(item.subtotal)}</span></li>)}</ul></div>}</div>;
}

export function AdminPedidos() { return <section className="space-y-5"><header><h1 className="text-2xl font-bold text-slate-900">Pedidos</h1><p className="mt-1 text-sm text-slate-500">Consulta pedidos y actualiza su estado.</p></header><PedidoRows admin /></section>; }
export function MisPedidos() { return <section className="space-y-5"><header><h1 className="text-2xl font-bold text-slate-900">Mis pedidos</h1><p className="mt-1 text-sm text-slate-500">Historial y detalle de tus compras.</p></header><PedidoRows admin={false} /></section>; }

type Comision = { referido_id: number; referido_nombre?: string; monto: number; estado: string };
export function AdminLiquidaciones() {
  const [comisiones, setComisiones] = useState<Comision[]>([]); const [message, setMessage] = useState('');
  const load = () => api<Comision[]>('/api/comisiones/pendientes').then(data => setComisiones(data??[])).catch(e => setMessage(e.message));
  useEffect(() => { void load(); }, []);
  const groups = useMemo(() => Object.values(comisiones.reduce<Record<number,{id:number,nombre:string,total:number,count:number}>>((acc,c) => { const row=acc[c.referido_id] ?? {id:c.referido_id,nombre:c.referido_nombre??`Referido #${c.referido_id}`,total:0,count:0}; row.total+=Number(c.monto); row.count++; acc[c.referido_id]=row; return acc; },{})), [comisiones]);
  const pay = async (g: {id:number;total:number}) => { setMessage(''); try { await api('/api/liquidaciones', {method:'POST',body:JSON.stringify({referido_id:g.id,total:g.total})}); await load(); setMessage('Liquidación registrada.'); } catch(e) { setMessage(e instanceof Error ? e.message : 'No se pudo liquidar.'); } };
  return <section className="space-y-5"><header><h1 className="text-2xl font-bold text-slate-900">Liquidaciones</h1><p className="mt-1 text-sm text-slate-500">Comisiones pendientes agrupadas por distribuidor.</p></header>{message && <p role="status" className="rounded-lg bg-slate-100 p-3 text-sm">{message}</p>}<div className="overflow-x-auto rounded-xl border border-slate-200 bg-white"><table className="w-full min-w-[500px] text-left text-sm"><thead className="bg-slate-50 text-slate-500"><tr><th className="p-3">Referido</th><th className="p-3">Comisiones</th><th className="p-3">Pendiente</th><th className="p-3">Acción</th></tr></thead><tbody>{groups.map(g=><tr key={g.id} className="border-t"><td className="p-3">{g.nombre}</td><td className="p-3">{g.count}</td><td className="p-3">{money(g.total)}</td><td className="p-3"><button className={primary} onClick={()=>void pay(g)}>Liquidar</button></td></tr>)}</tbody></table>{groups.length===0&&<p className="p-8 text-center text-slate-500">No hay comisiones pendientes.</p>}</div><h2 className="text-lg font-bold">Historial de pagos</h2><LiquidationHistory /></section>;
}
function LiquidationHistory() {
  const [rows,setRows]=useState<{id:number;referido_id:number;total:number;estado:string;creado_en?:string}[]>([]);
  useEffect(()=>{void api<typeof rows>('/api/liquidaciones').then(data=>setRows(data??[])).catch(()=>setRows([]));},[]);
  return <div className="space-y-2">{rows.map(r=><div key={r.id} className="flex justify-between rounded-lg border border-slate-200 bg-white p-3 text-sm"><span>Liquidación #{r.id} · referido #{r.referido_id}</span><span>{money(r.total)} · {r.estado}</span></div>)}{!rows.length&&<p className="text-sm text-slate-500">Aún no hay liquidaciones.</p>}</div>;
}

type Usuario = {id:number;email:string;nombre:string;rol:string;activo:boolean;referido_id?:number;creado_en?:string};
export function AdminUsuarios() {
  const [rows,setRows]=useState<Usuario[]>([]); const [error,setError]=useState('');
  const [page,setPage]=useState(1); const [total,setTotal]=useState(0);
  const load=()=>api<Pagina<Usuario>>(`/api/admin/usuarios?page=${page}&limit=10`).then(data=>{setRows(data?.data??[]);setTotal(data?.total??0);}).catch(e=>setError(e.message));
  useEffect(()=>{void load();},[page]);
  const patch=async(u:Usuario,body:Partial<Usuario>)=>{try{await api(`/api/admin/usuarios/${u.id}`,{method:'PATCH',body:JSON.stringify(body)});await load();}catch(e){setError(e instanceof Error?e.message:'No se pudo actualizar.');}};
  return <section className="space-y-5"><header><h1 className="text-2xl font-bold text-slate-900">Usuarios</h1><p className="mt-1 text-sm text-slate-500">Cuentas registradas y estado de acceso.</p></header>{error&&<p role="alert" className="rounded-lg bg-red-50 p-3 text-sm text-red-700">{error}</p>}<div className="overflow-x-auto rounded-xl border border-slate-200 bg-white"><table className="w-full min-w-[650px] text-left text-sm"><thead className="bg-slate-50 text-slate-500"><tr><th className="p-3">Nombre</th><th className="p-3">Correo</th><th className="p-3">Rol</th><th className="p-3">Estado</th><th className="p-3">Alta</th><th className="p-3">Acción</th></tr></thead><tbody>{rows.map(u=><tr key={u.id} className="border-t border-slate-100"><td className="p-3">{u.nombre}</td><td className="p-3">{u.email}</td><td className="p-3"><select value={u.rol} onChange={e=>void patch(u,{rol:e.target.value})} className="rounded border p-1"><option value="cliente">Cliente</option><option value="admin">Admin</option></select></td><td className="p-3">{u.activo?'Activo':'Inactivo'}</td><td className="p-3">{u.creado_en??'—'}</td><td className="p-3"><button className={secondary} onClick={()=>void patch(u,{activo:!u.activo})}>{u.activo?'Desactivar':'Activar'}</button></td></tr>)}</tbody></table>{rows.length===0&&<p className="p-8 text-center text-slate-500">No hay usuarios.</p>}</div><div className="flex items-center justify-between text-sm text-slate-600"><span>{total} usuarios</span><div className="flex items-center gap-2"><button className={secondary} disabled={page<=1} onClick={()=>setPage(p=>p-1)}>Anterior</button><span>Página {page} de {Math.max(1,Math.ceil(total/10))}</span><button className={secondary} disabled={page*10>=total} onClick={()=>setPage(p=>p+1)}>Siguiente</button></div></div></section>;
}
