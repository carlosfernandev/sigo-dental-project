import { Link } from "react-router-dom";

type ModulePlaceholderPageProps = {
  title: string;
  description: string;
  detail: string;
};

export function ModulePlaceholderPage({ title, description, detail }: ModulePlaceholderPageProps) {
  return (
    <section className="module-page" aria-labelledby="module-title">
      <header className="page-heading">
        <p className="eyebrow">Módulo de Fase 1</p>
        <h1 id="module-title">{title}</h1>
        <p>{description}</p>
      </header>

      <div className="module-empty-state" role="status">
        <span className="empty-state-mark" aria-hidden="true">＋</span>
        <div>
          <h2>Vista en preparación</h2>
          <p>{detail}</p>
          <Link className="text-link" to="/dashboard">Volver al inicio</Link>
        </div>
      </div>
    </section>
  );
}
