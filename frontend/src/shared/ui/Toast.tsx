import { useEffect, useState } from "react";

type ToastProps = {
  message: string;
  actionLabel?: string;
  onAction?: () => void;
  onClose: () => void;
};

export function Toast({ message, actionLabel, onAction, onClose }: ToastProps) {
  const [paused, setPaused] = useState(false);

  useEffect(() => {
    if (paused) return;
    const timeout = window.setTimeout(onClose, 8000);
    return () => window.clearTimeout(timeout);
  }, [message, onClose, paused]);

  return (
    <div className="app-toast" role="status" onMouseEnter={() => setPaused(true)} onMouseLeave={() => setPaused(false)} onFocus={() => setPaused(true)} onBlur={(event) => { if (!event.currentTarget.contains(event.relatedTarget)) setPaused(false); }}>
      <span>{message}</span>
      {actionLabel && onAction && <button className="toast-action" type="button" onClick={onAction}>{actionLabel}</button>}
      <button className="toast-close" type="button" aria-label="Cerrar notificación" onClick={onClose}>×</button>
    </div>
  );
}
