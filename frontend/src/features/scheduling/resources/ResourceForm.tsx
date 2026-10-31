import { useState, type FormEvent } from "react";
import type { Chair, ChairInput, Dentist, DentistInput, Operatory, OperatoryInput, Resource, ResourceKind } from "./api";

type ResourceFormValue = OperatoryInput | ChairInput | DentistInput;
type Props = {
  kind: ResourceKind;
  initialData?: Resource;
  operatories: Operatory[];
  isBusy: boolean;
  onDirtyChange: (dirty: boolean) => void;
  onCancel: () => void;
  onSubmit: (value: ResourceFormValue) => void;
};

export function ResourceForm({ kind, initialData, operatories, isBusy, onDirtyChange, onCancel, onSubmit }: Props) {
  const operatory = kind === "operatories" ? initialData as Operatory | undefined : undefined;
  const chair = kind === "chairs" ? initialData as Chair | undefined : undefined;
  const dentist = kind === "dentists" ? initialData as Dentist | undefined : undefined;
  const [name, setName] = useState(kind === "dentists" ? "" : (operatory?.name ?? chair?.name ?? ""));
  const [description, setDescription] = useState(operatory?.description ?? "");
  const [operatoryID, setOperatoryID] = useState(chair?.operatory_id ?? "");
  const [firstName, setFirstName] = useState(dentist?.first_name ?? "");
  const [lastName, setLastName] = useState(dentist?.last_name ?? "");
  const [specialty, setSpecialty] = useState(dentist?.specialty ?? "");
  const [active, setActive] = useState(dentist?.active ?? true);

  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (kind === "operatories") onSubmit({ name: name.trim(), description: description.trim() });
    else if (kind === "chairs") onSubmit({ name: name.trim(), operatory_id: operatoryID });
    else onSubmit({ first_name: firstName.trim(), last_name: lastName.trim(), specialty: specialty.trim(), active });
  }

  return (
    <form className="resource-form" onChange={() => onDirtyChange(true)} onSubmit={submit}>
      {kind === "operatories" && <>
        <label className="form-field">Nombre del recinto<input required maxLength={100} autoFocus value={name} onChange={(event) => setName(event.target.value)} /></label>
        <label className="form-field">Descripción<textarea rows={3} maxLength={300} value={description} onChange={(event) => setDescription(event.target.value)} /></label>
      </>}
      {kind === "chairs" && <>
        <label className="form-field">Nombre del sillón<input required maxLength={100} autoFocus value={name} onChange={(event) => setName(event.target.value)} /></label>
        <label className="form-field">Recinto<select required value={operatoryID} onChange={(event) => setOperatoryID(event.target.value)}><option value="">Seleccionar recinto</option>{operatories.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}</select></label>
      </>}
      {kind === "dentists" && <>
        <label className="form-field">Nombres<input required maxLength={100} autoFocus value={firstName} onChange={(event) => setFirstName(event.target.value)} /></label>
        <label className="form-field">Apellidos<input required maxLength={100} value={lastName} onChange={(event) => setLastName(event.target.value)} /></label>
        <label className="form-field">Especialidad<input maxLength={120} value={specialty} onChange={(event) => setSpecialty(event.target.value)} /></label>
        <label className="resource-active-toggle"><input type="checkbox" checked={active} onChange={(event) => setActive(event.target.checked)} /> Profesional activo</label>
      </>}
      <div className="resource-form-actions"><button className="secondary-button" type="button" disabled={isBusy} onClick={onCancel}>Cancelar</button><button className="button-primary" type="submit" disabled={isBusy}>{isBusy ? "Guardando…" : initialData ? "Guardar cambios" : "Registrar"}</button></div>
    </form>
  );
}
