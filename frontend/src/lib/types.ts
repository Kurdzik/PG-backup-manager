
export interface BackupDestination {
  id: number;
  connection_id: number;
  name: string;
  endpoint_url: string;
  region: string;
  bucket_name: string;
  access_key_id: string;
  secret_access_key: string;
  path_prefix: string;
  use_ssl: boolean;
  verify_ssl: boolean;
  created_at: string;
  updated_at: string;
}

export interface DatabaseConnection {
  id: number;
  postgres_db_name: string;
  postgres_host: string;
  postgres_port: string;
  postgres_user: string;
  postgres_password: string;
  created_at: string;
  updated_at: string;
}