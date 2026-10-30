import { useEffect, useState, type FormEvent } from "react";
import { FormField } from "../../../shared/ui/FormField";
import type { ManagedUser } from "../api";
import { createUserSchema, updateUserFormSchema, userRoles } from "../schemas";

const roleLabels: Record<(typeof userRoles)[number], string> = {
  admin_ti: "Administrador TI",
  recepcionista_cajero: "Recepcionista / Cajero",
  odontologo: "Odontólogo",
  asistente_dental: "Asistente dental",
  auditor_contador: "Auditor / Contador",
};

type UserFormValues = { email: string; password: string; role_id: string };

type UserFormProps = {
  initialData?: ManagedUser;
  isBusy: boolean;
  onDirtyChange: (dirty: boolean) => void;
  onCancel: () => void;
  onSubmit: (values: UserFormValues, isEditing: boolean) => Promise<void>;
};

export function UserForm({ initialData, isBusy, onDirtyChange, onCancel, onSubmit }: UserFormProps) {
  const [email, setEmail] = useState(initialData?.email ?? "");
  const [password, setPassword] = useState("");
  const [role, setRole] = useState(initialData?.role ?? "");
  const [errors, setErrors] = useState<Record<string, string>>({});
  const isEditing = Boolean(initialData);

  useEffect(() => {
    const dirty = isEditing
      ? role !== initialData?.role
      : email !== "" || password !== "" || role !== "";
    onDirtyChange(dirty);
  }, [email, initialData, isEditing, onDirtyChange, password, role]);

  function validateField(field: "email" | "password" | "role_id") {
    const values = { email, password, role_id: role };
    const result = createUserSchema.safeParse(values);
    const issue = result.success ? undefined : result.error.issues.find((item) => item.path[0] === field);
    setErrors((current) => ({ ...current, [field]: issue?.message ?? "" }));
  }

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const result = isEditing
      ? updateUserFormSchema.safeParse({ role_id: role })
      : createUserSchema.safeParse({ email, password, role_id: role });
    if (!result.success) {
      const nextErrors: Record<string, string> = {};
      for (const issue of result.error.issues) {
        const key = String(issue.path[0] ?? "form");
        if (!nextErrors[key]) nextErrors[key] = issue.message;
      }
      setErrors(nextErrors);
      const firstInvalid = ["email", "password", "role_id"].find((field) => nextErrors[field]);
      if (firstInvalid) document.getElementById(`user-${firstInvalid}`)?.focus();
      return;
    }

    setErrors({});
    await onSubmit({ email, password, role_id: role }, isEditing);
  }

  return (
    <form className="entity-form" onSubmit={submit} noValidate>
      <FormField id="user-email" label="Correo electrónico" required={!isEditing} error={errors.email} hint={isEditing ? "El correo registrado no se puede modificar." : undefined}>
        <input id="user-email" type="email" autoComplete="email" value={email} onChange={(event) => setEmail(event.target.value)} onBlur={() => !isEditing && validateField("email")} disabled={isBusy || isEditing} maxLength={254} required={!isEditing} aria-invalid={Boolean(errors.email)} />
      </FormField>

      {!isEditing && (
        <FormField id="user-password" label="Contraseña temporal" required error={errors.password}>
          <input id="user-password" type="password" autoComplete="new-password" value={password} onChange={(event) => setPassword(event.target.value)} onBlur={() => validateField("password")} disabled={isBusy} maxLength={72} required aria-invalid={Boolean(errors.password)} />
        </FormField>
      )}

      <FormField id="user-role_id" label="Rol" required error={errors.role_id}>
        <select id="user-role_id" value={role} onChange={(event) => setRole(event.target.value)} onBlur={() => validateField("role_id")} disabled={isBusy} required aria-invalid={Boolean(errors.role_id)}>
          <option value="">Selecciona un rol</option>
          {userRoles.map((userRole) => <option key={userRole} value={userRole}>{roleLabels[userRole]}</option>)}
        </select>
      </FormField>

      <footer className="entity-form-actions">
        <button className="button-primary" type="submit" disabled={isBusy}>{isBusy ? "Guardando…" : "Guardar usuario"}</button>
        <button className="secondary-button" type="button" onClick={onCancel} disabled={isBusy}>Cancelar</button>
      </footer>
    </form>
  );
}
