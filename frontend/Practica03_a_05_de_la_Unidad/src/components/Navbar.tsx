import { Link, useNavigate } from "react-router-dom";
import { useCart } from "../context/CartContext";
import { useAuth } from "../context/AuthContext";

interface NavbarProps {
  isCollapsed: boolean;
  onToggleCollapsed: () => void;
  onToggleMobile: () => void;
}

const Navbar = ({ isCollapsed, onToggleCollapsed, onToggleMobile }: NavbarProps) => {
  const { totalItems } = useCart();
  const { logout, userEmail } = useAuth();
  const navigate = useNavigate();

  const handleLogout = () => {
    logout();
    navigate("/login");
  };

  return (
    <header className="h-16 bg-white border-b border-slate-200 flex items-center justify-between px-4 md:px-8">
      <div className="flex items-center gap-3">
        {/* Botón para Móviles (☰) */}
        <button
          onClick={onToggleMobile}
          className="p-2 text-slate-600 hover:bg-slate-100 rounded-lg md:hidden text-xl"
        >
          ☰
        </button>

        {/* Botón para Escritorio (Flecha Colapsar) */}
        <button
          onClick={onToggleCollapsed}
          className="hidden md:flex p-2 text-slate-600 hover:bg-slate-100 rounded-lg transition text-xl items-center justify-center"
          title={isCollapsed ? "Expandir menú" : "Colapsar menú"}
        >
          {isCollapsed ? "➔" : "⬅"}
        </button>

        <h2 className="text-slate-600 font-medium text-base md:text-lg">
          Panel de Administración
        </h2>
      </div>

      <div className="flex items-center gap-4 md:gap-6">
        <Link to="/carrito" className="relative p-2 hover:bg-slate-100 rounded-full transition">
          <span className="text-xl">🛒</span>
          {totalItems > 0 && (
            <span className="absolute top-0 right-0 bg-indigo-600 text-white text-xs font-bold w-5 h-5 flex items-center justify-center rounded-full transform translate-x-1 -translate-y-1">
              {totalItems}
            </span>
          )}
        </Link>

        <div className="flex items-center gap-3">
          <span className="hidden sm:inline text-sm text-slate-500">{userEmail}</span>
          <div className="relative group cursor-pointer pb-2">
            <div className="w-8 h-8 rounded-full bg-slate-200 overflow-hidden border border-slate-300 flex items-center justify-center">
              <img
                src="https://fastly.picsum.photos/id/64/4326/2884.jpg?hmac=9_SzX666YRpR_fOyYStXpfSiJ_edO3ghlSRnH2w09Kg"
                alt="Avatar"
                className="w-full h-full object-cover"
              />
            </div>
            <div className="absolute right-0 top-full mt-1 w-36 bg-white border border-slate-200 rounded-lg shadow-lg opacity-0 invisible group-hover:opacity-100 group-hover:visible transition-all duration-200 z-50">
              <button
                onClick={handleLogout}
                className="w-full text-left px-4 py-2 text-sm text-red-600 font-semibold hover:bg-red-50 rounded-md transition-colors"
              >
                Cerrar Sesión
              </button>
            </div>
          </div>
        </div>
      </div>
    </header>
  );
};

export default Navbar;