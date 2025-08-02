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

-- Create an index on commonly queried fields
CREATE INDEX idx_connections_host_port ON connections(postgres_host, postgres_port);

-- Update trigger for updated_at
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

-- Create indexes on commonly queried fields
CREATE INDEX idx_destinations_name ON destinations(name);
CREATE INDEX idx_destinations_endpoint ON destinations(endpoint_url);
CREATE INDEX idx_destinations_bucket ON destinations(bucket_name);
CREATE INDEX idx_destinations_connection_id ON destinations(connection_id);

-- Update trigger for updated_at
CREATE TRIGGER update_destinations_updated_at 
    BEFORE UPDATE ON destinations 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();