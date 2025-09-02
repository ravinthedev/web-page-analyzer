CREATE TABLE IF NOT EXISTS schema_migrations (
    version BIGINT PRIMARY KEY,
    applied_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

INSERT INTO schema_migrations (version) VALUES (1) ON CONFLICT (version) DO NOTHING;
