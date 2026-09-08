CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE entities (
    id TEXT PRIMARY KEY,
    role TEXT NOT NULL,

    -- Validations to prevent impossible game states
    level INT NOT NULL DEFAULT 1 CHECK (level >= 1),
    hp INT NOT NULL CHECK (hp >= 0),
    max_hp INT NOT NULL CHECK (max_hp > 0),
    pos_x INT NOT NULL,
    pos_y INT NOT NULL,

    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CHECK (hp <= max_hp)
);

CREATE TABLE memories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entity_id TEXT NOT NULL REFERENCES entities(id) ON DELETE CASCADE,
    memory_text TEXT NOT NULL,

    -- 768 dimensions: this project standardises ALL embeddings on Gemini's
    -- gemini-embedding-001, regardless of which provider generated the NPC's
    -- decision text. See embeddings/gemini.go. If you ever switch the fixed
    -- embedding provider to one with a different output size, this column
    -- (and every row in it) needs a fresh migration + re-embedding pass —
    -- you cannot mix dimensions in one column.
    embedding vector(768),

    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Crucial for fast lookups of a specific NPC's memories
CREATE INDEX idx_memories_entity ON memories(entity_id);

-- HNSW is superior to IVFFlat for live-updating applications
CREATE INDEX idx_memories_embedding ON memories
USING hnsw (embedding vector_cosine_ops);

-- Trigger to auto-update the timestamp
CREATE OR REPLACE FUNCTION update_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_entities_updated
BEFORE UPDATE ON entities
FOR EACH ROW
EXECUTE FUNCTION update_timestamp();
