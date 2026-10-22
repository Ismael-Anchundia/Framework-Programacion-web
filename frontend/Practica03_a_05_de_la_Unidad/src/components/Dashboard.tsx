import { useEffect, useState } from 'react';
import { api } from '../services/api';

type Resumen = { total_pedidos: number; total_vendido: number; ticket_promedio: number; comisiones_devengadas: number };
type TopProducto = { id: number; nombre: string; total_vendido: number; total_ingresos: number };
type TopReferido = { id: number; nombre: string; nivel: number; ventas: number; total_comisiones: number };

const Dashboard = () => {
  const [resumen,setResumen]=useState<Resumen|null>(null);
  const [productos,setProductos]=useState<TopProducto[]>([]);
  const [referidos,setReferidos]=useState<TopReferido[]>([]);
  const [error,setError]=useState('');
  useEffect(()=>{
    Promise.all([
      api<Resumen>('/api/reportes/resumen'),
      api<TopProducto[]>('/api/reportes/top-productos?limit=5'),
      api<TopReferido[]>('/api/reportes/top-referidos?limit=5'),
    ]).then(([r,p,refs])=>{setResumen(r);setProductos(p??[]);setReferidos(refs??[]);}).catch(e=>setError(e.message));
  },[]);
  const cards=[['Ventas',`$${Number(resumen?.total_vendido??0).toFixed(2)}`],['Pedidos',String(resumen?.total_pedidos??0)],['Ticket promedio',`$${Number(resumen?.ticket_promedio??0).toFixed(2)}`],['Comisiones devengadas',`$${Number(resumen?.comisiones_devengadas??0).toFixed(2)}`]];
  return <main className="space-y-7"><header><h1 className="text-2xl font-bold text-slate-900">Resumen del negocio</h1><p className="mt-1 text-sm text-slate-500">Indicadores calculados desde los pedidos y comisiones registrados.</p></header>
    {error&&<p role="alert" className="rounded-lg bg-amber-50 p-3 text-sm text-amber-800">No fue posible actualizar los reportes: {error}</p>}
    <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">{cards.map(([label,value])=><article key={label} className="rounded-xl border border-slate-200 bg-white p-5"><p className="text-sm font-semibold uppercase text-slate-500">{label}</p><p className="mt-2 text-3xl font-bold text-indigo-700">{value}</p></article>)}</div>
    <div className="grid gap-5 xl:grid-cols-2"><section className="rounded-xl border border-slate-200 bg-white p-5"><h2 className="mb-4 text-lg font-bold">Productos más vendidos</h2>{productos.length?productos.map((p,i)=><div key={p.id} className="flex justify-between border-t py-3 text-sm"><span>{i+1}. {p.nombre} <span className="text-slate-500">({p.total_vendido} unidades)</span></span><strong>${Number(p.total_ingresos).toFixed(2)}</strong></div>):<p className="py-4 text-sm text-slate-500">Aún no hay ventas registradas.</p>}</section>
    <section className="rounded-xl border border-slate-200 bg-white p-5"><h2 className="mb-4 text-lg font-bold">Referidos destacados</h2>{referidos.length?referidos.map((r,i)=><div key={r.id} className="flex justify-between border-t py-3 text-sm"><span>{i+1}. {r.nombre} <span className="text-slate-500">(nivel {r.nivel})</span></span><strong>${Number(r.ventas).toFixed(2)}</strong></div>):<p className="py-4 text-sm text-slate-500">Aún no hay referidos registrados.</p>}</section></div>
  </main>;
};

export default Dashboard;
