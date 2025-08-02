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