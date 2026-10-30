import { httpRequest } from "../../shared/lib/http";
import { createUserSchema, publicRegistrationSchema, type CreateUserInput } from "./schemas";
import { z } from "zod";

const registrationStatusSchema = z.object({ data: z.object({ enabled: z.boolean() }) });
const managedUserSchema = z.object({
  id: z.string().uuid(),
  email: z.string().email(),
  role: z.string().min(1),
  is_active: z.boolean(),
});
const managedUsersSchema = z.object({ data: z.array(managedUserSchema) });
const updateUserSchema = z.object({
  role: z.enum(["admin_ti", "recepcionista_cajero", "odontologo", "asistente_dental", "auditor_contador"]),
  is_active: z.boolean(),
});

export type ManagedUser = z.infer<typeof managedUsersSchema>["data"][number];

export async function getRegistrationEnabled() {
  const response = await httpRequest("/auth/registration-status");
  return registrationStatusSchema.parse(response).data.enabled;
}

export async function createUser(input: CreateUserInput) {
  const request = createUserSchema.parse(input);
  const payload = { email: request.email, password: request.password, role: request.role_id };
  return httpRequest("/users", { method: "POST", body: payload });
}

export async function listUsers(): Promise<ManagedUser[]> {
  const response = await httpRequest("/users");
  return managedUsersSchema.parse(response).data;
}

export async function updateUser(id: string, input: { role: ManagedUser["role"]; is_active: boolean }): Promise<ManagedUser> {
  const userId = z.string().uuid().parse(id);
  const request = updateUserSchema.parse(input);
  const response = await httpRequest(`/users/${userId}`, { method: "PUT", body: request });
  return z.object({ data: managedUserSchema }).parse(response).data;
}

export async function deactivateUser(id: string): Promise<void> {
  const userId = z.string().uuid().parse(id);
  await httpRequest(`/users/${userId}`, { method: "DELETE" });
}

export async function registerAccount(input: { email: string; password: string }) {
  const request = publicRegistrationSchema.parse(input);
  return httpRequest("/auth/register", { method: "POST", body: request });
}
