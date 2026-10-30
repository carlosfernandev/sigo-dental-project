import { Fragment, useEffect, useState, type FormEvent } from "react";
import { Link, useLocation, useNavigate } from "react-router-dom";
import { deactivateUser, listUsers, updateUser, type ManagedUser } from "../api";
import { userRoles } from "../schemas";

const roleLabels: Record<string, string> = {
  admin_ti: "Administrador TI",
  recepcionista_cajero: "Recepcionista / Cajero",
  odontologo: "Odontólogo",
  asistente_dental: "Asistente dental",
  auditor_contador: "Auditor / Contador",
};

export function UsersPage() {
  const [users, setUsers] = useState<ManagedUser[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [loadError, setLoadError] = useState("");
  const [errorMessage, setErrorMessage] = useState("");
  const [notice, setNotice] = useState("");
  const [editingUser, setEditingUser] = useState<ManagedUser | null>(null);
  const [editRole, setEditRole] = useState("");
  const [editIsActive, setEditIsActive] = useState(false);
  const [pendingDeactivationId, setPendingDeactivationId] = useState("");
  const [busyUserId, setBusyUserId] = useState("");
  const location = useLocation();
  const navigate = useNavigate();

  useEffect(() => {
    const state = location.state as { notice?: string } | null;
    if (state?.notice) {
      setNotice(state.notice);
      navigate(location.pathname, { replace: true, state: null });
    }
  }, [location.pathname, location.state, navigate]);

  useEffect(() => {
    let active = true;
    setIsLoading(true);
    listUsers()
      .then((result) => { if (active) setUsers(result); })
      .catch((error: unknown) => {
        if (active) setLoadError(error instanceof Error ? error.message : "No se pudo cargar la lista de usuarios.");
      })
      .finally(() => { if (active) setIsLoading(false); });
    return () => { active = false; };
  }, []);

  function startEditing(user: ManagedUser) {
    setErrorMessage("");
    setEditingUser(user);
    setEditRole(user.role);
    setEditIsActive(user.is_active);
    setPendingDeactivationId("");
  }

  async function saveUser(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!editingUser) return;

    setErrorMessage("");
    setBusyUserId(editingUser.id);
    try {
      const updatedUser = await updateUser(editingUser.id, { role: editRole, is_active: editIsActive });
      setUsers((currentUsers) => currentUsers.map((user) => user.id === updatedUser.id ? updatedUser : user));
      setEditingUser(null);
      setNotice("Usuario actualizado correctamente.");
    } catch (error) {
      setErrorMessage(error instanceof Error ? error.message : "No se pudo actualizar el usuario.");
    } finally {
      setBusyUserId("");
    }
  }

  async function deactivate(user: ManagedUser) {
    setErrorMessage("");
    setBusyUserId(user.id);
    try {
      await deactivateUser(user.id);
      setUsers((currentUsers) => currentUsers.map((currentUser) => currentUser.id === user.id ? { ...currentUser, is_active: false } : currentUser));
      setPendingDeactivationId("");
      setNotice("Usuario desactivado. Su registro se conserva y puedes reactivarlo.");
    } catch (error) {
      setErrorMessage(error instanceof Error ? error.message : "No se pudo desactivar el usuario.");
    } finally {
      setBusyUserId("");
    }
  }

  async function reactivate(user: ManagedUser) {
    setErrorMessage("");
    setBusyUserId(user.id);
    try {
      const updatedUser = await updateUser(user.id, { role: user.role, is_active: true });
      setUsers((currentUsers) => currentUsers.map((currentUser) => currentUser.id === user.id ? updatedUser : currentUser));
      setNotice("Usuario reactivado correctamente.");
    } catch (error) {
      setErrorMessage(error instanceof Error ? error.message : "No se pudo reactivar el usuario.");
    } finally {
      setBusyUserId("");
    }
  }

  return (
    <section className="module-page users-page" aria-labelledby="users-title">
      <header className="page-heading users-page-heading">
        <div>
          <p className="eyebrow">Administración</p>
          <h1 id="users-title">Usuarios</h1>
          <p>Consulta y administra las cuentas del personal.</p>
        </div>
        <Link className="button-primary users-create-button" to="/users/new">Registrar usuario</Link>
      </header>

      {notice && <div className="form-success users-notice" role="status">{notice}</div>}
      {(errorMessage || loadError) && <div className="form-alert users-notice" role="alert">{errorMessage || loadError}</div>}

      <div className="users-table-container">
        <table className="users-table">
          <caption className="visually-hidden">Usuarios registrados en SIGO-Dental</caption>
          <thead>
            <tr>
              <th scope="col">Correo electrónico</th>
              <th scope="col">Rol</th>
              <th scope="col">Estado</th>
              <th scope="col">Acciones</th>
            </tr>
          </thead>
          <tbody>
            {isLoading && <tr><td className="users-table-message" colSpan={4}>Cargando usuarios…</td></tr>}
            {!isLoading && loadError && <tr><td className="users-table-message" colSpan={4}>No se pudo mostrar la lista.</td></tr>}
            {!isLoading && !loadError && users.length === 0 && (
              <tr><td className="users-table-message" colSpan={4}>Aún no hay usuarios registrados.</td></tr>
            )}
            {!isLoading && !loadError && users.map((user) => (
              <Fragment key={user.id}>
                <tr>
                  <td>{user.email}</td>
                  <td>{roleLabels[user.role] ?? user.role.replaceAll("_", " ")}</td>
                  <td><span className={`user-status${user.is_active ? " is-active" : " is-inactive"}`}>{user.is_active ? "Activo" : "Inactivo"}</span></td>
                  <td className="users-actions-cell">
                    {pendingDeactivationId === user.id ? (
                      <span className="users-confirm-actions">
                        <span>¿Desactivar?</span>
                        <button className="users-action-button is-danger" type="button" disabled={busyUserId === user.id} onClick={() => void deactivate(user)}>Confirmar</button>
                        <button className="users-action-button" type="button" onClick={() => setPendingDeactivationId("")}>Cancelar</button>
                      </span>
                    ) : (
                      <span className="users-row-actions">
                        <button className="users-action-button" type="button" disabled={busyUserId === user.id} onClick={() => startEditing(user)}>Editar</button>
                        {user.is_active
                          ? <button className="users-action-button is-danger" type="button" disabled={busyUserId === user.id} onClick={() => setPendingDeactivationId(user.id)}>Desactivar</button>
                          : <button className="users-action-button" type="button" disabled={busyUserId === user.id} onClick={() => void reactivate(user)}>Reactivar</button>}
                      </span>
                    )}
                  </td>
                </tr>
                {editingUser?.id === user.id && (
                  <tr className="users-edit-row">
                    <td colSpan={4}>
                      <form className="users-edit-form" onSubmit={saveUser}>
                        <div className="form-field">
                          <label htmlFor={`edit-role-${user.id}`}>Rol de {user.email}</label>
                          <select id={`edit-role-${user.id}`} value={editRole} onChange={(event) => setEditRole(event.target.value)} disabled={busyUserId === user.id}>
                            {userRoles.map((role) => <option key={role} value={role}>{roleLabels[role]}</option>)}
                          </select>
                        </div>
                        <label className="users-active-control" htmlFor={`edit-active-${user.id}`}>
                          <input id={`edit-active-${user.id}`} type="checkbox" checked={editIsActive} onChange={(event) => setEditIsActive(event.target.checked)} disabled={busyUserId === user.id} />
                          Cuenta activa
                        </label>
                        <button className="button-primary users-save-button" type="submit" disabled={busyUserId === user.id}>{busyUserId === user.id ? "Guardando…" : "Guardar cambios"}</button>
                        <button className="users-action-button" type="button" disabled={busyUserId === user.id} onClick={() => setEditingUser(null)}>Cancelar</button>
                      </form>
                    </td>
                  </tr>
                )}
              </Fragment>
            ))}
          </tbody>
        </table>
      </div>
    </section>
  );
}
