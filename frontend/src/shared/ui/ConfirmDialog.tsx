import { useEffect, useRef, useState, type FormEvent } from "react";
import { FormField } from "./FormField";

type ConfirmDialogProps = {
  open: boolean;
  title: string;
  description: string;
  confirmLabel: string;
  requiresReason?: boolean;
  isBusy?: boolean;
  errorMessage?: string;
  onCancel: () => void;
  onConfirm: (reason: string) => void;
};

export function ConfirmDialog({
  open,
  title,
  description,
  confirmLabel,
  requiresReason = false,
  isBusy = false,
  errorMessage = "",
  onCancel,
  onConfirm,
}: ConfirmDialogProps) {
  const dialogRef = useRef<HTMLDialogElement>(null);
  const [reason, setReason] = useState("");

  useEffect(() => {
    const dialog = dialogRef.current;
    if (!dialog) return;
    if (open && !dialog.open) {
      setReason("");
      dialog.showModal();
    }
    if (!open && dialog.open) dialog.close();
  }, [open]);

  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (requiresReason && reason.trim().length < 3) return;
    onConfirm(reason.trim());
  }

  return (
    <dialog
      ref={dialogRef}
      className="confirm-dialog"
      aria-labelledby="confirm-dialog-title"
      aria-describedby="confirm-dialog-description"
      onCancel={(event) => { event.preventDefault(); if (!isBusy) onCancel(); }}
    >
      <form onSubmit={submit}>
        <header className="confirm-dialog-header">
          <p className="eyebrow">Confirmación</p>
          <h2 id="confirm-dialog-title">{title}</h2>
        </header>
        <p id="confirm-dialog-description" className="confirm-dialog-description">{description}</p>
        {errorMessage && <div className="form-alert" role="alert">{errorMessage}</div>}
        {requiresReason && (
          <FormField id="confirm-reason" label="Motivo" required error={reason.length > 0 && reason.trim().length < 3 ? "Escribe un motivo de al menos 3 caracteres." : undefined}>
            <textarea id="confirm-reason" value={reason} onChange={(event) => setReason(event.target.value)} minLength={3} maxLength={500} required disabled={isBusy} />
          </FormField>
        )}
        <footer className="confirm-dialog-actions">
          <button className="secondary-button" type="button" onClick={onCancel} disabled={isBusy}>Cancelar</button>
          <button className="destructive-button" type="submit" disabled={isBusy || (requiresReason && reason.trim().length < 3)}>
            {isBusy ? "Guardando…" : confirmLabel}
          </button>
        </footer>
      </form>
    </dialog>
  );
}
