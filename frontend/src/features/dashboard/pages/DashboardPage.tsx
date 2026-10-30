import { Link } from "react-router-dom";
import { useAuth } from "../../../app/AuthProvider";

const workAreas = [
  {
    title: "Pacientes",
    description: "Consulta y gestiona las fichas de la clínica.",
    href: "/patients",
    mark: "P",
  },
  {
    title: "Agenda",
    description: "Organiza citas por sillón y profesional.",
    href: "/scheduling",
    mark: "A",
  },
  {
    title: "Presupuestos",
    description: "Revisa tratamientos y aprobaciones.",
    href: "/billing",
    mark: "$",
  },
] as const;

export function DashboardPage() {
  const { user } = useAuth();
  const displayName = user?.email.split("@")[0] || "equipo";
  const currentDate = new Intl.DateTimeFormat("es-EC", {
    weekday: "long",
    day: "numeric",
    month: "long",
    timeZone: "America/Guayaquil",
  }).format(new Date());

  return (
    <section className="dashboard-page" aria-labelledby="dashboard-title">
      <header className="dashboard-welcome">
        <div>
          <p className="eyebrow">{currentDate}</p>
          <h1 id="dashboard-title">Hola, {displayName}</h1>
          <p>Accede a los módulos de trabajo de SIGO-Dental.</p>
        </div>
        <span className="welcome-mark" aria-hidden="true">S</span>
      </header>

      <section className="dashboard-section" aria-labelledby="work-areas-title">
        <div className="section-heading">
          <div>
            <h2 id="work-areas-title">Áreas de trabajo</h2>
            <p>Selecciona un módulo para continuar.</p>
          </div>
        </div>

        <div className="module-card-grid">
          {workAreas.map((area) => (
            <Link className="module-shortcut" key={area.href} to={area.href}>
              <span className="module-shortcut-mark" aria-hidden="true">{area.mark}</span>
              <span className="module-shortcut-copy">
                <span className="module-shortcut-title">{area.title}</span>
                <span className="module-shortcut-description">{area.description}</span>
              </span>
              <span className="module-shortcut-arrow" aria-hidden="true">→</span>
            </Link>
          ))}
        </div>
      </section>
    </section>
  );
}
