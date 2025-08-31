-- Create analyses table
CREATE TABLE IF NOT EXISTS analyses (
    id UUID PRIMARY KEY,
    url TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    result JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_analyses_url ON analyses(url);

CREATE INDEX IF NOT EXISTS idx_analyses_status ON analyses(status);

CREATE INDEX IF NOT EXISTS idx_analyses_created_at ON analyses(created_at);

COMMENT ON TABLE analyses IS 'Stores web page analysis requests and results';
