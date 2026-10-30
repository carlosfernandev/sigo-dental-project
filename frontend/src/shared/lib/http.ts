import { z } from "zod";

const API_BASE_URL = (import.meta.env.VITE_API_BASE_URL ?? "http://localhost:8080").replace(/\/$/, "");

const apiErrorSchema = z.object({
  error: z.object({
    code: z.string().optional(),
    message: z.string().optional(),
    details: z.unknown().optional(),
  }),
});

export class HttpError extends Error {
  readonly status: number;
  readonly code?: string;
  readonly details?: unknown;

  constructor(status: number, message: string, code?: string, details?: unknown) {
    super(message);
    this.name = "HttpError";
    this.status = status;
    this.code = code;
    this.details = details;
  }
}

type RequestOptions = Omit<RequestInit, "body"> & { body?: unknown };

export async function httpRequest(path: string, options: RequestOptions = {}): Promise<unknown> {
  const headers = new Headers(options.headers);
  if (options.body !== undefined) headers.set("Content-Type", "application/json");

  const accessToken = window.localStorage.getItem("sigo.access_token");
  if (accessToken) headers.set("Authorization", `Bearer ${accessToken}`);

  let response: Response;
  try {
    response = await fetch(`${API_BASE_URL}${path}`, {
      ...options,
      headers,
      body: options.body === undefined ? undefined : JSON.stringify(options.body),
    });
  } catch {
    throw new HttpError(0, "No se pudo conectar con el servidor. Inténtalo nuevamente.");
  }

  const payload: unknown = await response.json().catch(() => null);
  if (!response.ok) {
    if (response.status === 401) window.dispatchEvent(new Event("sigo:unauthorized"));
    const parsedError = apiErrorSchema.safeParse(payload);
    const error = parsedError.success ? parsedError.data.error : undefined;
    throw new HttpError(
      response.status,
      error?.message ?? "No se pudo completar la solicitud.",
      error?.code,
      error?.details,
    );
  }

  return payload;
}
