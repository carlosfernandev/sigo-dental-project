import { z } from "zod";
import { httpRequest } from "../../../shared/lib/http";

const timestamps = { created_at: z.string(), updated_at: z.string() };
const operatorySchema = z.object({
  id: z.string().uuid(), name: z.string(), description: z.string(), ...timestamps,
});
const chairSchema = z.object({
  id: z.string().uuid(), operatory_id: z.string().uuid(), name: z.string(), ...timestamps,
});
const dentistSchema = z.object({
  id: z.string().uuid(), first_name: z.string(), last_name: z.string(), specialty: z.string(), active: z.boolean(), ...timestamps,
});

export type Operatory = z.infer<typeof operatorySchema>;
export type Chair = z.infer<typeof chairSchema>;
export type Dentist = z.infer<typeof dentistSchema>;
export type Resource = Operatory | Chair | Dentist;
export type ResourceKind = "operatories" | "chairs" | "dentists";
export type OperatoryInput = Pick<Operatory, "name" | "description">;
export type ChairInput = Pick<Chair, "name" | "operatory_id">;
export type DentistInput = Pick<Dentist, "first_name" | "last_name" | "specialty" | "active">;

function parseList<T>(payload: unknown, itemSchema: z.ZodType<T>, label: string): T[] {
  const result = z.object({ data: z.array(itemSchema) }).safeParse(payload);
  if (!result.success) throw new Error(`El servidor devolvió una lista de ${label} inválida.`);
  return result.data.data as T[];
}

function parseItem<T>(payload: unknown, itemSchema: z.ZodType<T>, label: string): T {
  const result = z.object({ data: itemSchema }).safeParse(payload);
  if (!result.success) throw new Error(`El servidor devolvió un registro de ${label} inválido.`);
  return result.data.data as T;
}

export async function listOperatories() {
  return parseList(await httpRequest("/scheduling/operatories"), operatorySchema, "recintos");
}
export async function listChairs() {
  return parseList(await httpRequest("/scheduling/chairs"), chairSchema, "sillones");
}
export async function listDentists() {
  return parseList(await httpRequest("/scheduling/dentists"), dentistSchema, "odontólogos");
}

export async function saveOperatory(input: OperatoryInput, id?: string) {
  return parseItem(await httpRequest(id ? `/scheduling/operatories/${id}` : "/scheduling/operatories", {
    method: id ? "PUT" : "POST", body: input,
  }), operatorySchema, "recinto");
}
export async function saveChair(input: ChairInput, id?: string) {
  return parseItem(await httpRequest(id ? `/scheduling/chairs/${id}` : "/scheduling/chairs", {
    method: id ? "PUT" : "POST", body: input,
  }), chairSchema, "sillón");
}
export async function saveDentist(input: DentistInput, id?: string) {
  return parseItem(await httpRequest(id ? `/scheduling/dentists/${id}` : "/scheduling/dentists", {
    method: id ? "PUT" : "POST", body: input,
  }), dentistSchema, "odontólogo");
}
