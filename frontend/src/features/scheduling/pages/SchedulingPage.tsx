import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useAuth } from "../../../app/AuthProvider";
import { listPatients, type PatientSummary } from "../../patients/api";
import { ConfirmDialog } from "../../../shared/ui/ConfirmDialog";
import { DataTable, type DataTableColumn } from "../../../shared/ui/DataTable";
import { EntityDrawer } from "../../../shared/ui/EntityDrawer";
import { StatusChip } from "../../../shared/ui/StatusChip";
import { listChairs, listDentists, listOperatories, type Chair, type Dentist, type Operatory } from "../resources/api";
import { AppointmentForm } from "../appointments/AppointmentForm";
import { createAppointment, getAppointment, listAppointments, updateAppointment, type Appointment, type AppointmentInput, type AppointmentState, type AppointmentSummary } from "../appointments/api";

const labels: Record<AppointmentState, string> = {
  scheduled: "Programada", confirmed: "Confirmada", checked_in: "En sala de espera",
  in_progress: "En atención", completed: "Atendida", cancelled: "Cancelada", no_show: "No asistió",
};
const transitions: Record<AppointmentState, AppointmentState[]> = {
  scheduled: ["confirmed", "cancelled", "no_show"],
  confirmed: ["checked_in", "cancelled", "no_show"],
  checked_in: ["in_progress", "cancelled"],
  in_progress: ["completed"],
  completed: [], cancelled: [], no_show: [],
};
const schedulerRoles = ["admin_ti", "recepcionista_cajero", "odontologo", "asistente_dental"];
const pageSize = 25;

