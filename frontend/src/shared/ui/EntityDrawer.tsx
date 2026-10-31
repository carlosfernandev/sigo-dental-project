import { useEffect, useRef, type ReactNode } from "react";

type EntityDrawerProps = {
  open: boolean;
  title: string;
  wide?: boolean;
  centered?: boolean;
  onRequestClose: () => void;
  children: ReactNode;
};

const focusableSelector = "a[href], button:not([disabled]), input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex='-1'])";

export function EntityDrawer({ open, title, wide = false, centered = false, onRequestClose, children }: EntityDrawerProps) {
  const drawerRef = useRef<HTMLElement>(null);
  const previousFocusRef = useRef<HTMLElement | null>(null);

  useEffect(() => {
    if (!open) return;
    previousFocusRef.current = document.activeElement instanceof HTMLElement ? document.activeElement : null;
    const originalOverflow = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    drawerRef.current?.querySelector<HTMLElement>(focusableSelector)?.focus();

    function handleKeyDown(event: KeyboardEvent) {
      if (event.key === "Escape") {
        event.preventDefault();
        onRequestClose();
        return;
      }
      if (event.key !== "Tab" || !drawerRef.current) return;
      const focusable = Array.from(drawerRef.current.querySelectorAll<HTMLElement>(focusableSelector));
      if (focusable.length === 0) return;
      const first = focusable[0];
      const last = focusable[focusable.length - 1];
      if (event.shiftKey && document.activeElement === first) {
        event.preventDefault();
        last.focus();
      } else if (!event.shiftKey && document.activeElement === last) {
        event.preventDefault();
        first.focus();
      }
    }

    document.addEventListener("keydown", handleKeyDown);
    return () => {
      document.removeEventListener("keydown", handleKeyDown);
      document.body.style.overflow = originalOverflow;
      previousFocusRef.current?.focus();
    };
  }, [open, onRequestClose]);

  if (!open) return null;

  return (
    <div className={`entity-drawer-layer${centered ? " entity-drawer-layer--centered" : ""}`} onMouseDown={(event) => { if (event.target === event.currentTarget) onRequestClose(); }}>
      <aside ref={drawerRef} className={`entity-drawer${wide ? " entity-drawer--wide" : ""}${centered ? " entity-drawer--centered" : ""}`} role="dialog" aria-modal="true" aria-labelledby="entity-drawer-title">
        <header className="entity-drawer-header">
          <h2 id="entity-drawer-title">{title}</h2>
          <button className="drawer-close-button" type="button" aria-label="Cerrar formulario" onClick={onRequestClose}>×</button>
        </header>
        <div className="entity-drawer-content">{children}</div>
      </aside>
    </div>
  );
}
