import { useEffect, useState, type FormEvent } from "react";
import { Link, useNavigate } from "react-router-dom";
import { useAuth } from "../../../app/AuthProvider";
import { getRegistrationEnabled } from "../../users/api";
import { loginInputSchema } from "../schemas";

export function LoginPage() {
  const navigate = useNavigate();
  const { login } = useAuth();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [errorMessage, setErrorMessage] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [registrationEnabled, setRegistrationEnabled] = useState(false);

  useEffect(() => {
    let active = true;
    getRegistrationEnabled()
      .then((enabled) => { if (active) setRegistrationEnabled(enabled); })
      .catch(() => { if (active) setRegistrationEnabled(false); });
    return () => { active = false; };
  }, []);
  const localCredentialValues = import.meta.env.DEV
    ? {
      email: import.meta.env.VITE_LOCAL_LOGIN_EMAIL,
      password: import.meta.env.VITE_LOCAL_LOGIN_PASSWORD,
    }
    : undefined;
  const localCredentials = localCredentialValues?.email && localCredentialValues.password
    ? { email: localCredentialValues.email, password: localCredentialValues.password }
    : null;

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setErrorMessage("");
    const parsed = loginInputSchema.safeParse({ email, password });
    if (!parsed.success) {
      setErrorMessage(parsed.error.issues[0]?.message ?? "Revisa los datos ingresados.");
      return;
    }

    setIsSubmitting(true);
    try {
      await login(parsed.data);
      navigate("/dashboard", { replace: true });
    } catch (error) {
      setErrorMessage(error instanceof Error ? error.message : "No se pudo iniciar sesión.");
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <main className="login-page">
      <section className="login-card" aria-labelledby="login-title">
        <div className="brand-lockup" aria-label="SIGO-Dental">
          <span className="brand-mark" aria-hidden="true">S</span>
          <span className="brand-name">SIGO<span>-Dental</span></span>
        </div>

        <div className="login-heading">
          <p className="eyebrow">Portal de gestión clínica</p>
          <h1 id="login-title">Inicia sesión</h1>
          <p>Ingresa con tu cuenta para continuar.</p>
        </div>

        <form className="login-form" onSubmit={handleSubmit} noValidate>
          {localCredentials && (
            <aside className="local-credentials" aria-label="Credenciales locales de prueba">
              <div className="local-credentials-heading">
                <span className="local-credentials-title">Credenciales de prueba</span>
                <span className="local-credentials-badge">Solo desarrollo</span>
              </div>
              <dl>
                <div>
                  <dt>Correo</dt>
                  <dd><code>{localCredentials.email}</code></dd>
                </div>
                <div>
                  <dt>Contraseña</dt>
                  <dd><code>{localCredentials.password}</code></dd>
                </div>
              </dl>
              <button
                className="credentials-fill-button"
                type="button"
                onClick={() => {
                  setEmail(localCredentials.email);
                  setPassword(localCredentials.password);
                  setErrorMessage("");
                }}
              >
                Cargar en el formulario
              </button>
            </aside>
          )}

          {errorMessage && (
            <div className="form-alert" role="alert" aria-live="polite">
              <span className="alert-symbol" aria-hidden="true">!</span>
              <span>{errorMessage}</span>
            </div>
          )}

          <div className="form-field">
            <label htmlFor="email">Correo electrónico</label>
            <input
              id="email"
              name="email"
              type="email"
              autoComplete="username"
              placeholder="nombre@clinica.com"
              value={email}
              onChange={(event) => setEmail(event.target.value)}
              required
              maxLength={254}
              disabled={isSubmitting}
            />
          </div>

          <div className="form-field">
            <label htmlFor="password">Contraseña</label>
            <input
              id="password"
              name="password"
              type="password"
              autoComplete="current-password"
              placeholder="Ingresa tu contraseña"
              value={password}
              onChange={(event) => setPassword(event.target.value)}
              required
              maxLength={72}
              disabled={isSubmitting}
            />
          </div>

          <button className="button-primary login-submit" type="submit" disabled={isSubmitting}>
            {isSubmitting ? "Ingresando…" : "Ingresar"}
          </button>
        </form>

        {registrationEnabled && <p className="login-footnote">¿Aún no tienes cuenta? <Link className="text-link" to="/register">Regístrate</Link></p>}
      </section>

    </main>
  );
}
