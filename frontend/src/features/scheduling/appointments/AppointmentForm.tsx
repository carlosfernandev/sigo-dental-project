import { useMemo, useState, type FormEvent } from "react";
import { Link } from "react-router-dom";
import type { PatientSummary } from "../../patients/api";
import type { Chair, Dentist, Operatory } from "../resources/api";
import type { AppointmentInput } from "./api";

type Props = {
  selectedDate: string;
  patients: PatientSummary[];
  chairs: Chair[];
  dentists: Dentist[];
  operatories: Operatory[];
  isBusy: boolean;
  onDirtyChange: (dirty: boolean) => void;
  onCancel: () => void;
  onSubmit: (input: AppointmentInput) => void;
};

function dateTimeInput(date: Date) {
  const local = new Date(date.getTime() - date.getTimezoneOffset() * 60_000);
  return local.toISOString().slice(0, 16);
}

function initialRange(selectedDate: string) {
  const start = new Date(`${selectedDate}T09:00:00`);
  const today = new Date();
  const todayValue = `${today.getFullYear()}-${String(today.getMonth() + 1).padStart(2, "0")}-${String(today.getDate()).padStart(2, "0")}`;
  if (selectedDate === todayValue) {
    const now = new Date();
    now.setMinutes(now.getMinutes() + 10);
    now.setMinutes(Math.ceil(now.getMinutes() / 15) * 15, 0, 0);
    start.setTime(now.getTime());
  }
  const end = new Date(start.getTime() + 30 * 60_000);
  return { start: dateTimeInput(start), end: dateTimeInput(end) };
}

export function AppointmentForm({ selectedDate, patients, chairs, dentists, operatories, isBusy, onDirtyChange, onCancel, onSubmit }: Props) {
  const range = useMemo(() => initialRange(selectedDate), [selectedDate]);
  const [patientQuery, setPatientQuery] = useState("");
  const [patientID, setPatientID] = useState("");
  const [dentistID, setDentistID] = useState("");
  const [chairID, setChairID] = useState("");
  const [startTime, setStartTime] = useState(range.start);
  const [endTime, setEndTime] = useState(range.end);
  const [notes, setNotes] = useState("");
  const [validationError, setValidationError] = useState("");
  const availablePatients = patients.filter((patient) => `${patient.first_name} ${patient.last_name} ${patient.identification}`.toLocaleLowerCase().includes(patientQuery.trim().toLocaleLowerCase()));
  const activeDentists = dentists.filter((dentist) => dentist.active);

  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const start = new Date(startTime);
    const end = new Date(endTime);
    const duration = end.getTime() - start.getTime();
    if (duration < 15 * 60_000 || duration > 8 * 60 * 60_000) {
      setValidationError("La cita debe durar entre 15 minutos y 8 horas.");
      return;
    }
    if (start.getTime() < Date.now()) {
      setValidationError("La fecha y hora de inicio debe ser futura.");
      return;
    }
    setValidationError("");
    onSubmit({
      patient_id: patientID, dentist_id: dentistID, chair_id: chairID,
      start_time: start.toISOString(), end_time: end.toISOString(), state: "scheduled", notes: notes.trim(),
    });
  }

  return (
    <form className="appointment-form" onChange={() => onDirtyChange(true)} onSubmit={submit}>
      {validationError && <div className="form-alert" role="alert">{validationError}</div>}
      <div className="appointment-form-grid">
        <section className="appointment-form-column appointment-form-column--patients" aria-labelledby="appointment-patient-heading">
          <h3 id="appointment-patient-heading">Paciente</h3>
          <label className="form-field">Buscar<input type="search" value={patientQuery} onChange={(event) => { setPatientQuery(event.target.value); setPatientID(""); }} placeholder="Nombre o cédula" /></label>
          <div className="appointment-option-list" role="radiogroup" aria-labelledby="appointment-patient-heading">
            {availablePatients.map((patient) => <label className={`appointment-option${patientID === patient.id ? " is-selected" : ""}`} key={patient.id}><input required type="radio" name="appointment-patient" value={patient.id} checked={patientID === patient.id} onChange={() => setPatientID(patient.id)} /><span><strong>{patient.first_name} {patient.last_name}</strong>{patient.identification && <small>{patient.identification}</small>}</span></label>)}
            {availablePatients.length === 0 && <p className="appointment-option-empty">{patients.length ? "No hay coincidencias." : <>No hay pacientes. <Link to="/patients">Ir a pacientes</Link></>}</p>}
          </div>
        </section>
        <section className="appointment-form-column appointment-form-column--options" aria-labelledby="appointment-dentist-heading">
          <h3 id="appointment-dentist-heading">Odontólogo</h3>
          <div className="appointment-option-list" role="radiogroup" aria-labelledby="appointment-dentist-heading">
            {activeDentists.map((dentist) => <label className={`appointment-option${dentistID === dentist.id ? " is-selected" : ""}`} key={dentist.id}><input required type="radio" name="appointment-dentist" value={dentist.id} checked={dentistID === dentist.id} onChange={() => setDentistID(dentist.id)} /><span><strong>{dentist.first_name} {dentist.last_name}</strong>{dentist.specialty && <small>{dentist.specialty}</small>}</span></label>)}
            {activeDentists.length === 0 && <p className="appointment-option-empty">No hay odontólogos activos.</p>}
          </div>
        </section>
        <section className="appointment-form-column appointment-form-column--options" aria-labelledby="appointment-chair-heading">
          <h3 id="appointment-chair-heading">Sillón</h3>
          <div className="appointment-option-list" role="radiogroup" aria-labelledby="appointment-chair-heading">
            {chairs.map((chair) => <label className={`appointment-option${chairID === chair.id ? " is-selected" : ""}`} key={chair.id}><input required type="radio" name="appointment-chair" value={chair.id} checked={chairID === chair.id} onChange={() => setChairID(chair.id)} /><span><strong>{chair.name}</strong><small>{operatories.find((operatory) => operatory.id === chair.operatory_id)?.name ?? "Recinto"}</small></span></label>)}
            {chairs.length === 0 && <p className="appointment-option-empty">No hay sillones disponibles.</p>}
          </div>
        </section>
        <section className="appointment-form-column appointment-form-column--schedule" aria-labelledby="appointment-time-heading">
          <h3 id="appointment-time-heading">Horario y nota</h3>
          <label className="form-field">Inicio (obligatorio)<input type="datetime-local" required min={dateTimeInput(new Date())} value={startTime} onChange={(event) => setStartTime(event.target.value)} /></label>
          <label className="form-field">Fin (obligatorio)<input type="datetime-local" required min={startTime} value={endTime} onChange={(event) => setEndTime(event.target.value)} /></label>
          <label className="form-field">Nota<textarea rows={3} maxLength={2000} value={notes} onChange={(event) => setNotes(event.target.value)} /></label>
        </section>
      </div>
      <p className="appointment-form-hint">La cita se registrará como programada. El sistema impide reservar el mismo sillón en horarios superpuestos.</p>
      <div className="appointment-form-actions"><button className="secondary-button" type="button" onClick={onCancel} disabled={isBusy}>Cancelar</button><button className="button-primary" type="submit" disabled={isBusy || patients.length === 0 || activeDentists.length === 0 || chairs.length === 0}>{isBusy ? "Guardando…" : "Agendar cita"}</button></div>
    </form>
  );
}
