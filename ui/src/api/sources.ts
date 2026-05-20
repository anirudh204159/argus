import { apiClient } from "./client";
import type { Source, SourceCreate } from "../types/api";

export async function listSources(): Promise<Source[]> {
  const { data } = await apiClient.get<Source[]>("/sources");
  return data;
}

export async function getSource(id: number): Promise<Source> {
  const { data } = await apiClient.get<Source>(`/sources/${id}`);
  return data;
}

export async function createSource(payload: SourceCreate): Promise<Source> {
  const { data } = await apiClient.post<Source>("/sources", payload);
  return data;
}

export async function updateSource(
  id: number,
  payload: Partial<SourceCreate>
): Promise<Source> {
  const { data } = await apiClient.patch<Source>(`/sources/${id}`, payload);
  return data;
}

export async function deleteSource(id: number): Promise<void> {
  await apiClient.delete(`/sources/${id}`);
}