function localDateValue(date: Date) {
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, "0")}-${String(date.getDate()).padStart(2, "0")}`;
}
function dayRange(date: string) {
  const start = new Date(`${date}T00:00:00`);
  const end = new Date(`${date}T23:59:59.999`);
  return { startDate: start.toISOString(), endDate: end.toISOString() };
}
function displayTime(value: string) {
  return new Intl.DateTimeFormat("es-EC", { hour: "2-digit", minute: "2-digit" }).format(new Date(value));
}
function statusTone(state: AppointmentState): "positive" | "negative" | "neutral" {
  if (state === "confirmed" || state === "completed") return "positive";
  if (state === "cancelled" || state === "no_show") return "negative";
  return "neutral";
}

export function SchedulingPage() {
  const { user } = useAuth();
  const [selectedDate, setSelectedDate] = useState(() => localDateValue(new Date()));
  const [selectedOperatory, setSelectedOperatory] = useState("");
  const [selectedChair, setSelectedChair] = useState("");
  const [selectedDentist, setSelectedDentist] = useState("");
  const [operatories, setOperatories] = useState<Operatory[]>([]);
  const [chairs, setChairs] = useState<Chair[]>([]);
  const [dentists, setDentists] = useState<Dentist[]>([]);
  const [patients, setPatients] = useState<PatientSummary[]>([]);
  const [optionsLoading, setOptionsLoading] = useState(true);
  const [optionsError, setOptionsError] = useState("");
  const [appointments, setAppointments] = useState<AppointmentSummary[]>([]);
  const [appointmentsLoading, setAppointmentsLoading] = useState(true);
  const [appointmentsError, setAppointmentsError] = useState("");
  const [refresh, setRefresh] = useState(0);
  const [optionsRefresh, setOptionsRefresh] = useState(0);
  const [page, setPage] = useState(1);
  const [creating, setCreating] = useState(false);
  const [selectedAppointment, setSelectedAppointment] = useState<Appointment | null>(null);
  const [selectedSummary, setSelectedSummary] = useState<AppointmentSummary | null>(null);
  const [detailLoading, setDetailLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [operationError, setOperationError] = useState("");
  const [targetState, setTargetState] = useState<AppointmentState | "">("");
  const [toast, setToast] = useState<{ message: string; isError: boolean } | null>(null);
  const [discardOpen, setDiscardOpen] = useState(false);
  const dirty = useRef(false);
  const detailRequest = useRef(0);

  useEffect(() => {
    let active = true;
    setOptionsLoading(true); setOptionsError("");
    Promise.all([listOperatories(), listChairs(), listDentists(), listPatients()])
      .then(([nextOperatories, nextChairs, nextDentists, nextPatients]) => {
        if (!active) return;
        setOperatories(nextOperatories); setChairs(nextChairs); setDentists(nextDentists); setPatients(nextPatients);
      })
      .catch((cause: unknown) => { if (active) { setOptionsError(cause instanceof Error ? cause.message : "No se pudieron cargar los filtros de agenda."); setAppointmentsLoading(false); } })
      .finally(() => { if (active) setOptionsLoading(false); });
    return () => { active = false; };
  }, [optionsRefresh]);

  useEffect(() => {
    if (optionsLoading || optionsError) return;
    let active = true;
    setAppointmentsLoading(true); setAppointmentsError("");
    listAppointments({ ...dayRange(selectedDate), chairID: selectedChair || undefined, dentistID: selectedDentist || undefined })
      .then((rows) => { if (active) setAppointments(rows); })
      .catch((cause: unknown) => { if (active) setAppointmentsError(cause instanceof Error ? cause.message : "No se pudieron cargar las citas."); })
      .finally(() => { if (active) setAppointmentsLoading(false); });
    return () => { active = false; };
  }, [optionsLoading, optionsError, selectedDate, selectedChair, selectedDentist, selectedOperatory, refresh]);

  useEffect(() => {
    if (!toast) return;
    const timeout = window.setTimeout(() => setToast(null), 8000);
    return () => window.clearTimeout(timeout);
  }, [toast]);

  const filteredAppointments = useMemo(() => selectedOperatory
    ? appointments.filter((appointment) => chairs.some((chair) => chair.id === appointment.chair_id && chair.operatory_id === selectedOperatory))
    : appointments, [appointments, chairs, selectedOperatory]);
  const currentPage = Math.min(page, Math.max(1, Math.ceil(filteredAppointments.length / pageSize)));
  const rows = filteredAppointments.slice((currentPage - 1) * pageSize, currentPage * pageSize);
  const visibleChairs = selectedOperatory ? chairs.filter((chair) => chair.operatory_id === selectedOperatory) : chairs;

  const columns = useMemo<DataTableColumn<AppointmentSummary>[]>(() => [
    { key: "time", header: "Hora", render: (row) => <span className="appointment-time">{displayTime(row.start_time)}–{displayTime(row.end_time)}</span> },
    { key: "patient", header: "Paciente", render: (row) => row.patient_name },
    { key: "dentist", header: "Odontólogo", render: (row) => row.dentist_name },
    { key: "chair", header: "Sillón", render: (row) => row.chair_name },
    { key: "state", header: "Estado", render: (row) => <StatusChip label={labels[row.state]} tone={statusTone(row.state)} /> },
    { key: "actions", header: "Acciones", render: (row) => <button className="secondary-button" type="button" onClick={() => void openDetails(row)}>Ver detalle</button> },
  ], []);

  async function openDetails(summary: AppointmentSummary) {
    const requestID = ++detailRequest.current;
    setOperationError(""); setSelectedSummary(summary); setSelectedAppointment(null); setDetailLoading(true); setCreating(false);
    try {
      const appointment = await getAppointment(summary.id);
      if (detailRequest.current !== requestID) return;
      setSelectedAppointment(appointment);
      const available = transitions[appointment.state].filter((state) => state !== "no_show" || Date.now() >= new Date(appointment.start_time).getTime());
      setTargetState(available[0] ?? "");
    } catch (cause) {
      if (detailRequest.current === requestID) setToast({ message: cause instanceof Error ? cause.message : "No se pudo cargar el detalle de la cita.", isError: true });
    } finally { if (detailRequest.current === requestID) setDetailLoading(false); }
  }

  function closeDrawer() {
    detailRequest.current += 1;
    dirty.current = false; setDiscardOpen(false); setCreating(false); setSelectedAppointment(null); setSelectedSummary(null); setDetailLoading(false); setTargetState(""); setOperationError("");
  }
  function requestClose() {
    if (dirty.current) setDiscardOpen(true);
    else closeDrawer();
  }

  async function create(input: AppointmentInput) {
    setSaving(true); setOperationError("");
    try {
      await createAppointment(input);
      closeDrawer(); setRefresh((value) => value + 1); setToast({ message: "Cita agendada correctamente.", isError: false });
    } catch (cause) {
      const message = cause instanceof Error ? cause.message : "No se pudo agendar la cita.";
      setOperationError(message); setToast({ message, isError: true });
    } finally { setSaving(false); }
  }

  async function updateState() {
    if (!selectedAppointment || !targetState) return;
    const input: AppointmentInput = {
      chair_id: selectedAppointment.chair_id, dentist_id: selectedAppointment.dentist_id,
      patient_id: selectedAppointment.patient_id, start_time: selectedAppointment.start_time,
      end_time: selectedAppointment.end_time, notes: selectedAppointment.notes, state: targetState,
    };
    setSaving(true); setOperationError("");
    try {
      const updated = await updateAppointment(selectedAppointment.id, input);
      setSelectedAppointment(updated); setTargetState(""); setRefresh((value) => value + 1);
      setToast({ message: `Estado actualizado: ${labels[updated.state]}.`, isError: false });
    } catch (cause) {
      const message = cause instanceof Error ? cause.message : "No se pudo cambiar el estado de la cita.";
      setOperationError(message); setToast({ message, isError: true });
    } finally { setSaving(false); }
  }

  const clearFilters = useCallback(() => {
    setSelectedOperatory(""); setSelectedChair(""); setSelectedDentist(""); setPage(1);
  }, []);
  const drawerOpen = creating || selectedAppointment !== null || detailLoading;
  const dateLabel = new Intl.DateTimeFormat("es-EC", { dateStyle: "full" }).format(new Date(`${selectedDate}T12:00:00`));
  const noRecords = !appointmentsLoading && !appointmentsError && filteredAppointments.length === 0;
  const hasFilters = Boolean(selectedOperatory || selectedChair || selectedDentist);

  if (!schedulerRoles.includes(user?.role ?? "")) return <section className="module-page"><div className="table-state table-error" role="alert"><p>Tu rol no tiene acceso a la agenda.</p></div></section>;

  return (
    <section className="module-page appointments-page" aria-labelledby="scheduling-title">
      <header className="users-page-heading"><div><p className="eyebrow">Agenda clínica</p><h1 id="scheduling-title">Citas</h1><p className="appointments-date-label">{dateLabel}</p></div><button className="button-primary" type="button" onClick={() => { setOperationError(""); setSelectedAppointment(null); setCreating(true); dirty.current = false; }} disabled={optionsLoading || Boolean(optionsError)}>Nueva cita</button></header>
      {optionsError && <div className="form-alert" role="alert"><p>{optionsError}</p><button className="secondary-button" type="button" onClick={() => setOptionsRefresh((value) => value + 1)}>Reintentar</button></div>}
      <div className="appointment-filters" aria-label="Filtros de citas">
        <label className="form-field">Día<input type="date" required value={selectedDate} onChange={(event) => { if (event.target.value) setSelectedDate(event.target.value); setPage(1); }} /></label>
        <label className="form-field">Recinto<select value={selectedOperatory} onChange={(event) => { setSelectedOperatory(event.target.value); setSelectedChair(""); setPage(1); }} disabled={optionsLoading}><option value="">Todos los recintos</option>{operatories.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}</select></label>
        <label className="form-field">Sillón<select value={selectedChair} onChange={(event) => { setSelectedChair(event.target.value); setPage(1); }} disabled={optionsLoading}><option value="">Todos los sillones</option>{visibleChairs.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}</select></label>
        <label className="form-field">Odontólogo<select value={selectedDentist} onChange={(event) => { setSelectedDentist(event.target.value); setPage(1); }} disabled={optionsLoading}><option value="">Todos los odontólogos</option>{dentists.map((item) => <option key={item.id} value={item.id}>{item.first_name} {item.last_name}</option>)}</select></label>
        {hasFilters && <button className="secondary-button appointment-clear-filters" type="button" onClick={clearFilters}>Limpiar filtros</button>}
      </div>
      {noRecords ? <div className="table-state table-empty"><h2>{hasFilters ? "Sin citas para estos filtros" : "No hay citas para este día"}</h2><p>{hasFilters ? "Prueba con otro recinto, sillón u odontólogo." : "Puedes agendar una cita para este día desde aquí."}</p>{hasFilters ? <button className="secondary-button" type="button" onClick={clearFilters}>Limpiar filtros</button> : <button className="button-primary" type="button" onClick={() => setCreating(true)}>Nueva cita</button>}</div> : <DataTable caption="Citas del día ordenadas por hora" columns={columns} rows={rows} rowKey={(row) => row.id} loading={appointmentsLoading || optionsLoading} error={appointmentsError} empty={false} hasFilters={hasFilters} onRetry={() => setRefresh((value) => value + 1)} onResetFilters={clearFilters} emptyAction={null} page={currentPage} pageSize={pageSize} total={filteredAppointments.length} onPageChange={setPage} sortBy="start_time" sortOrder="asc" onSort={() => undefined} />}

      <EntityDrawer open={drawerOpen} title={creating ? "Agendar cita" : detailLoading ? "Cargando cita" : "Detalle de la cita"} wide={creating} centered={creating} onRequestClose={requestClose}>
        {operationError && <div className="form-alert" role="alert">{operationError}</div>}
        {creating && <AppointmentForm key={`create-${selectedDate}`} selectedDate={selectedDate} patients={patients} chairs={chairs} dentists={dentists} operatories={operatories} isBusy={saving} onDirtyChange={(value) => { dirty.current = value; }} onCancel={requestClose} onSubmit={create} />}
        {detailLoading && <div className="table-state table-loading" role="status">Cargando detalle…</div>}
        {selectedAppointment && selectedSummary && <AppointmentDetails appointment={selectedAppointment} summary={selectedSummary} targetState={targetState} isBusy={saving} onTargetStateChange={setTargetState} onUpdateState={() => void updateState()} />}
      </EntityDrawer>
      <ConfirmDialog open={discardOpen} title="Descartar cambios" description="Hay cambios sin guardar. Si continúas, se perderán." confirmLabel="Descartar cambios" onCancel={() => setDiscardOpen(false)} onConfirm={closeDrawer} />
      {toast && <div className={`app-toast${toast.isError ? " appointment-toast-error" : ""}`} role={toast.isError ? "alert" : "status"}><span>{toast.message}</span><button className="toast-close" type="button" aria-label="Cerrar notificación" onClick={() => setToast(null)}>×</button></div>}
    </section>
  );
}

function AppointmentDetails({ appointment, summary, targetState, isBusy, onTargetStateChange, onUpdateState }: {
  appointment: Appointment;
  summary: AppointmentSummary;
  targetState: AppointmentState | "";
  isBusy: boolean;
  onTargetStateChange: (state: AppointmentState | "") => void;
  onUpdateState: () => void;
}) {
  const availableStates = transitions[appointment.state].filter((state) => state !== "no_show" || Date.now() >= new Date(appointment.start_time).getTime());
  return (
    <div className="appointment-details">
      <StatusChip label={labels[appointment.state]} tone={statusTone(appointment.state)} />
      <dl>
        <div><dt>Fecha</dt><dd>{new Intl.DateTimeFormat("es-EC", { dateStyle: "long" }).format(new Date(appointment.start_time))}</dd></div>
        <div><dt>Horario</dt><dd>{displayTime(appointment.start_time)}–{displayTime(appointment.end_time)}</dd></div>
        <div><dt>Paciente</dt><dd>{summary.patient_name}</dd></div>
        <div><dt>Odontólogo</dt><dd>{summary.dentist_name}</dd></div>
        <div><dt>Sillón</dt><dd>{summary.chair_name}</dd></div>
        {appointment.notes && <div><dt>Notas</dt><dd>{appointment.notes}</dd></div>}
      </dl>
      {availableStates.length > 0 ? <div className="appointment-state-change"><label className="form-field" htmlFor="appointment-next-state">Cambiar estado<select id="appointment-next-state" value={targetState} onChange={(event) => onTargetStateChange(event.target.value as AppointmentState | "")}><option value="">Seleccionar estado</option>{availableStates.map((state) => <option key={state} value={state}>{labels[state]}</option>)}</select></label><button className="button-primary" type="button" onClick={onUpdateState} disabled={isBusy || !targetState}>{isBusy ? "Actualizando…" : "Guardar estado"}</button></div> : <p className="appointment-form-hint">Esta cita ya está en un estado final y no admite más cambios.</p>}
    </div>
  );
}
