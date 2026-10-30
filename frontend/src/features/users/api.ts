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
const paginatedUsersSchema = z.object({
  items: z.array(managedUserSchema),
  total: z.number().int().nonnegative(),
  page: z.number().int().positive(),
  page_size: z.literal(25),
});
const managedUsersSchema = z.object({ data: z.union([paginatedUsersSchema, z.array(managedUserSchema)]) });
const updateUserSchema = z.object({
  role: z.enum(["admin_ti", "recepcionista_cajero", "odontologo", "asistente_dental", "auditor_contador"]),
  is_active: z.boolean().optional(),
});

export type ManagedUser = z.infer<typeof managedUserSchema>;
export type UserListResponse = z.infer<typeof paginatedUsersSchema>;

export async function getRegistrationEnabled() {
  const response = await httpRequest("/auth/registration-status");
  return registrationStatusSchema.parse(response).data.enabled;
}

export async function createUser(input: CreateUserInput) {
  const request = createUserSchema.parse(input);
  const payload = { email: request.email, password: request.password, role: request.role_id };
  return httpRequest("/users", { method: "POST", body: payload });
}

export type UserListQuery = {
  page: number;
  search: string;
  role: string;
  isActive: string;
  sortBy: string;
  sortOrder: "asc" | "desc";
};

export async function listUsers(query: UserListQuery): Promise<UserListResponse> {
  const params = new URLSearchParams({
    page: String(query.page),
    page_size: "25",
    search: query.search,
    sort_by: query.sortBy,
    sort_order: query.sortOrder,
  });
  if (query.role) params.set("role", query.role);
  if (query.isActive) params.set("is_active", query.isActive);
  const response = await httpRequest(`/users?${params.toString()}`);
  const parsed = managedUsersSchema.safeParse(response);
  if (!parsed.success) throw new Error("El servidor devolvió una respuesta de usuarios inválida.");

  if (!Array.isArray(parsed.data.data)) return parsed.data.data;

  const filtered = parsed.data.data.filter((user) => {
    const matchesSearch = !query.search || `${user.email} ${user.role}`.toLocaleLowerCase().includes(query.search.toLocaleLowerCase());
    const matchesRole = !query.role || user.role === query.role;
    const matchesStatus = !query.isActive || user.is_active === (query.isActive === "true");
    return matchesSearch && matchesRole && matchesStatus;
  });
  filtered.sort((left, right) => {
    const a = query.sortBy === "role" ? left.role : query.sortBy === "is_active" ? String(left.is_active) : left.email;
    const b = query.sortBy === "role" ? right.role : query.sortBy === "is_active" ? String(right.is_active) : right.email;
    const comparison = a.localeCompare(b);
    return query.sortOrder === "desc" ? -comparison : comparison;
  });
  const start = (query.page - 1) * 25;
  return {
    items: filtered.slice(start, start + 25),
    total: filtered.length,
    page: query.page,
    page_size: 25,
  };
}

export async function updateUser(id: string, input: { role: ManagedUser["role"]; is_active?: boolean }): Promise<ManagedUser> {
  const userId = z.string().uuid().parse(id);
  const request = updateUserSchema.parse(input);
  const response = await httpRequest(`/users/${userId}`, { method: "PUT", body: request });
  return z.object({ data: managedUserSchema }).parse(response).data;
}

export async function deactivateUser(id: string, reason: string): Promise<void> {
  const userId = z.string().uuid().parse(id);
  await httpRequest(`/users/${userId}`, { method: "DELETE", body: { reason } });
}

export async function registerAccount(input: { email: string; password: string }) {
  const request = publicRegistrationSchema.parse(input);
  return httpRequest("/auth/register", { method: "POST", body: request });
}
