import { useEffect, useState, type FormEvent } from "react";
import { Link } from "react-router-dom";
import { getRegistrationEnabled, registerAccount } from "../api";
import { publicRegistrationSchema } from "../schemas";

export function PublicRegisterPage() {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [confirmation, setConfirmation] = useState("");
  const [errorMessage, setErrorMessage] = useState("");
  const [successMessage, setSuccessMessage] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [registrationEnabled, setRegistrationEnabled] = useState<boolean | null>(null);
  const [registrationCheckFailed, setRegistrationCheckFailed] = useState(false);
  const [registrationCompleted, setRegistrationCompleted] = useState(false);

  useEffect(() => {
    let active = true;
    getRegistrationEnabled()
      .then((enabled) => { if (active) setRegistrationEnabled(enabled); })
      .catch(() => {
        if (active) {
          setRegistrationEnabled(false);
          setRegistrationCheckFailed(true);
        }
      });
    return () => { active = false; };
  }, []);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setErrorMessage("");
    setSuccessMessage("");

    const parsed = publicRegistrationSchema.safeParse({ email, password });
    if (!parsed.success) {
      setErrorMessage(parsed.error.issues[0]?.message ?? "Revisa los datos ingresados.");
      return;
    }
    if (password !== confirmation) {
      setErrorMessage("Las contraseñas no coinciden.");
      return;
    }

    setIsSubmitting(true);
    try {
      await registerAccount(parsed.data);
      setPassword("");
      setConfirmation("");
      setSuccessMessage("Cuenta de administrador creada. Ya puedes iniciar sesión.");
      setRegistrationCompleted(true);
      setRegistrationEnabled(false);
    } catch (error) {
      setErrorMessage(error instanceof Error ? error.message : "No se pudo crear la cuenta.");
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <main className="login-page">
      <section className="login-card" aria-labelledby="register-title">
        <div className="brand-lockup" aria-label="SIGO-Dental by Carlos">
          <span className="brand-mark" aria-hidden="true">S</span>
          <span className="brand-name">SIGO<span>-Dental</span> by Carlos</span>
        </div>

        <div className="login-heading">
          <p className="eyebrow">Portal de gestión clínica</p>
          <h1 id="register-title">Crear cuenta</h1>
          <p>Crea la cuenta inicial de administración.</p>
        </div>

        {registrationEnabled === null && <div className="form-alert" role="status">Comprobando si el registro inicial está disponible…</div>}
        {registrationEnabled === false && (
          <div className={registrationCompleted ? "form-success" : "form-alert"} role="status">
            {registrationCompleted
              ? successMessage
              : registrationCheckFailed
                ? "No se pudo verificar el estado del registro. Inténtalo más tarde."
                : "El registro inicial ya se completó. Solicita una cuenta a un administrador."}
          </div>
        )}

        {registrationEnabled && <form className="login-form" onSubmit={handleSubmit} noValidate>
          {errorMessage && <div className="form-alert" role="alert">{errorMessage}</div>}
          {successMessage && <div className="form-success" role="status">{successMessage}</div>}

          <div className="form-field">
            <label htmlFor="register-email">Correo electrónico</label>
            <input
              id="register-email"
              name="email"
              type="email"
              autoComplete="email"
              value={email}
              onChange={(event) => setEmail(event.target.value)}
              disabled={isSubmitting}
              required
              maxLength={254}
            />
          </div>

          <div className="form-field">
            <label htmlFor="register-password">Contraseña</label>
            <input
              id="register-password"
              name="password"
              type="password"
              autoComplete="new-password"
              value={password}
              onChange={(event) => setPassword(event.target.value)}
              disabled={isSubmitting}
              required
            />
          </div>

          <div className="form-field">
            <label htmlFor="register-confirmation">Confirma la contraseña</label>
            <input
              id="register-confirmation"
              name="password_confirmation"
              type="password"
              autoComplete="new-password"
              value={confirmation}
              onChange={(event) => setConfirmation(event.target.value)}
              disabled={isSubmitting}
              required
            />
          </div>

          <p className="register-role-note">Esta cuenta recibirá el rol Administrador TI. El registro se cerrará después de crearla.</p>

          <button className="button-primary login-submit" type="submit" disabled={isSubmitting}>
            {isSubmitting ? "Creando cuenta…" : "Crear cuenta"}
          </button>
        </form>}

        <p className="login-footnote">¿Ya tienes cuenta? <Link className="text-link" to="/login">Inicia sesión</Link></p>
      </section>
    </main>
  );
}
