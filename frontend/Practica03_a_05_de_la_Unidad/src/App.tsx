// src/App.tsx
import { type ReactNode, lazy, Suspense } from 'react';
import { BrowserRouter, Routes, Route, Navigate, Outlet } from 'react-router-dom';
import Layout from './components/Layout';
import Storefront from './components/Storefront';
import Catalogo from './components/Catalogo';
import DetalleProducto from './components/DetalleProducto';
import Carrito from './components/Carrito';
import Checkout from './components/Checkout';
import Confirmacion from './components/Confirmacion';
import Login from './components/Login';
import ServiceWorkerUpdate from './components/ServiceWorkerUpdate';
import NetworkStatusBanner from './components/NetworkStatusBanner';
import { AdminProductos, AdminPedidos, MisPedidos, AdminLiquidaciones, AdminUsuarios } from './components/Backoffice';
import { CartProvider } from './context/CartContext';
import { AuthProvider, useAuth } from './context/AuthContext';

// Lazy loading de componentes pesados
const Dashboard = lazy(() => import('./components/Dashboard'));
const MiRed = lazy(() => import('./components/MiRed'));
const Registro = lazy(() => import('./components/Registro'));

// Componente para proteger las rutas privadas: exige estar autenticado
const ProtectedRoute = () => {
  const { isAuthenticated } = useAuth();

  // Si no está autenticado, lo enviamos al login
  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  // Si está autenticado, renderiza las rutas hijas (Outlet)
  return <Outlet />;
};

// Componente para rutas exclusivas del administrador (Tema 5: roles)
const AdminRoute = () => {
  const { isAuthenticated, user } = useAuth();

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  // Si el usuario no es admin, lo enviamos a la tienda (su vista por defecto)
  if (user?.rol !== 'admin') {
    return <Navigate to="/tienda" replace />;
  }

  return <Outlet />;
};

const RoleHome = () => {
  const { user } = useAuth();
  if (user?.rol !== 'admin') return <Navigate to="/tienda" replace />;
  return <Suspense fallback={<div className="p-8 text-center">Cargando...</div>}><Dashboard /></Suspense>;
};

// Proveedor intermedio del carrito: al cambiar de usuario se remonta con key,
// de modo que cada cuenta lea y persista SU propio carrito en localStorage.
const CartBoundary = ({ children }: { children: ReactNode }) => {
  const { user } = useAuth();
  return (
    <CartProvider key={user?.email ?? "anonimo"}>
      {children}
    </CartProvider>
  );
};

function App() {
  return (
    <AuthProvider> {/* Proveedor de Autenticación */}
      <CartBoundary> {/* Proveedor del Carrito (uno por usuario) */}
        <NetworkStatusBanner />
        <BrowserRouter>
          <Routes>
            {/* Ruta pública */}
            <Route path="/login" element={<Login />} />
            <Route path="/registro" element={<Suspense fallback={<div className="p-8 text-center">Cargando...</div>}><Registro /></Suspense>} />

            {/* Rutas protegidas (requieren login) */}
            <Route element={<ProtectedRoute />}>
              <Route path="/" element={<Layout />}>
                <Route index element={<RoleHome />} />
                {/* Rutas administrativas */}
                <Route element={<AdminRoute />}>
                  <Route path="admin/productos" element={<AdminProductos />} />
                  <Route path="admin/pedidos" element={<AdminPedidos />} />
                  <Route path="admin/usuarios" element={<AdminUsuarios />} />
                  <Route path="admin/liquidaciones" element={<AdminLiquidaciones />} />
                </Route>

                <Route path="mi-red" element={<Suspense fallback={<div className="p-8 text-center">Cargando...</div>}><MiRed /></Suspense>} />
                <Route path="mis-pedidos" element={<MisPedidos />} />

                {/* Ambos roles: flujo de compra en línea */}
                <Route path="tienda" element={<Storefront />} />
                <Route path="catalogo" element={<Catalogo />} />
                <Route path="producto/:id" element={<DetalleProducto />} />
                <Route path="carrito" element={<Carrito />} />
                <Route path="checkout" element={<Checkout />} />
                <Route path="confirmacion" element={<Confirmacion />} />
              </Route>
            </Route>

            {/* Ruta comodín para capturar 404 y redirigir */}
            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        </BrowserRouter>
        <ServiceWorkerUpdate />
      </CartBoundary>
    </AuthProvider>
  );
}

export default App;
