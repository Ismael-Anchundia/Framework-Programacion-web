import { Link } from "react-router-dom";

interface SidebarProps {
  isCollapsed: boolean;
  isMobileOpen: boolean;
  onCloseMobile: () => void;
}

const Sidebar = ({ isCollapsed, isMobileOpen, onCloseMobile }: SidebarProps) => {
  // En escritorio oculta el texto si está colapsado; en móvil siempre lo muestra
  const showText = !isCollapsed;

  return (
    <aside
      className={`
        fixed md:static top-0 left-0 h-full bg-emerald-950 text-white flex flex-col z-40
        transition-all duration-300 ease-in-out
        ${isMobileOpen ? "translate-x-0" : "-translate-x-full md:translate-x-0"}
        ${isCollapsed ? "md:w-20" : "md:w-64"} w-64
      `}
    >
      {/* Encabezado */}
      <div className="p-6 text-2xl font-bold border-b border-slate-700 flex items-center justify-between overflow-hidden">
        <div className="flex items-center justify-center w-full md:w-auto">
          {isCollapsed ? (
            <span className="hidden md:inline text-xl" title="MultiCatálogo">MC</span>
          ) : null}
          <span className={isCollapsed ? "md:hidden" : "block"}>MultiCatálogo</span>
        </div>
        
        {/* Botón X para cerrar solo en móvil */}
        <button onClick={onCloseMobile} className="md:hidden text-slate-400 hover:text-white">
          ✕
        </button>
      </div>

      {/* Menú de Navegación */}
      <nav className="flex-1 p-4 space-y-2">
        <Link
          to="/"
          onClick={onCloseMobile}
          className={`flex items-center gap-4 p-3 rounded hover:bg-slate-800 transition ${
            isCollapsed ? "md:justify-center" : ""
          }`}
          title="Dashboard"
        >
          <span className="text-xl">📊</span>
          <span className={showText ? "block" : "hidden md:hidden"}>Dashboard</span>
        </Link>

        <Link
          to="/catalogo"
          onClick={onCloseMobile}
          className={`flex items-center gap-4 p-3 rounded hover:bg-slate-800 transition ${
            isCollapsed ? "md:justify-center" : ""
          }`}
          title="Catálogo"
        >
          <span className="text-xl">🛍️</span>
          <span className={showText ? "block" : "hidden md:hidden"}>Catálogo</span>
        </Link>

        <Link
          to="/mi-red"
          onClick={onCloseMobile}
          className={`flex items-center gap-4 p-3 rounded hover:bg-slate-800 transition ${
            isCollapsed ? "md:justify-center" : ""
          }`}
          title="Mi Red"
        >
          <span className="text-xl">👥</span>
          <span className={showText ? "block" : "hidden md:hidden"}>Mi Red</span>
        </Link>
      </nav>
    </aside>
  );
};

export default Sidebar;