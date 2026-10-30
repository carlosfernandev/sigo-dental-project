import { useEffect, useState, type FormEvent } from "react";
import type { Patient, PatientInput } from "../api";

type Props = {
  initialData?: Patient;
  isBusy: boolean;
  onDirtyChange: (dirty: boolean) => void;
  onCancel: () => void;
  onSubmit: (input: PatientInput) => void;
};

const blank: PatientInput = {
  first_name: "", last_name: "", identification: "", date_of_birth: "", email: "", phone: "", address: "",
  emergency_contact_name: "", emergency_contact_phone: "", legal_guardian: null, insurance_plan: "",
  medical_history: { systemic_diseases: [], allergies: [], surgical_history: [], coagulopathies: [] },
};
const insuranceOptions = [
  "Particular",
  "Seguro Nacional Dental",
  "Seguro Vida Dental",
  "Convenio Empresa Acme",
  "Convenio Colegio de Odontólogos",
  "Prepagada DentalMax",
];

function dateValue(value?: string) { return value?.slice(0, 10) ?? ""; }
function lines(value: string) { return value.split("\n").map((item) => item.trim()).filter(Boolean); }
function isMinor(date: string) {
  if (!date) return false;
  const birth = new Date(`${date}T00:00:00`);
  const cutoff = new Date(); cutoff.setFullYear(cutoff.getFullYear() - 18);
  return birth > cutoff;
}

