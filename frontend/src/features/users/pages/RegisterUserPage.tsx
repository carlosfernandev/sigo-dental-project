import { useState, type FormEvent } from "react";
import { Link, useNavigate } from "react-router-dom";
import { createUser } from "../api";
import { createUserSchema, userRoles } from "../schemas";

const roleLabels: Record<(typeof userRoles)[number], string> = {
  admin_ti: "Administrador TI",
  recepcionista_cajero: "Recepcionista / Cajero",
  odontologo: "Odontólogo",
  asistente_dental: "Asistente dental",
  auditor_contador: "Auditor / Contador",
};

const emptyForm = { email: "", password: "", role_id: "" };

export function RegisterUserPage() {
  const navigate = useNavigate();
  const [form, setForm] = useState(emptyForm);
  const [errorMessage, setErrorMessage] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setErrorMessage("");

    const parsed = createUserSchema.safeParse(form);
    if (!parsed.success) {
      setErrorMessage(parsed.error.issues[0]?.message ?? "Revisa los datos ingresados.");
      return;
    }

    setIsSubmitting(true);
    try {
      await createUser(parsed.data);
      navigate("/users", { state: { notice: "Usuario creado correctamente." } });
    } catch (error) {
      setErrorMessage(error instanceof Error ? error.message : "No se pudo crear el usuario.");
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <section className="module-page" aria-labelledby="register-user-title">
      <header className="page-heading">
        <p className="eyebrow">Administración</p>
        <h1 id="register-user-title">Registrar usuario</h1>
        <p>Crea una cuenta para un integrante del personal de la clínica.</p>
        <Link className="text-link" to="/users">Volver a usuarios</Link>
      </header>

      <form className="register-user-card" onSubmit={handleSubmit} noValidate>
        {errorMessage && <div className="form-alert" role="alert">{errorMessage}</div>}

        <div className="form-field">
          <label htmlFor="new-user-email">Correo electrónico</label>
          <input
            id="new-user-email"
            name="email"
            type="email"
            autoComplete="email"
            value={form.email}
            onChange={(event) => setForm({ ...form, email: event.target.value })}
            disabled={isSubmitting}
            required
            maxLength={254}
          />
        </div>

        <div className="form-field">
          <label htmlFor="new-user-password">Contraseña temporal</label>
          <input
            id="new-user-password"
            name="password"
            type="password"
            autoComplete="new-password"
            value={form.password}
            onChange={(event) => setForm({ ...form, password: event.target.value })}
            disabled={isSubmitting}
            required
            maxLength={72}
          />
        </div>

        <div className="form-field">
          <label htmlFor="new-user-role">Rol</label>
          <select
            id="new-user-role"
            name="role_id"
            value={form.role_id}
            onChange={(event) => setForm({ ...form, role_id: event.target.value })}
            disabled={isSubmitting}
            required
          >
            <option value="">Selecciona un rol</option>
            {userRoles.map((role) => <option key={role} value={role}>{roleLabels[role]}</option>)}
          </select>
        </div>

        <button className="button-primary" type="submit" disabled={isSubmitting}>
          {isSubmitting ? "Creando usuario…" : "Crear Usuario"}
        </button>
      </form>
    </section>
  );
}
