CREATE TABLE connections (
    id SERIAL PRIMARY KEY,
    postgres_host VARCHAR(255) NOT NULL,
    postgres_port INTEGER NOT NULL DEFAULT 5432,
    postgres_db_name VARCHAR(255) NOT NULL,
    postgres_user VARCHAR(255) NOT NULL,
    postgres_password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(postgres_host, postgres_port, postgres_db_name)
);

CREATE INDEX idx_connections_host_port ON connections(postgres_host, postgres_port);

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_connections_updated_at 
    BEFORE UPDATE ON connections 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE destinations (
    id SERIAL PRIMARY KEY,
    connection_id INTEGER NOT NULL,
    name VARCHAR(255) NOT NULL UNIQUE,
    endpoint_url VARCHAR(500) NOT NULL,
    region VARCHAR(100),
    bucket_name VARCHAR(255) NOT NULL,
    access_key_id VARCHAR(255) NOT NULL,
    secret_access_key VARCHAR(255) NOT NULL,
    path_prefix VARCHAR(500) DEFAULT '',
    use_ssl BOOLEAN DEFAULT TRUE,
    verify_ssl BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(endpoint_url, bucket_name, path_prefix),
    CONSTRAINT fk_destinations_connection 
        FOREIGN KEY (connection_id) 
        REFERENCES connections(id) 
        ON DELETE CASCADE 
        ON UPDATE CASCADE
);

CREATE INDEX idx_destinations_name ON destinations(name);
CREATE INDEX idx_destinations_endpoint ON destinations(endpoint_url);
CREATE INDEX idx_destinations_bucket ON destinations(bucket_name);
CREATE INDEX idx_destinations_connection_id ON destinations(connection_id);

CREATE TRIGGER update_destinations_updated_at 
    BEFORE UPDATE ON destinations 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE backup_schedules (
    id SERIAL PRIMARY KEY,
    connection_id INTEGER NOT NULL,
    destination_id INTEGER NOT NULL,
    schedule VARCHAR(255) NOT NULL,
    enabled BOOLEAN DEFAULT TRUE,
    last_run TIMESTAMP,
    next_run TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(connection_id, destination_id),
    CONSTRAINT fk_backup_schedules_connection 
        FOREIGN KEY (connection_id) 
        REFERENCES connections(id) 
        ON DELETE CASCADE 
        ON UPDATE CASCADE,
    CONSTRAINT fk_backup_schedules_destination 
        FOREIGN KEY (destination_id) 
        REFERENCES destinations(id) 
        ON DELETE CASCADE 
        ON UPDATE CASCADE
);

CREATE INDEX idx_backup_schedules_connection_id ON backup_schedules(connection_id);
CREATE INDEX idx_backup_schedules_destination_id ON backup_schedules(destination_id);
CREATE INDEX idx_backup_schedules_enabled ON backup_schedules(enabled);
CREATE INDEX idx_backup_schedules_next_run ON backup_schedules(next_run);

CREATE TRIGGER update_backup_schedules_updated_at 
    BEFORE UPDATE ON backup_schedules 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();



CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_username ON users(username);

CREATE TRIGGER update_users_updated_at 
    BEFORE UPDATE ON users 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
