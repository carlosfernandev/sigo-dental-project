import { z } from "zod";
import { httpRequest } from "../../shared/lib/http";

const allergySchema = z.object({ name: z.string(), is_critical: z.boolean() });
const historySchema = z.object({
  systemic_diseases: z.array(z.string()),
  allergies: z.array(allergySchema),
  surgical_history: z.array(z.string()),
  coagulopathies: z.array(z.string()),
});
const guardianSchema = z.object({ name: z.string(), identification: z.string(), phone: z.string() });
const patientSchema = z.object({
  id: z.string().uuid(), first_name: z.string(), last_name: z.string(), identification: z.string(), date_of_birth: z.string(),
  email: z.string(), phone: z.string(), address: z.string(), emergency_contact_name: z.string(),
  emergency_contact_phone: z.string(), legal_guardian: guardianSchema.nullish(), insurance_plan: z.string(),
  medical_history: historySchema, critical_alerts: z.array(z.string()),
  created_at: z.string().optional(), updated_at: z.string().optional(),
});
const patientSummarySchema = z.object({
  id: z.string().uuid(), first_name: z.string(), last_name: z.string(), identification: z.string().optional().default(""), date_of_birth: z.string(), phone: z.string(), insurance_plan: z.string().optional().default(""),
});

export type Patient = z.infer<typeof patientSchema>;
export type PatientSummary = z.infer<typeof patientSummarySchema>;
export type PatientInput = Omit<Patient, "id" | "critical_alerts" | "created_at" | "updated_at">;

export async function listPatients(): Promise<PatientSummary[]> {
  const payload = await httpRequest("/patients");
  const parsed = z.object({ data: z.array(patientSummarySchema) }).safeParse(payload);
  if (!parsed.success) throw new Error("El servidor devolvió una lista de pacientes inválida.");
  return parsed.data.data;
}

export async function getPatient(id: string): Promise<Patient> {
  const payload = await httpRequest(`/patients/${encodeURIComponent(id)}`);
  const parsed = z.object({ data: patientSchema }).safeParse(payload);
  if (!parsed.success) throw new Error("No se pudo leer la ficha del paciente.");
  return parsed.data.data;
}

export async function savePatient(input: PatientInput, id?: string): Promise<Patient> {
  const payload = await httpRequest(id ? `/patients/${encodeURIComponent(id)}` : "/patients", {
    method: id ? "PUT" : "POST",
    body: input,
  });
  const parsed = z.object({ data: patientSchema }).safeParse(payload);
  if (!parsed.success) throw new Error("El servidor devolvió una ficha de paciente inválida.");
  return parsed.data.data;
}
