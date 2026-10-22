import { useOnlineStatus } from '../hooks/useOnlineStatus';

export default function NetworkStatusBanner() {
  const online = useOnlineStatus();
  if (online) return null;

  return (
    <aside className="fixed inset-x-0 top-0 z-[100] border-b border-amber-300 bg-amber-50 px-4 py-3 text-amber-950 shadow-sm" role="status" aria-live="polite">
      <div className="mx-auto flex max-w-7xl flex-wrap items-center justify-between gap-2 text-sm">
        <p><strong>Sin conexión.</strong> El catálogo disponible en caché puede estar desactualizado. Iniciar sesión, consultar datos privados y comprar requiere internet.</p>
        <button type="button" onClick={() => window.location.reload()} className="rounded-md border border-amber-700 px-3 py-1 font-semibold hover:bg-amber-100">
          Reintentar
        </button>
      </div>
    </aside>
  );
}
