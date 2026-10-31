import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { Link, useLocation } from "react-router-dom";
import { ConfirmDialog } from "../../../shared/ui/ConfirmDialog";
import { DataTable, type DataTableColumn } from "../../../shared/ui/DataTable";
import { EntityDrawer } from "../../../shared/ui/EntityDrawer";
import { StatusChip } from "../../../shared/ui/StatusChip";
import { Toast } from "../../../shared/ui/Toast";
import { listChairs, listDentists, listOperatories, saveChair, saveDentist, saveOperatory, type Chair, type Dentist, type Operatory, type Resource, type ResourceKind } from "./api";
import { ResourceForm } from "./ResourceForm";

const tabs: { kind: ResourceKind; label: string; path: string; title: string; singular: string }[] = [
  { kind: "operatories", label: "Recintos", path: "/admin/operatories", title: "Recintos", singular: "recinto" },
  { kind: "chairs", label: "Sillones", path: "/admin/chairs", title: "Sillones", singular: "sillón" },
  { kind: "dentists", label: "Dentistas", path: "/admin/dentists", title: "Dentistas", singular: "dentista" },
];
const pageSize = 25;

export function ResourcesPage() {
  const location = useLocation();
  const current = tabs.find((tab) => tab.path === location.pathname) ?? tabs[0];
  const [operatories, setOperatories] = useState<Operatory[]>([]);
  const [chairs, setChairs] = useState<Chair[]>([]);
  const [dentists, setDentists] = useState<Dentist[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [reload, setReload] = useState(0);
  const [searchInput, setSearchInput] = useState("");
  const [page, setPage] = useState(1);
  const [sortBy, setSortBy] = useState("name");
  const [sortOrder, setSortOrder] = useState<"asc" | "desc">("asc");
  const [editing, setEditing] = useState<Resource | null>(null);
  const [creating, setCreating] = useState(false);
  const [saving, setSaving] = useState(false);
  const [operationError, setOperationError] = useState("");
  const [toast, setToast] = useState("");
  const [discardOpen, setDiscardOpen] = useState(false);
  const dirty = useRef(false);
  const searchRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    let active = true;
    setLoading(true); setError("");
    Promise.all([listOperatories(), listChairs(), listDentists()])
      .then(([nextOperatories, nextChairs, nextDentists]) => {
        if (!active) return;
        setOperatories(nextOperatories); setChairs(nextChairs); setDentists(nextDentists);
      })
      .catch((cause: unknown) => { if (active) setError(cause instanceof Error ? cause.message : "No se pudieron cargar los recursos de agenda."); })
      .finally(() => { if (active) setLoading(false); });
    return () => { active = false; };
  }, [current.kind, reload]);

  useEffect(() => {
    const handleKey = (event: KeyboardEvent) => {
      const target = event.target;
      if (event.key === "/" && !(target instanceof HTMLElement && (target.isContentEditable || ["INPUT", "TEXTAREA", "SELECT"].includes(target.tagName)))) {
        event.preventDefault(); searchRef.current?.focus();
      }
    };
    window.addEventListener("keydown", handleKey);
    return () => window.removeEventListener("keydown", handleKey);
  }, []);

  const sourceRows: Resource[] = current.kind === "operatories" ? operatories : current.kind === "chairs" ? chairs : dentists;
  const operatoriesByID = useMemo(() => new Map(operatories.map((item) => [item.id, item.name])), [operatories]);
  const filteredRows = useMemo(() => {
    const query = searchInput.trim().toLocaleLowerCase();
    const filtered = sourceRows.filter((row) => {
      if (!query) return true;
      if ("operatory_id" in row) return `${row.name} ${operatoriesByID.get(row.operatory_id) ?? ""}`.toLocaleLowerCase().includes(query);
      if ("first_name" in row) return `${row.first_name} ${row.last_name} ${row.specialty} ${row.active ? "activo" : "inactivo"}`.toLocaleLowerCase().includes(query);
      return `${row.name} ${row.description}`.toLocaleLowerCase().includes(query);
    });
    const value = (row: Resource) => {
      if ("first_name" in row) return sortBy === "specialty" ? row.specialty : sortBy === "active" ? String(row.active) : `${row.last_name} ${row.first_name}`;
      if ("operatory_id" in row) return sortBy === "operatory" ? operatoriesByID.get(row.operatory_id) ?? "" : row.name;
      return sortBy === "description" ? row.description : row.name;
    };
    return filtered.sort((left, right) => {
      const comparison = value(left).localeCompare(value(right), "es", { sensitivity: "base" });
      return sortOrder === "asc" ? comparison : -comparison;
    });
  }, [sourceRows, searchInput, operatoriesByID, sortBy, sortOrder]);
  const currentPage = Math.min(page, Math.max(1, Math.ceil(filteredRows.length / pageSize)));
  const rows = filteredRows.slice((currentPage - 1) * pageSize, currentPage * pageSize);

  const columns = useMemo<DataTableColumn<Resource>[]>(() => {
    if (current.kind === "operatories") return [
      { key: "name", header: "Recinto", sortKey: "name", render: (row) => "name" in row ? row.name : "—" },
      { key: "description", header: "Descripción", sortKey: "description", render: (row) => "description" in row && row.description ? row.description : "—" },
      { key: "actions", header: "Acciones", render: (row) => <button className="secondary-button" type="button" onClick={() => openEdit(row)}>Editar</button> },
    ];
    if (current.kind === "chairs") return [
      { key: "name", header: "Sillón", sortKey: "name", render: (row) => "name" in row ? row.name : "—" },
      { key: "operatory", header: "Recinto", sortKey: "operatory", render: (row) => "operatory_id" in row ? operatoriesByID.get(row.operatory_id) ?? "Recinto no disponible" : "—" },
      { key: "actions", header: "Acciones", render: (row) => <button className="secondary-button" type="button" onClick={() => openEdit(row)}>Editar</button> },
    ];
    return [
      { key: "name", header: "Dentista", sortKey: "name", render: (row) => "first_name" in row ? `${row.first_name} ${row.last_name}` : row.name },
      { key: "specialty", header: "Especialidad", sortKey: "specialty", render: (row) => "specialty" in row && row.specialty ? row.specialty : "—" },
      { key: "active", header: "Estado", sortKey: "active", render: (row) => "active" in row ? <StatusChip label={row.active ? "Activo" : "Inactivo"} tone={row.active ? "positive" : "negative"} /> : "—" },
      { key: "actions", header: "Acciones", render: (row) => <button className="secondary-button" type="button" onClick={() => openEdit(row)}>Editar</button> },
    ];
  }, [current.kind, operatoriesByID]);

  function openEdit(row: Resource) { setOperationError(""); dirty.current = false; setEditing(row); setCreating(false); }
  function openCreate() { setOperationError(""); dirty.current = false; setEditing(null); setCreating(true); }
  function closeDrawer() { dirty.current = false; setDiscardOpen(false); setEditing(null); setCreating(false); }
  function requestClose() { if (dirty.current) setDiscardOpen(true); else closeDrawer(); }

  const save = useCallback(async (input: { name: string; description: string } | { name: string; operatory_id: string } | { first_name: string; last_name: string; specialty: string; active: boolean }) => {
    setSaving(true); setOperationError("");
    try {
      if (current.kind === "operatories") await saveOperatory(input as { name: string; description: string }, editing?.id);
      else if (current.kind === "chairs") await saveChair(input as { name: string; operatory_id: string }, editing?.id);
      else await saveDentist(input as { first_name: string; last_name: string; specialty: string; active: boolean }, editing?.id);
      closeDrawer(); setReload((value) => value + 1); setToast(`${current.singular[0].toUpperCase()}${current.singular.slice(1)} ${editing ? "actualizado" : "registrado"} correctamente.`);
    } catch (cause) {
      setOperationError(cause instanceof Error ? cause.message : `No se pudo guardar el ${current.singular}.`);
    } finally { setSaving(false); }
  }, [current, editing]);

  function resetFilters() { setSearchInput(""); setPage(1); }
  const drawerOpen = creating || editing !== null;
  const initialData = editing ?? undefined;
  const emptyMessage = searchInput ? "No hay resultados" : `Aún no hay ${current.title.toLocaleLowerCase()}`;

  return (
    <section className="module-page resources-page" aria-labelledby="resources-title">
      <header className="users-page-heading resources-heading"><div><p className="eyebrow">Configuración de agenda</p><h1 id="resources-title">Recursos médicos</h1><p className="resources-lead">Administra los recintos, sillones y profesionales que se usan al agendar.</p></div></header>
      <nav className="resource-tabs" aria-label="Tipos de recursos">
        {tabs.map((tab) => <Link key={tab.kind} to={tab.path} className={`resource-tab${current.kind === tab.kind ? " is-active" : ""}`} aria-current={current.kind === tab.kind ? "page" : undefined}>{tab.label}</Link>)}
      </nav>
      <div className="resources-toolbar"><div className="users-search-field"><label htmlFor="resource-search">Buscar {current.title.toLocaleLowerCase()}</label><input ref={searchRef} id="resource-search" type="search" value={searchInput} onChange={(event) => { setSearchInput(event.target.value); setPage(1); }} placeholder={current.kind === "chairs" ? "Nombre o recinto" : current.kind === "dentists" ? "Nombre, especialidad o estado" : "Nombre o descripción"} /></div><button className="button-primary" type="button" onClick={openCreate}>Nuevo {current.singular}</button></div>
      {!loading && !error && filteredRows.length === 0 ? <div className="table-state table-empty"><h2>{emptyMessage}</h2><p>{searchInput ? "Prueba con otro término o limpia la búsqueda." : `Registra un ${current.singular} para que esté disponible al agendar.`}</p>{searchInput ? <button className="secondary-button" type="button" onClick={resetFilters}>Limpiar búsqueda</button> : <button className="button-primary" type="button" onClick={openCreate}>Nuevo {current.singular}</button>}</div> : <DataTable caption={`${current.title} configurados para la agenda`} columns={columns} rows={rows} rowKey={(row) => row.id} loading={loading} error={error} empty={false} hasFilters={Boolean(searchInput)} onRetry={() => setReload((value) => value + 1)} onResetFilters={resetFilters} emptyAction={null} page={currentPage} pageSize={pageSize} total={filteredRows.length} onPageChange={setPage} sortBy={sortBy} sortOrder={sortOrder} onSort={(key, order) => { setSortBy(key); setSortOrder(order); }} />}
      <EntityDrawer open={drawerOpen} title={`${editing ? "Editar" : "Nuevo"} ${current.singular}`} onRequestClose={requestClose}>
        {operationError && <div className="form-alert" role="alert">{operationError}</div>}
        <ResourceForm key={`${current.kind}-${editing?.id ?? "new"}`} kind={current.kind} initialData={initialData} operatories={operatories} isBusy={saving} onDirtyChange={(value) => { dirty.current = value; }} onCancel={requestClose} onSubmit={save} />
      </EntityDrawer>
      <ConfirmDialog open={discardOpen} title="Descartar cambios" description="Hay cambios sin guardar. Si continúas, se perderán." confirmLabel="Descartar cambios" onCancel={() => setDiscardOpen(false)} onConfirm={closeDrawer} />
      {toast && <Toast message={toast} onClose={() => setToast("")} />}
    </section>
  );
}
