import { cloneElement, isValidElement, type ReactElement, type ReactNode } from "react";

type FormFieldProps = {
  id: string;
  label: string;
  required?: boolean;
  error?: string;
  hint?: string;
  children: ReactNode;
};

export function FormField({ id, label, required = false, error, hint, children }: FormFieldProps) {
  const descriptionId = error ? `${id}-error` : hint ? `${id}-hint` : undefined;

  return (
    <div className="form-field">
      <label htmlFor={id}>
        {label}{required && <span className="required-label"> (obligatorio)</span>}
      </label>
      {isValidElement(children) && descriptionId
        ? cloneElement(children as ReactElement<{ "aria-describedby"?: string }>, { "aria-describedby": descriptionId })
        : children}
      {hint && !error && <span id={`${id}-hint`} className="form-field-hint">{hint}</span>}
      {error && <span id={`${id}-error`} className="form-field-error">{error}</span>}
    </div>
  );
}
