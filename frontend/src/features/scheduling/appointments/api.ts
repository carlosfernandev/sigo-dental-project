import { z } from "zod";
import { httpRequest } from "../../../shared/lib/http";

export const appointmentStates = ["scheduled", "confirmed", "checked_in", "in_progress", "completed", "cancelled", "no_show"] as const;
export type AppointmentState = (typeof appointmentStates)[number];

const appointmentSchema = z.object({
  id: z.string().uuid(), chair_id: z.string().uuid(), dentist_id: z.string().uuid(), patient_id: z.string().uuid(),
  start_time: z.string().datetime({ offset: true }), end_time: z.string().datetime({ offset: true }),
  state: z.enum(appointmentStates), notes: z.string(), created_at: z.string(), updated_at: z.string(),
});
const appointmentSummarySchema = z.object({
  id: z.string().uuid(), chair_id: z.string().uuid(), chair_name: z.string(), dentist_id: z.string().uuid(),
  dentist_name: z.string(), patient_id: z.string().uuid(), patient_name: z.string(),
  start_time: z.string().datetime({ offset: true }), end_time: z.string().datetime({ offset: true }),
  state: z.enum(appointmentStates),
});
const appointmentsSchema = z.object({ data: z.array(appointmentSummarySchema) });
const appointmentResponseSchema = z.object({ data: appointmentSchema });

export type Appointment = z.infer<typeof appointmentSchema>;
export type AppointmentSummary = z.infer<typeof appointmentSummarySchema>;
export type AppointmentInput = Pick<Appointment, "chair_id" | "dentist_id" | "patient_id" | "start_time" | "end_time" | "state" | "notes">;

export type AppointmentFilters = {
  startDate: string;
  endDate: string;
  chairID?: string;
  dentistID?: string;
};

export async function listAppointments(filters: AppointmentFilters): Promise<AppointmentSummary[]> {
  const params = new URLSearchParams({ start_date: filters.startDate, end_date: filters.endDate });
  if (filters.chairID) params.set("chair_id", filters.chairID);
  if (filters.dentistID) params.set("dentist_id", filters.dentistID);
  const parsed = appointmentsSchema.safeParse(await httpRequest(`/scheduling/appointments?${params.toString()}`));
  if (!parsed.success) throw new Error("El servidor devolvió una lista de citas inválida.");
  return parsed.data.data;
}

export async function getAppointment(id: string): Promise<Appointment> {
  const parsed = appointmentResponseSchema.safeParse(await httpRequest(`/scheduling/appointments/${id}`));
  if (!parsed.success) throw new Error("El servidor devolvió una cita inválida.");
  return parsed.data.data;
}

export async function createAppointment(input: AppointmentInput): Promise<Appointment> {
  const parsed = appointmentResponseSchema.safeParse(await httpRequest("/scheduling/appointments", { method: "POST", body: input }));
  if (!parsed.success) throw new Error("El servidor devolvió una cita inválida.");
  return parsed.data.data;
}

export async function updateAppointment(id: string, input: AppointmentInput): Promise<Appointment> {
  const parsed = appointmentResponseSchema.safeParse(await httpRequest(`/scheduling/appointments/${id}`, { method: "PUT", body: input }));
  if (!parsed.success) throw new Error("El servidor devolvió una cita inválida.");
  return parsed.data.data;
}
