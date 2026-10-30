import { useEffect, useMemo, useRef, useState } from "react";
import { Link, useLocation, useNavigate } from "react-router-dom";
import { ConfirmDialog } from "../../../shared/ui/ConfirmDialog";
import { DataTable, type DataTableColumn } from "../../../shared/ui/DataTable";
import { EntityDrawer } from "../../../shared/ui/EntityDrawer";
import { Toast } from "../../../shared/ui/Toast";
import { listPatients, savePatient, type PatientInput, type PatientSummary } from "../api";
import { PatientForm } from "../components/PatientForm";

const pageSize = 25;

export function PatientsPage() {
  const location = useLocation();
  const navigate = useNavigate();
  const isCreateRoute = location.pathname === "/patients/new";
  const [patients, setPatients] = useState<PatientSummary[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [searchInput, setSearchInput] = useState("");
  const [page, setPage] = useState(1);
  const [reload, setReload] = useState(0);
  const [saving, setSaving] = useState(false);
  const [operationError, setOperationError] = useState("");
  const [toast, setToast] = useState("");
  const [discardOpen, setDiscardOpen] = useState(false);
  const dirty = useRef(false);
  const searchRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    let active = true;
    setLoading(true); setError("");
    listPatients().then((rows) => { if (active) setPatients(rows); })
      .catch((cause: unknown) => { if (active) setError(cause instanceof Error ? cause.message : "No se pudo cargar la lista de pacientes."); })
      .finally(() => { if (active) setLoading(false); });
    return () => { active = false; };
  }, [reload]);

  useEffect(() => {
    const state = location.state as { notice?: string } | null;
    if (state?.notice) { setToast(state.notice); navigate(location.pathname, { replace: true, state: null }); }
  }, [location.pathname, location.state, navigate]);
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

  const filtered = useMemo(() => {
    const query = searchInput.trim().toLocaleLowerCase();
    return patients.filter((patient) => !query || `${patient.first_name} ${patient.last_name} ${patient.phone} ${patient.identification}`.toLocaleLowerCase().includes(query))
      .sort((a, b) => `${a.last_name} ${a.first_name}`.localeCompare(`${b.last_name} ${b.first_name}`));
  }, [patients, searchInput]);
  const currentPage = Math.min(page, Math.max(1, Math.ceil(filtered.length / pageSize)));
  const rows = filtered.slice((currentPage - 1) * pageSize, currentPage * pageSize);
  const columns = useMemo<DataTableColumn<PatientSummary>[]>(() => [
    { key: "name", header: "Nombre", render: (patient) => <Link className="table-row-primary-action" to={`/patients/${patient.id}`}>{patient.first_name} {patient.last_name}</Link> },
    { key: "document", header: "Cédula", render: (patient) => patient.identification || "No registrada" },
    { key: "phone", header: "Teléfono", render: (patient) => patient.phone },
    { key: "insurance", header: "Seguro / Convenio", render: (patient) => patient.insurance_plan || "—" },
    { key: "actions", header: "Acciones", render: (patient) => <Link className="secondary-button" to={`/patients/${patient.id}`}>Ver ficha</Link> },
  ], []);

  function closeCreate() {
    dirty.current = false; setDiscardOpen(false); setOperationError("");
    if (isCreateRoute) navigate("/patients");
  }
  function requestClose() {
    if (dirty.current) { setDiscardOpen(true); return; }
    closeCreate();
  }
  async function submit(input: PatientInput) {
    setSaving(true); setOperationError("");
    try {
      await savePatient(input);
      dirty.current = false;
      navigate("/patients", { state: { notice: "Paciente registrado correctamente." } });
    } catch (cause) {
      setOperationError(cause instanceof Error ? cause.message : "No se pudo registrar el paciente.");
    } finally { setSaving(false); }
  }
  function resetFilters() { setSearchInput(""); setPage(1); }

  return (
    <section className="module-page patients-page" aria-labelledby="patients-title">
      <header className="users-page-heading"><div><p className="eyebrow">Atención</p><h1 id="patients-title">Pacientes</h1><p className="patients-lead">Fichas únicas, contacto y antecedentes médicos.</p></div><Link className="button-primary users-create-button" to="/patients/new">Nuevo paciente</Link></header>
      <div className="users-toolbar"><div className="users-search-field"><label htmlFor="patients-search">Buscar pacientes</label><input ref={searchRef} id="patients-search" type="search" value={searchInput} onChange={(event) => { setSearchInput(event.target.value); setPage(1); }} placeholder="Nombre, teléfono o ID" /><span className="users-search-shortcut" aria-hidden="true">/</span></div></div>
      {!loading && !error && filtered.length === 0 ? <div className="table-state table-empty"><h2>{searchInput ? "Sin resultados" : "Aún no hay pacientes"}</h2><p>{searchInput ? "No encontramos pacientes con ese criterio. Ajusta la búsqueda." : "Registra un paciente para crear su ficha e historial médico."}</p>{searchInput ? <button className="secondary-button" type="button" onClick={resetFilters}>Limpiar búsqueda</button> : <Link className="button-primary" to="/patients/new">Nuevo paciente</Link>}</div> : <DataTable caption="Pacientes registrados en SIGO-Dental" columns={columns} rows={rows} rowKey={(row) => row.id} loading={loading} error={error} empty={false} hasFilters={Boolean(searchInput)} onRetry={() => setReload((value) => value + 1)} onResetFilters={resetFilters} emptyAction={<Link className="button-primary" to="/patients/new">Nuevo paciente</Link>} page={currentPage} pageSize={pageSize} total={filtered.length} onPageChange={setPage} sortBy="name" sortOrder="asc" onSort={() => undefined} />}
      <EntityDrawer open={isCreateRoute} title="Registrar paciente" onRequestClose={requestClose}>
        {operationError && <div className="form-alert" role="alert">{operationError}</div>}
        <PatientForm key="new-patient" isBusy={saving} onDirtyChange={(value) => { dirty.current = value; }} onCancel={requestClose} onSubmit={submit} />
      </EntityDrawer>
      <ConfirmDialog open={discardOpen} title="Descartar cambios" description="Hay cambios sin guardar. Si continúas, se perderán." confirmLabel="Descartar cambios" onCancel={() => setDiscardOpen(false)} onConfirm={closeCreate} />
      {toast && <Toast message={toast} onClose={() => setToast("")} />}
    </section>
  );
}
