import { z } from "zod";

export const loginInputSchema = z.object({
  email: z.string().trim().email("Ingresa un correo válido.").max(254),
  password: z.string().min(1, "Ingresa tu contraseña.").max(72).refine(
    (password) => new TextEncoder().encode(password).length <= 72,
    "La contraseña no puede superar los 72 bytes.",
  ),
});

export const loginResponseSchema = z.object({
  data: z.object({
    status: z.literal("authenticated"),
    must_change_password: z.boolean(),
    access_token: z.string().min(1),
    expires_at: z.string().datetime({ offset: true }),
  }),
});

export const currentUserResponseSchema = z.object({
  data: z.object({
    id: z.string().uuid(),
    role: z.string().min(1),
  }),
});

export type LoginInput = z.infer<typeof loginInputSchema>;
export type AuthUser = z.infer<typeof currentUserResponseSchema>["data"] & {
  email: string;
  mustChangePassword: boolean;
};
