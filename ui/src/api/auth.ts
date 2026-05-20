import { apiClient } from "./client";
import type { Token, User } from "../types/api";

export async function register(email: string, password: string): Promise<User> {
  const { data } = await apiClient.post<User>("/auth/register", { email, password });
  return data;
}

export async function login(email: string, password: string): Promise<Token> {
  const { data } = await apiClient.post<Token>("/auth/login", { email, password });
  return data;
}