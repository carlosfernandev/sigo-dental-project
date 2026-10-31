import { NavLink, Outlet, useLocation } from "react-router-dom";
import { useAuth } from "../../app/AuthProvider";

const navigationItems = [
  { to: "/dashboard", label: "Inicio", icon: "home" },
  { to: "/patients", label: "Pacientes", icon: "patients" },
  { to: "/scheduling", label: "Agenda", icon: "calendar" },
  { to: "/billing", label: "Presupuestos", icon: "billing" },
  { to: "/users", label: "Usuarios", icon: "users" },
  { to: "/admin/operatories", label: "Recursos", icon: "resources" },
] as const;

const roleLabels: Record<string, string> = {
  admin_ti: "Administración de TI",
  recepcionista_cajero: "Recepción y caja",
  odontologo: "Odontólogo/a",
  asistente_dental: "Asistente dental",
  auditor_contador: "Auditoría y contabilidad",
};

function NavigationIcon({ name }: { name: string }) {
  const paths: Record<string, string> = {
    home: "M3 10.5 12 3l9 7.5M5.5 9v11h13V9M9 20v-6h6v6",
    patients: "M16 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2M10 11a4 4 0 1 0 0-8 4 4 0 0 0 0 8ZM20 8v6m3-3h-6",
    users: "M16 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2M10 11a4 4 0 1 0 0-8 4 4 0 0 0 0 8ZM18 8h5m-2.5-2.5v5",
    calendar: "M8 3v4m8-4v4M4 9h16M5 5h14a1 1 0 0 1 1 1v13H4V6a1 1 0 0 1 1-1Zm3 8h3v3H8z",
    billing: "M5 3h14v18l-2-1-2 1-3-1-3 1-2-1-2 1V3Zm4 5h6m-6 4h6m-6 4h4",
    resources: "M4 5h16v4H4zM6 9v10m12-10v10M4 19h16M9 13h6",
  };

  return (
    <svg className="nav-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.7" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      <path d={paths[name] ?? paths.home} />
    </svg>
  );
}

function roleLabel(role: string) {
  return roleLabels[role] ?? role.replaceAll("_", " ");
}

export function AppShell() {
  const { user, logout } = useAuth();
  const { pathname } = useLocation();
  const displayName = user?.email.split("@")[0] || "Usuario";
  const currentSection = pathname.startsWith("/admin/") ? "Recursos" : navigationItems.find((item) => pathname === item.to || pathname.startsWith(`${item.to}/`))?.label ?? "Inicio";
  const canManageUsers = user?.role === "admin_ti";
  const canUseScheduling = ["admin_ti", "recepcionista_cajero", "odontologo", "asistente_dental"].includes(user?.role ?? "");

  return (
    <div className="app-shell">
      <aside className="sidebar" aria-label="Navegación principal">
        <div className="sidebar-brand">
          <span className="brand-mark" aria-hidden="true">S</span>
          <span className="sidebar-brand-name">SIGO<span>-Dental</span></span>
        </div>

        <nav className="sidebar-navigation" aria-label="Módulos">
          <p className="sidebar-section-label">Módulos</p>
          {navigationItems.filter((item) => (item.to !== "/users" || canManageUsers) && (item.to !== "/admin/operatories" || canManageUsers) && (item.to !== "/scheduling" || canUseScheduling)).map((item) => (
            <NavLink key={item.to} to={item.to} aria-label={item.label} title={item.label} className={({ isActive }) => `sidebar-link${isActive || (item.to === "/admin/operatories" && pathname.startsWith("/admin/")) ? " is-active" : ""}`}>
              <NavigationIcon name={item.icon} />
              <span>{item.label}</span>
            </NavLink>
          ))}
        </nav>

        <div className="sidebar-account" aria-label={`Cuenta: ${displayName}, ${user ? roleLabel(user.role) : ""}`} title={`${displayName} · ${user ? roleLabel(user.role) : ""}`}>
          <div className="account-avatar" aria-hidden="true">{displayName.slice(0, 1).toUpperCase()}</div>
          <div className="account-copy">
            <p className="account-name" title={user?.email}>{displayName}</p>
            <p className="account-role">{user ? roleLabel(user.role) : ""}</p>
          </div>
          <button className="logout-button" type="button" onClick={logout} aria-label="Cerrar sesión" title="Cerrar sesión">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
              <path d="M10 17l5-5-5-5m5 5H3m9-9h6a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2h-6" />
            </svg>
          </button>
        </div>
      </aside>

      <div className="workspace">
        <header className="topbar">
          <div>
            <p className="topbar-overline">SIGO-Dental / {currentSection}</p>
            <p className="topbar-context">Sistema de gestión clínica</p>
          </div>
        </header>
        <main className="workspace-content">
          <Outlet />
        </main>
      </div>
    </div>
  );
}