export function PatientForm({ initialData, isBusy, onDirtyChange, onCancel, onSubmit }: Props) {
  const [form, setForm] = useState<PatientInput>(() => initialData ? {
    first_name: initialData.first_name, last_name: initialData.last_name, identification: initialData.identification, date_of_birth: dateValue(initialData.date_of_birth),
    email: initialData.email, phone: initialData.phone, address: initialData.address,
    emergency_contact_name: initialData.emergency_contact_name, emergency_contact_phone: initialData.emergency_contact_phone,
    legal_guardian: initialData.legal_guardian ?? null, insurance_plan: initialData.insurance_plan,
    medical_history: initialData.medical_history,
  } : blank);
  const [criticalAllergies, setCriticalAllergies] = useState(() => initialData?.medical_history.allergies.some((item) => item.is_critical) ?? false);
  const minor = isMinor(form.date_of_birth);

  useEffect(() => () => onDirtyChange(false), [onDirtyChange]);

  function update<K extends keyof PatientInput>(key: K, value: PatientInput[K]) {
    setForm((current) => ({ ...current, [key]: value }));
  }
  function updateHistory(key: "systemic_diseases" | "surgical_history" | "coagulopathies", value: string) {
    setForm((current) => ({ ...current, medical_history: { ...current.medical_history, [key]: lines(value) } }));
  }
  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    onSubmit({
      ...form,
      legal_guardian: minor ? form.legal_guardian : null,
      medical_history: {
        ...form.medical_history,
        allergies: form.medical_history.allergies.map((item) => ({ ...item, is_critical: criticalAllergies })),
      },
    });
  }

  const guardian = form.legal_guardian ?? { name: "", identification: "", phone: "" };
  return (
    <form className="patient-form" onSubmit={submit} onChange={() => onDirtyChange(true)}>
      <fieldset className="patient-form-section">
        <legend>Datos personales</legend>
        <div className="patient-form-grid">
          <label className="form-field">Nombres *<input required maxLength={100} value={form.first_name} onChange={(e) => update("first_name", e.target.value)} /></label>
          <label className="form-field">Apellidos (*)<input required maxLength={100} value={form.last_name} onChange={(e) => update("last_name", e.target.value)} /></label>
          <label className="form-field">Cédula ecuatoriana (*)<input required inputMode="numeric" pattern="[0-9]{10}" maxLength={10} minLength={10} autoComplete="off" value={form.identification} onChange={(e) => update("identification", e.target.value.replace(/\D/g, "").slice(0, 10))} /></label>
          <label className="form-field">Fecha de nacimiento (*)<input required type="date" value={form.date_of_birth} onChange={(e) => update("date_of_birth", e.target.value)} /></label>
          <label className="form-field">Correo electrónico<input type="email" maxLength={254} value={form.email} onChange={(e) => update("email", e.target.value)} /></label>
          <label className="form-field">Teléfono (*)<input required maxLength={32} value={form.phone} onChange={(e) => update("phone", e.target.value)} /></label>
          <label className="form-field">Seguro o convenio<select value={form.insurance_plan} onChange={(e) => update("insurance_plan", e.target.value)}><option value="">Seleccionar</option>{insuranceOptions.map((option) => <option key={option} value={option}>{option}</option>)}</select></label>
          <label className="form-field patient-form-wide">Dirección<textarea maxLength={300} rows={2} value={form.address} onChange={(e) => update("address", e.target.value)} /></label>
        </div>
        <h3>Contacto de emergencia</h3>
        <div className="patient-form-grid">
          <label className="form-field">Nombre (*)<input required maxLength={120} value={form.emergency_contact_name} onChange={(e) => update("emergency_contact_name", e.target.value)} /></label>
          <label className="form-field">Teléfono (*)<input required maxLength={32} value={form.emergency_contact_phone} onChange={(e) => update("emergency_contact_phone", e.target.value)} /></label>
        </div>
        {minor && <fieldset className="patient-form-section patient-form-guardian"><legend>Tutor legal (obligatorio para menores)</legend><div className="patient-form-grid">
          <label className="form-field">Nombre<input required maxLength={120} value={guardian.name} onChange={(e) => update("legal_guardian", { ...guardian, name: e.target.value })} /></label>
          <label className="form-field">Identificación<input required maxLength={32} value={guardian.identification} onChange={(e) => update("legal_guardian", { ...guardian, identification: e.target.value })} /></label>
          <label className="form-field">Teléfono<input required maxLength={32} value={guardian.phone} onChange={(e) => update("legal_guardian", { ...guardian, phone: e.target.value })} /></label>
        </div></fieldset>}
      </fieldset>

      <fieldset className="patient-form-section">
        <legend>Historial médico</legend>
        <p className="patient-form-note">Se registra junto con la ficha. Escribe un antecedente por línea.</p>
        <div className="patient-form-grid">
          <label className="form-field">Enfermedades sistémicas<textarea rows={3} value={form.medical_history.systemic_diseases.join("\n")} onChange={(e) => updateHistory("systemic_diseases", e.target.value)} /></label>
          <label className="form-field">Alergias (una por línea)<textarea rows={3} value={form.medical_history.allergies.map((item) => item.name).join("\n")} onChange={(e) => setForm((current) => ({ ...current, medical_history: { ...current.medical_history, allergies: lines(e.target.value).map((name) => ({ name, is_critical: criticalAllergies })) } }))} /></label>
          <label className="patient-critical-toggle"><input type="checkbox" checked={criticalAllergies} onChange={(e) => setCriticalAllergies(e.target.checked)} /> Marcar las alergias como críticas</label>
          <label className="form-field">Antecedentes quirúrgicos<textarea rows={3} value={form.medical_history.surgical_history.join("\n")} onChange={(e) => updateHistory("surgical_history", e.target.value)} /></label>
          <label className="form-field">Coagulopatías (una por línea)<textarea rows={3} value={form.medical_history.coagulopathies.join("\n")} onChange={(e) => updateHistory("coagulopathies", e.target.value)} /></label>
        </div>
      </fieldset>
      <div className="patient-form-actions"><button className="secondary-button" type="button" onClick={onCancel} disabled={isBusy}>Cancelar</button><button className="button-primary" type="submit" disabled={isBusy}>{isBusy ? "Guardando…" : initialData ? "Guardar cambios" : "Registrar paciente"}</button></div>
    </form>
  );
}
