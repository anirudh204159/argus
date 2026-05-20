// Types matching the FastAPI backend schemas (app/schemas.py).

export interface User {
  id: number;
  email: string;
  created_at: string;
  last_login_at: string | null;
}

export interface Token {
  access_token: string;
  token_type: string;
}

export interface Source {
  id: number;
  name: string;
  host: string;
  port: number;
  database_name: string;
  replication_user: string;
  status: "disconnected" | "connected" | "error";
  last_error: string | null;
  created_at: string;
}

export interface SourceCreate {
  name: string;
  host: string;
  port: number;
  database_name: string;
  replication_user: string;
  replication_password: string;
}

export type Operation = "INSERT" | "UPDATE" | "DELETE";

export interface Subscription {
  id: number;
  source_id: number;
  name: string;
  tables: string[];
  operations: Operation[];
  webhook_url: string;
  retry_max: number;
  active: boolean;
  created_at: string;
}

export interface SubscriptionCreate {
  source_id: number;
  name: string;
  tables: string[];
  operations: Operation[];
  webhook_url: string;
  retry_max?: number;
}

export interface SubscriptionCreated extends Subscription {
  hmac_secret: string;
}

export interface ApiError {
  detail: string;
}