import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { Link, useLocation, useNavigate } from "react-router-dom";
import { useAuth } from "../../../app/AuthProvider";
import { ConfirmDialog } from "../../../shared/ui/ConfirmDialog";
import { DataTable, type DataTableColumn } from "../../../shared/ui/DataTable";
import { EntityDrawer } from "../../../shared/ui/EntityDrawer";
import { StatusChip } from "../../../shared/ui/StatusChip";
import { Toast } from "../../../shared/ui/Toast";
import { deactivateUser, listUsers, updateUser, createUser, type ManagedUser, type UserListResponse } from "../api";
import { UserForm } from "../components/UserForm";
import { userRoles } from "../schemas";

const roleLabels: Record<string, string> = {
  admin_ti: "Administrador TI",
  recepcionista_cajero: "Recepcionista / Cajero",
  odontologo: "Odontólogo",
  asistente_dental: "Asistente dental",
  auditor_contador: "Auditor / Contador",
};

export function UsersPage() {
  const { user: currentUser } = useAuth();
  const [result, setResult] = useState<UserListResponse>({ items: [], total: 0, page: 1, page_size: 25 });
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState("");
  const [operationError, setOperationError] = useState("");
  const [toast, setToast] = useState<{ message: string; undoUser?: ManagedUser } | null>(null);
  const [retryCount, setRetryCount] = useState(0);
  const [page, setPage] = useState(1);
  const [searchInput, setSearchInput] = useState("");
  const [search, setSearch] = useState("");
  const [role, setRole] = useState("");
  const [isActive, setIsActive] = useState("");
  const [sortBy, setSortBy] = useState("email");
  const [sortOrder, setSortOrder] = useState<"asc" | "desc">("asc");
  const [editingUser, setEditingUser] = useState<ManagedUser | null>(null);
  const [deactivatingUser, setDeactivatingUser] = useState<ManagedUser | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [showDiscardDialog, setShowDiscardDialog] = useState(false);
  const dirtyRef = useRef(false);
  const searchRef = useRef<HTMLInputElement>(null);
  const location = useLocation();
  const navigate = useNavigate();
  const isCreateRoute = location.pathname === "/users/new";
  const drawerOpen = isCreateRoute || editingUser !== null;

  useEffect(() => {
    const state = location.state as { notice?: string } | null;
    if (state?.notice) {
      setToast({ message: state.notice });
      navigate(location.pathname, { replace: true, state: null });
    }
  }, [location.pathname, location.state, navigate]);

  useEffect(() => {
    const focusSearch = (event: KeyboardEvent) => {
      const target = event.target;
      const isTyping = target instanceof HTMLElement && (target.isContentEditable || ["INPUT", "TEXTAREA", "SELECT"].includes(target.tagName));
      if (event.key === "/" && !isTyping) {
        event.preventDefault();
        searchRef.current?.focus();
      }
    };
    window.addEventListener("keydown", focusSearch);
    return () => window.removeEventListener("keydown", focusSearch);
  }, []);

  useEffect(() => {
    const timeout = window.setTimeout(() => {
      setSearch(searchInput.trim());
      setPage(1);
    }, 250);
    return () => window.clearTimeout(timeout);
  }, [searchInput]);

  useEffect(() => {
    let active = true;
    setLoading(true);
    setLoadError("");
    listUsers({ page, search, role, isActive, sortBy, sortOrder })
      .then((data) => {
        if (!active) return;
        if (data.items.length === 0 && data.total > 0 && page > 1) {
          setPage(Math.max(1, Math.ceil(data.total / 25)));
          return;
        }
        setResult(data);
      })
      .catch((error: unknown) => {
        if (active) setLoadError(error instanceof Error ? error.message : "No se pudo cargar la lista de usuarios.");
      })
      .finally(() => { if (active) setLoading(false); });
    return () => { active = false; };
  }, [page, retryCount, role, search, sortBy, sortOrder, isActive]);

  const setDirty = useCallback((dirty: boolean) => {
    dirtyRef.current = dirty;
  }, []);

  const closeDrawerNow = useCallback(() => {
    dirtyRef.current = false;
    setEditingUser(null);
    setShowDiscardDialog(false);
    if (isCreateRoute) navigate("/users");
  }, [isCreateRoute, navigate]);

  const requestDrawerClose = useCallback(() => {
    if (dirtyRef.current) {
      setShowDiscardDialog(true);
      return;
    }
    closeDrawerNow();
  }, [closeDrawerNow]);

  function sortTable(nextSort: string, nextOrder: "asc" | "desc") {
    setSortBy(nextSort);
    setSortOrder(nextOrder);
    setPage(1);
  }

  async function saveUser(eventValues: { email: string; password: string; role_id: string }, isEditing: boolean) {
    setOperationError("");
    setIsSubmitting(true);
    try {
      if (isEditing && editingUser) {
        await updateUser(editingUser.id, { role: eventValues.role_id });
        setRetryCount((count) => count + 1);
        setToast({ message: "Usuario actualizado correctamente." });
        closeDrawerNow();
      } else {
        await createUser({ email: eventValues.email, password: eventValues.password, role_id: eventValues.role_id as (typeof userRoles)[number] });
        setDirty(false);
        setRetryCount((count) => count + 1);
        navigate("/users", { state: { notice: "Usuario creado correctamente." } });
      }
    } catch (error) {
      setOperationError(error instanceof Error ? error.message : "No se pudo guardar el usuario.");
    } finally {
      setIsSubmitting(false);
    }
  }

  async function confirmDeactivation(reason: string) {
    if (!deactivatingUser) return;
    setOperationError("");
    setIsSubmitting(true);
    try {
      await deactivateUser(deactivatingUser.id, reason);
      setRetryCount((count) => count + 1);
      setToast({ message: `Usuario desactivado: ${deactivatingUser.email}.`, undoUser: deactivatingUser });
      setDeactivatingUser(null);
    } catch (error) {
      setOperationError(error instanceof Error ? error.message : "No se pudo desactivar el usuario.");
    } finally {
      setIsSubmitting(false);
    }
  }

  async function undoDeactivation() {
    const target = toast?.undoUser;
    if (!target) return;
    setToast(null);
    setIsSubmitting(true);
    setOperationError("");
    try {
      await updateUser(target.id, { role: target.role, is_active: true });
      setRetryCount((count) => count + 1);
      setToast({ message: `Desactivación deshecha: ${target.email}.` });
    } catch (error) {
      setOperationError(error instanceof Error ? error.message : "No se pudo reactivar el usuario.");
    } finally {
      setIsSubmitting(false);
    }
  }

  async function reactivateUser(target: ManagedUser) {
    setOperationError("");
    try {
      await updateUser(target.id, { role: target.role, is_active: true });
      setRetryCount((count) => count + 1);
      setToast({ message: `Usuario reactivado: ${target.email}.` });
    } catch (error) {
      setOperationError(error instanceof Error ? error.message : "No se pudo reactivar el usuario.");
    }
  }

  const closeToast = useCallback(() => setToast(null), []);

  const columns = useMemo<DataTableColumn<ManagedUser>[]>(() => [
    {
      key: "email",
      header: "Correo electrónico",
      sortKey: "email",
      render: (row) => <button className="table-row-primary-action" type="button" onClick={() => { setOperationError(""); setEditingUser(row); setDirty(false); }} aria-label={`Editar usuario ${row.email}`}>{row.email}</button>,
    },
    { key: "role", header: "Rol", sortKey: "role", render: (row) => roleLabels[row.role] ?? row.role.replaceAll("_", " ") },
    { key: "is_active", header: "Estado", sortKey: "is_active", render: (row) => <StatusChip label={row.is_active ? "Activo" : "Inactivo"} tone={row.is_active ? "positive" : "negative"} /> },
    {
      key: "actions",
      header: "Acciones",
      render: (row) => (
        <div className="users-row-actions">
          {row.is_active && row.id !== currentUser?.id
            ? <button className="secondary-button destructive-text-button" type="button" onClick={() => { setOperationError(""); setDeactivatingUser(row); }}>Desactivar</button>
            : !row.is_active && <button className="secondary-button" type="button" onClick={() => void reactivateUser(row)}>Reactivar</button>}
        </div>
      ),
    },
  ], [currentUser?.id, setDirty]);

  function clearFilters() {
    setSearchInput("");
    setSearch("");
    setRole("");
    setIsActive("");
    setPage(1);
  }

  return (
    <section className="module-page users-page" aria-labelledby="users-title">
      <header className="users-page-heading">
        <div>
          <p className="eyebrow">Administración</p>
          <h1 id="users-title">Usuarios</h1>
        </div>
        <Link className="button-primary users-create-button" to="/users/new">Nuevo usuario</Link>
      </header>

      {operationError && !drawerOpen && !deactivatingUser && <div className="form-alert users-notice" role="alert">{operationError}</div>}

      <div className="users-toolbar">
        <div className="users-search-field">
          <label htmlFor="users-search">Buscar usuarios</label>
          <input ref={searchRef} id="users-search" type="search" value={searchInput} onChange={(event) => setSearchInput(event.target.value)} placeholder="Correo o rol" />
          <span className="users-search-shortcut" aria-hidden="true">/</span>
        </div>
        <label className="users-filter-field" htmlFor="users-role-filter">
          Rol
          <select id="users-role-filter" value={role} onChange={(event) => { setRole(event.target.value); setPage(1); }}>
            <option value="">Todos</option>
            {userRoles.map((value) => <option key={value} value={value}>{roleLabels[value]}</option>)}
          </select>
        </label>
        <label className="users-filter-field" htmlFor="users-status-filter">
          Estado
          <select id="users-status-filter" value={isActive} onChange={(event) => { setIsActive(event.target.value); setPage(1); }}>
            <option value="">Todos</option>
            <option value="true">Activos</option>
            <option value="false">Inactivos</option>
          </select>
        </label>
      </div>

      <DataTable
        caption="Usuarios registrados en SIGO-Dental"
        columns={columns}
        rows={result.items}
        rowKey={(row) => row.id}
        loading={loading}
        error={loadError}
        empty={!loading && result.items.length === 0}
        hasFilters={Boolean(search || role || isActive)}
        onRetry={() => setRetryCount((count) => count + 1)}
        onResetFilters={clearFilters}
        emptyAction={<Link className="button-primary users-create-button" to="/users/new">Nuevo usuario</Link>}
        page={page}
        pageSize={25}
        total={result.total}
        onPageChange={setPage}
        sortBy={sortBy}
        sortOrder={sortOrder}
        onSort={sortTable}
      />

      <EntityDrawer open={drawerOpen} title={isCreateRoute ? "Nuevo usuario" : "Editar usuario"} onRequestClose={requestDrawerClose}>
        {operationError && <div className="form-alert" role="alert">{operationError}</div>}
        <UserForm
          key={editingUser?.id ?? (isCreateRoute ? "create" : "closed")}
          initialData={editingUser ?? undefined}
          isBusy={isSubmitting}
          onDirtyChange={setDirty}
          onCancel={requestDrawerClose}
          onSubmit={saveUser}
        />
      </EntityDrawer>

      <ConfirmDialog
        open={Boolean(deactivatingUser)}
        title="Desactivar usuario"
        description={deactivatingUser ? `La cuenta ${deactivatingUser.email} perderá acceso. El registro se conservará y podrás reactivarlo.` : "Confirma la desactivación de la cuenta."}
        confirmLabel="Desactivar usuario"
        requiresReason
        isBusy={isSubmitting}
        errorMessage={operationError}
        onCancel={() => setDeactivatingUser(null)}
        onConfirm={(reason) => void confirmDeactivation(reason)}
      />

      {toast && <Toast message={toast.message} actionLabel={toast.undoUser ? "Deshacer" : undefined} onAction={toast.undoUser ? () => void undoDeactivation() : undefined} onClose={closeToast} />}

      <ConfirmDialog
        open={showDiscardDialog}
        title="Descartar cambios"
        description="Hay cambios sin guardar. Si continúas, se perderán."
        confirmLabel="Descartar cambios"
        onCancel={() => setShowDiscardDialog(false)}
        onConfirm={() => { setShowDiscardDialog(false); closeDrawerNow(); }}
      />
    </section>
  );
}
