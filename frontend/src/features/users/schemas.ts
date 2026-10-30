import { z } from "zod";

export const userRoles = [
  "admin_ti",
  "recepcionista_cajero",
  "odontologo",
  "asistente_dental",
  "auditor_contador",
] as const;

const commonPasswords = new Set([
  "password123456",
  "123456789012",
  "qwertyuiop123",
  "letmeinplease",
  "adminadmin123",
  "odontologia123",
]);

const emailPasswordSchema = z.object({
  email: z.string().trim().email("Ingresa un correo válido.").max(254),
  password: z.string().min(1, "Ingresa una contraseña temporal."),
}).superRefine(({ email, password }, context) => {
  if (Array.from(password).length < 12) {
    context.addIssue({ code: "custom", path: ["password"], message: "La contraseña debe tener al menos 12 caracteres." });
  }
  if (new TextEncoder().encode(password).length > 72) {
    context.addIssue({ code: "custom", path: ["password"], message: "La contraseña no puede superar los 72 bytes." });
  }
  if (commonPasswords.has(password.trim().toLowerCase())) {
    context.addIssue({ code: "custom", path: ["password"], message: "Elige una contraseña menos común." });
  }
  if (email.trim() && password.toLowerCase().includes(email.trim().toLowerCase())) {
    context.addIssue({ code: "custom", path: ["password"], message: "La contraseña no puede contener el correo." });
  }
});

export const publicRegistrationSchema = emailPasswordSchema;

export const createUserSchema = emailPasswordSchema.and(z.object({
  role_id: z.enum(userRoles, { message: "Selecciona un rol." }),
}));

export const updateUserFormSchema = z.object({
  role_id: z.enum(userRoles, { message: "Selecciona un rol." }),
});

export type CreateUserInput = z.infer<typeof createUserSchema>;
