import { httpRequest } from "../../shared/lib/http";
import {
  currentUserResponseSchema,
  loginInputSchema,
  loginResponseSchema,
  type LoginInput,
} from "./schemas";

export async function login(input: LoginInput) {
  const request = loginInputSchema.parse(input);
  const response = await httpRequest("/auth/login", { method: "POST", body: request });
  return loginResponseSchema.parse(response).data;
}

export async function getCurrentUser() {
  const response = await httpRequest("/auth/me");
  return currentUserResponseSchema.parse(response).data;
}
