import { useState, type FormEvent } from 'react';
import { Link, useNavigate, useSearchParams } from 'react-router-dom';
import { api } from '../services/api';
import { useOnlineStatus } from '../hooks/useOnlineStatus';

export default function Registro() {
  const [params] = useSearchParams();
  const navigate = useNavigate();
  const [nombre, setNombre] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [ref, setRef] = useState(params.get('ref') ?? '');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const online = useOnlineStatus();

  const submit = async (event: FormEvent) => {
    event.preventDefault();
    if (!online) {
      setError('Necesitas conexión a internet para crear una cuenta.');
      return;
    }
    setLoading(true);
    setError('');
    try {
      await api('/api/register', { method: 'POST', body: JSON.stringify({ nombre, email, password, ref: ref.trim() }) }, { auth: false });
      navigate('/login', { state: { message: 'Cuenta creada. Inicia sesión para continuar.' } });
    } catch (err) {
      setError(err instanceof Error ? err.message : 'No se pudo crear la cuenta.');
    } finally { setLoading(false); }
  };

  return <main className="min-h-screen bg-slate-100 px-4 py-10 flex items-center justify-center">
    <form onSubmit={submit} className="w-full max-w-md rounded-xl border border-slate-200 bg-white p-8 shadow-sm space-y-5">
      <div><p className="text-sm font-semibold text-indigo-600">MultiCatálogo</p><h1 className="mt-1 text-2xl font-bold text-slate-900">Crear cuenta</h1><p className="mt-2 text-sm text-slate-500">Regístrate para comprar y formar parte de la red.</p></div>
      {error && <p role="alert" className="rounded-lg bg-red-50 p-3 text-sm text-red-700">{error}</p>}
      <label className="block text-sm font-medium">Nombre completo<input required value={nombre} onChange={e => setNombre(e.target.value)} className="mt-1 w-full rounded-lg border border-slate-300 px-3 py-2" /></label>
      <label className="block text-sm font-medium">Correo electrónico<input required type="email" value={email} onChange={e => setEmail(e.target.value)} className="mt-1 w-full rounded-lg border border-slate-300 px-3 py-2" /></label>
      <label className="block text-sm font-medium">Contraseña<input required minLength={8} type="password" value={password} onChange={e => setPassword(e.target.value)} className="mt-1 w-full rounded-lg border border-slate-300 px-3 py-2" /></label>
      <label className="block text-sm font-medium">Código de referido (opcional)<input value={ref} onChange={e => setRef(e.target.value)} className="mt-1 w-full rounded-lg border border-slate-300 px-3 py-2" /></label>
      <button disabled={loading || !online} className="w-full rounded-lg bg-indigo-600 py-3 font-semibold text-white hover:bg-indigo-700 disabled:opacity-50">{loading ? 'Creando cuenta…' : online ? 'Crear cuenta' : 'Conéctate para crear cuenta'}</button>
      <p className="text-center text-sm text-slate-600">¿Ya tienes cuenta? <Link to="/login" className="font-semibold text-indigo-600">Inicia sesión</Link></p>
    </form>
  </main>;
}
