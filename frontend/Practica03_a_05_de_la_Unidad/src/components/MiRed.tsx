import { useEffect, useMemo, useState } from 'react';
import { api } from '../services/api';
import { useAuth } from '../context/AuthContext';

type Nodo = { id:number; nombre:string; nivel:number; ventas:number; codigo?:string; hijos?:Nodo[] };
type Comisiones = { comisiones:{id:number;pedido_id:number;nivel:number;monto:number;estado:string;creado_en?:string}[]; total_pendiente:number; total_pagada:number };
type Liquidacion = {id:number;total:number;estado:string;creado_en?:string;pagada_en?:string};
const count=(node:Nodo):number=>1+(node.hijos??[]).reduce((sum,child)=>sum+count(child),0);
const sales=(node:Nodo):number=>Number(node.ventas||0);

function Tree({node}:{node:Nodo}) { return <li><article className="flex flex-wrap items-center justify-between gap-3 rounded-lg border border-slate-200 bg-white p-4"><div><p className="font-semibold">{node.nombre}</p><p className="text-xs text-slate-500">Nivel {node.nivel} · código {node.codigo||'—'}</p></div><p className="text-sm">Ventas: <strong>${Number(node.ventas).toFixed(2)}</strong></p></article>{Boolean(node.hijos?.length)&&<ul className="ml-5 mt-3 space-y-3 border-l-2 border-slate-200 pl-4">{node.hijos?.map(child=><Tree key={child.id} node={child}/>)}</ul>}</li>; }

export default function MiRed() {
  const {user}=useAuth();
  const [red,setRed]=useState<Nodo|null>(null);
  const [ledger,setLedger]=useState<Comisiones|null>(null);
  const [liquidaciones,setLiquidaciones]=useState<Liquidacion[]>([]);
  const [error,setError]=useState(''); const [copied,setCopied]=useState(false);
  useEffect(()=>{
    const treePath = user?.rol === 'admin' ? '/api/red' : '/api/red/mia';
    Promise.all([
      api<Nodo>(treePath),
      user?.rol === 'admin' ? Promise.resolve<Comisiones>({comisiones:[],total_pendiente:0,total_pagada:0}) : api<Comisiones>('/api/comisiones/mias'),
      user?.rol === 'admin' ? Promise.resolve<Liquidacion[]>([]) : api<Liquidacion[]>('/api/liquidaciones/mias'),
    ])
      .then(([tree,commissions,payments])=>{setRed(tree);setLedger({...commissions,comisiones:commissions.comisiones??[]});setLiquidaciones(payments??[]);})
      .catch(e=>setError(e.message));
  },[user?.rol]);
  const invite=useMemo(()=>red?.codigo?`${window.location.origin}/registro?ref=${encodeURIComponent(red.codigo)}`:'',[red]);
  const copy=async()=>{if(!invite)return;try{await navigator.clipboard.writeText(invite);setCopied(true);setTimeout(()=>setCopied(false),1800);}catch{setError('No se pudo copiar el enlace.');}};
  return <main className="space-y-6"><header className="flex flex-wrap items-end justify-between gap-3"><div><h1 className="text-2xl font-bold text-slate-900">Mi red</h1><p className="mt-1 text-sm text-slate-500">Tu árbol de referidos y el estado de tus comisiones.</p></div>{invite&&<button onClick={()=>void copy()} className="rounded-lg bg-indigo-600 px-4 py-2 text-sm font-semibold text-white">{copied?'Enlace copiado':'Copiar enlace de invitación'}</button>}</header>
    {error&&<p role="alert" className="rounded-lg bg-amber-50 p-3 text-sm text-amber-800">{error}</p>}
    {invite&&<div className="break-all rounded-lg border border-indigo-100 bg-indigo-50 p-3 text-sm text-indigo-800">{invite}</div>}
    <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">{[['Miembros',red?String(Math.max(0,count(red)-1)):'—'],['Ventas de la red',red?`$${sales(red).toFixed(2)}`:'—'],['Comisiones pendientes',ledger?`$${Number(ledger.total_pendiente).toFixed(2)}`:'—'],['Comisiones pagadas',ledger?`$${Number(ledger.total_pagada).toFixed(2)}`:'—']].map(([label,value])=><article key={label} className="rounded-xl border border-slate-200 bg-white p-5"><p className="text-sm font-semibold uppercase text-slate-500">{label}</p><p className="mt-2 text-2xl font-bold text-indigo-700">{value}</p></article>)}</div>
    <section className="rounded-xl border border-slate-200 bg-slate-50 p-5"><h2 className="mb-4 font-bold">Jerarquía de referidos</h2>{red?.hijos?.length?<ul className="space-y-3">{red.hijos.map(node=><Tree key={node.id} node={node}/>)}</ul>:<p className="text-sm text-slate-500">Tu red todavía no tiene miembros. Comparte el enlace para invitar a alguien.</p>}<p className="mt-4 text-xs text-slate-500">Cuenta: {user?.email}</p></section>
    <section className="overflow-x-auto rounded-xl border border-slate-200 bg-white p-5"><h2 className="mb-4 font-bold">Mis comisiones</h2><table className="w-full min-w-[550px] text-left text-sm"><thead><tr className="text-slate-500"><th className="p-2">Pedido</th><th className="p-2">Nivel</th><th className="p-2">Fecha</th><th className="p-2">Estado</th><th className="p-2">Monto</th></tr></thead><tbody>{ledger?.comisiones.map(row=><tr key={row.id} className="border-t"><td className="p-2">#{row.pedido_id}</td><td className="p-2">{row.nivel}</td><td className="p-2">{row.creado_en??'—'}</td><td className="p-2 capitalize">{row.estado}</td><td className="p-2">${Number(row.monto).toFixed(2)}</td></tr>)}</tbody></table>{!ledger?.comisiones.length&&<p className="py-4 text-sm text-slate-500">Todavía no tienes comisiones registradas.</p>}</section>
    {user?.rol !== 'admin' && <section className="space-y-3"><h2 className="font-bold">Mis liquidaciones</h2>{liquidaciones.map(row=><article key={row.id} className="flex flex-wrap justify-between gap-2 rounded-lg border border-slate-200 bg-white p-4 text-sm"><span>Liquidación #{row.id} · {row.creado_en??'—'}</span><span className="capitalize">{row.estado} · ${Number(row.total).toFixed(2)}</span></article>)}{liquidaciones.length===0&&<p className="text-sm text-slate-500">Aún no tienes pagos registrados.</p>}</section>}
  </main>;
}
