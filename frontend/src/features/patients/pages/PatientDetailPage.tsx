import { useEffect, useRef, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { ConfirmDialog } from "../../../shared/ui/ConfirmDialog";
import { EntityDrawer } from "../../../shared/ui/EntityDrawer";
import { Toast } from "../../../shared/ui/Toast";
import { getPatient, savePatient, type Patient, type PatientInput } from "../api";
import { PatientForm } from "../components/PatientForm";

export function PatientDetailPage() {
  const { id = "" } = useParams();
  const [patient, setPatient] = useState<Patient | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [editing, setEditing] = useState(false);
  const [saving, setSaving] = useState(false);
  const [operationError, setOperationError] = useState("");
  const [toast, setToast] = useState("");
  const [discardOpen, setDiscardOpen] = useState(false);
  const dirty = useRef(false);

  useEffect(() => {
    let active = true;
    setLoading(true); setError("");
    getPatient(id).then((data) => { if (active) setPatient(data); })
      .catch((cause: unknown) => { if (active) setError(cause instanceof Error ? cause.message : "No se pudo cargar la ficha del paciente."); })
      .finally(() => { if (active) setLoading(false); });
    return () => { active = false; };
  }, [id]);

  function requestClose() {
    if (dirty.current) setDiscardOpen(true);
    else setEditing(false);
  }
  function closeEdit() { dirty.current = false; setDiscardOpen(false); setEditing(false); }
  async function submit(input: PatientInput) {
    setSaving(true); setOperationError("");
    try {
      const updated = await savePatient(input, id);
      setPatient(updated); closeEdit(); setToast("Ficha actualizada correctamente.");
    } catch (cause) {
      setOperationError(cause instanceof Error ? cause.message : "No se pudo actualizar la ficha.");
    } finally { setSaving(false); }
  }

  if (loading) return <main className="module-page"><div className="table-state table-loading" role="status">Cargando ficha del paciente…</div></main>;
  if (error || !patient) return <main className="module-page"><div className="table-state table-error" role="alert"><p>{error || "No encontramos este paciente."}</p><Link className="secondary-button" to="/patients">Volver a pacientes</Link></div></main>;

  const history = patient.medical_history;
  return (
    <section className="module-page patients-page patient-detail-page" aria-labelledby="patient-detail-title">
      <Link className="patients-back-link" to="/patients">← Volver a pacientes</Link>
      <header className="users-page-heading"><div><p className="eyebrow">Ficha del paciente</p><h1 id="patient-detail-title">{patient.first_name} {patient.last_name}</h1><p className="patients-lead">Fecha de nacimiento: {patient.date_of_birth}</p></div><button className="button-primary" type="button" onClick={() => { setOperationError(""); setEditing(true); }}>Editar ficha</button></header>
      {patient.critical_alerts.length > 0 && <section className="patient-critical-banner" aria-label="Alertas médicas críticas"><h2>Alertas médicas</h2><ul>{patient.critical_alerts.map((alert) => <li key={alert}><span aria-hidden="true">!</span> {alert}</li>)}</ul></section>}
      <div className="patient-detail-grid">
        <section className="patient-detail-card"><h2>Datos personales</h2><dl>
          <div><dt>Cédula</dt><dd>{patient.identification || "No registrada"}</dd></div><div><dt>Correo electrónico</dt><dd>{patient.email || "No registrado"}</dd></div><div><dt>Teléfono</dt><dd>{patient.phone}</dd></div><div><dt>Dirección</dt><dd>{patient.address || "No registrada"}</dd></div><div><dt>Seguro / Convenio</dt><dd>{patient.insurance_plan || "No registrado"}</dd></div><div><dt>Contacto de emergencia</dt><dd>{patient.emergency_contact_name} · {patient.emergency_contact_phone}</dd></div>
          {patient.legal_guardian && <div><dt>Tutor legal</dt><dd>{patient.legal_guardian.name} · {patient.legal_guardian.identification} · {patient.legal_guardian.phone}</dd></div>}
        </dl></section>
        <section className="patient-detail-card"><h2>Historial médico</h2><MedicalList title="Enfermedades sistémicas" values={history.systemic_diseases} /><MedicalList title="Alergias" values={history.allergies.map((item) => `${item.name}${item.is_critical ? " · Crítica" : ""}`)} /><MedicalList title="Antecedentes quirúrgicos" values={history.surgical_history} /><MedicalList title="Coagulopatías" values={history.coagulopathies} /></section>
      </div>
      <EntityDrawer open={editing} title="Editar ficha del paciente" onRequestClose={requestClose}>
        {operationError && <div className="form-alert" role="alert">{operationError}</div>}
        <PatientForm key={patient.updated_at ?? patient.id} initialData={patient} isBusy={saving} onDirtyChange={(value) => { dirty.current = value; }} onCancel={requestClose} onSubmit={submit} />
      </EntityDrawer>
      <ConfirmDialog open={discardOpen} title="Descartar cambios" description="Hay cambios sin guardar. Si continúas, se perderán." confirmLabel="Descartar cambios" onCancel={() => setDiscardOpen(false)} onConfirm={closeEdit} />
      {toast && <Toast message={toast} onClose={() => setToast("")} />}
    </section>
  );
}

function MedicalList({ title, values }: { title: string; values: string[] }) {
  return <div className="patient-medical-list"><h3>{title}</h3>{values.length ? <ul>{values.map((value) => <li key={value}>{value}</li>)}</ul> : <p>Sin registros</p>}</div>;
}
