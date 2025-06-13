-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create enum types for status
CREATE TYPE node_status AS ENUM ('active', 'inactive', 'maintenance', 'failed');
CREATE TYPE shard_status AS ENUM ('active', 'migrating', 'failed', 'maintenance');

-- Nodes table
CREATE TABLE nodes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    location VARCHAR(255) NOT NULL,
    capacity BIGINT NOT NULL,
    status node_status NOT NULL DEFAULT 'active',
    last_heartbeat TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    current_load BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Shards table
CREATE TABLE shards (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    type VARCHAR(50) NOT NULL,
    size BIGINT NOT NULL,
    node_id UUID REFERENCES nodes(id) ON DELETE SET NULL,
    status shard_status NOT NULL DEFAULT 'active',
    version INTEGER NOT NULL DEFAULT 1,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Shard version history
CREATE TABLE shard_versions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    shard_id UUID NOT NULL REFERENCES shards(id) ON DELETE CASCADE,
    version INTEGER NOT NULL,
    type VARCHAR(50) NOT NULL,
    size BIGINT NOT NULL,
    node_id UUID REFERENCES nodes(id) ON DELETE SET NULL,
    status shard_status NOT NULL,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(shard_id, version)
);

-- Shard migration history
CREATE TABLE shard_migrations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    shard_id UUID NOT NULL REFERENCES shards(id) ON DELETE CASCADE,
    from_node_id UUID REFERENCES nodes(id) ON DELETE SET NULL,
    to_node_id UUID REFERENCES nodes(id) ON DELETE SET NULL,
    status VARCHAR(50) NOT NULL,
    started_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP WITH TIME ZONE,
    error_message TEXT
);

-- Policies table
CREATE TABLE policies (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    policy_type VARCHAR(50) NOT NULL,
    parameters JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Failure reports table
CREATE TABLE failure_reports (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    type VARCHAR(50) NOT NULL,
    entity_id UUID NOT NULL,
    details JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes
CREATE INDEX idx_shards_node_id ON shards(node_id);
CREATE INDEX idx_shards_status ON shards(status);
CREATE INDEX idx_shards_version ON shards(version);
CREATE INDEX idx_nodes_status ON nodes(status);
CREATE INDEX idx_nodes_last_heartbeat ON nodes(last_heartbeat);
CREATE INDEX idx_shard_migrations_shard_id ON shard_migrations(shard_id);
CREATE INDEX idx_shard_versions_shard_id ON shard_versions(shard_id);
CREATE INDEX idx_failure_reports_entity_id ON failure_reports(entity_id);

-- Create updated_at trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for updated_at
CREATE TRIGGER update_nodes_updated_at
    BEFORE UPDATE ON nodes
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_shards_updated_at
    BEFORE UPDATE ON shards
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_policies_updated_at
    BEFORE UPDATE ON policies
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column(); 