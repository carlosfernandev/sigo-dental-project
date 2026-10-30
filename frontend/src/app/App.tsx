import { lazy, Suspense } from "react";
import { Navigate, Outlet, Route, Routes, useLocation } from "react-router-dom";
import { useAuth } from "./AuthProvider";

const AppShell = lazy(() => import("../features/navigation/AppShell").then((module) => ({ default: module.AppShell })));
const LoginPage = lazy(() => import("../features/auth/pages/LoginPage").then((module) => ({ default: module.LoginPage })));
const PublicRegisterPage = lazy(() => import("../features/users/pages/PublicRegisterPage").then((module) => ({ default: module.PublicRegisterPage })));
const DashboardPage = lazy(() => import("../features/dashboard/pages/DashboardPage").then((module) => ({ default: module.DashboardPage })));
const PatientsPage = lazy(() => import("../features/patients/pages/PatientsPage").then((module) => ({ default: module.PatientsPage })));
const PatientDetailPage = lazy(() => import("../features/patients/pages/PatientDetailPage").then((module) => ({ default: module.PatientDetailPage })));
const SchedulingPage = lazy(() => import("../features/scheduling/pages/SchedulingPage").then((module) => ({ default: module.SchedulingPage })));
const BillingPage = lazy(() => import("../features/billing/pages/BillingPage").then((module) => ({ default: module.BillingPage })));
const UsersPage = lazy(() => import("../features/users/pages/UsersPage").then((module) => ({ default: module.UsersPage })));

function PublicLayout() {
  const { user, isLoading } = useAuth();
  if (isLoading) return <SessionLoading />;
  if (user) return <Navigate to="/dashboard" replace />;
  return <Outlet />;
}

function PrivateLayout() {
  const { user, isLoading } = useAuth();
  const location = useLocation();
  if (isLoading) return <SessionLoading />;
  if (!user) return <Navigate to="/login" replace state={{ from: location }} />;
  return <Outlet />;
}

function AdminOnlyLayout() {
  const { user } = useAuth();
  return user?.role === "admin_ti" ? <Outlet /> : <Navigate to="/dashboard" replace />;
}

function SessionLoading() {
  return (
    <main className="session-loading" role="status" aria-live="polite">
      <span className="visually-hidden">Verificando sesión…</span>
      <div className="session-skeleton" aria-hidden="true">
        <span className="skeleton-block skeleton-mark" />
        <span className="skeleton-block skeleton-heading" />
        <span className="skeleton-block skeleton-line" />
      </div>
    </main>
  );
}

function HomeRedirect() {
  const { user, isLoading } = useAuth();
  if (isLoading) return <SessionLoading />;
  return <Navigate to={user ? "/dashboard" : "/login"} replace />;
}

function NotFoundRedirect() {
  return <Navigate to="/" replace />;
}

export function App() {
  return (
    <Routes>
      <Route element={<PublicLayout />}>
        <Route path="/login" element={<Suspense fallback={<SessionLoading />}><LoginPage /></Suspense>} />
        <Route path="/register" element={<Suspense fallback={<SessionLoading />}><PublicRegisterPage /></Suspense>} />
      </Route>
      <Route element={<PrivateLayout />}>
        <Route element={<Suspense fallback={<SessionLoading />}><AppShell /></Suspense>}>
          <Route path="/dashboard" element={<Suspense fallback={<SessionLoading />}><DashboardPage /></Suspense>} />
          <Route path="/patients" element={<Suspense fallback={<SessionLoading />}><PatientsPage /></Suspense>} />
          <Route path="/patients/new" element={<Suspense fallback={<SessionLoading />}><PatientsPage /></Suspense>} />
          <Route path="/patients/:id" element={<Suspense fallback={<SessionLoading />}><PatientDetailPage /></Suspense>} />
          <Route path="/scheduling" element={<Suspense fallback={<SessionLoading />}><SchedulingPage /></Suspense>} />
          <Route path="/billing" element={<Suspense fallback={<SessionLoading />}><BillingPage /></Suspense>} />
          <Route element={<AdminOnlyLayout />}>
            <Route path="/users" element={<Suspense fallback={<SessionLoading />}><UsersPage /></Suspense>} />
            <Route path="/users/new" element={<Suspense fallback={<SessionLoading />}><UsersPage /></Suspense>} />
            <Route path="/admin/users" element={<Navigate to="/users" replace />} />
          </Route>
        </Route>
      </Route>
      <Route path="/" element={<HomeRedirect />} />
      <Route path="*" element={<NotFoundRedirect />} />
    </Routes>
  );
}
