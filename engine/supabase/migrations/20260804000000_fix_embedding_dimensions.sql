git log --format=fuller -1-- Fixes the "expected 1536 dimensions, not 3072" runtime error.
--
-- What actually happened: embeddings/gemini.go was never sending Gemini the
-- `output_dimensionality` field, so gemini-embedding-001 always returned its
-- native 3072-dim vector — regardless of what any `vector(N)` column expected.
-- Meanwhile the `memories.embedding` column in THIS database is currently
-- vector(1536) (likely set up by hand at some point, or left over from an
-- earlier embedding model), while the original migration/comment in
-- 20260731155306_init_schema.sql documents the intended size as vector(768).
--
-- This migration makes 768 the one real answer, matching:
--   1) embeddings/gemini.go's `Dimensions` constant (now also sends
--      output_dimensionality=768 in the API request, see that file), and
--   2) the original init_schema.sql column comment.
-- 768 is also the cheapest/fastest of Gemini's three recommended sizes
-- (768/1536/3072) with only ~0.26% quality loss vs. the full 3072 — a good
-- trade for a demo world with a handful of NPCs.
--
-- Existing memory rows cannot be cast between vector sizes (a 1536-dim
-- vector is not a valid 768-dim vector), so this clears stored embeddings
-- rather than failing the migration. The memory_text itself (and the row)
-- is kept; only the embedding is reset to NULL and will be regenerated the
-- next time that NPC forms a new memory. If you'd rather keep 1536 (and
-- request the DB stay that size), tell me and I'll flip this migration and
-- the Go constant the other way instead.

DROP INDEX IF EXISTS idx_memories_embedding;

ALTER TABLE memories
    ALTER COLUMN embedding TYPE vector(768) USING NULL;

CREATE INDEX IF NOT EXISTS idx_memories_embedding ON memories
USING hnsw (embedding vector_cosine_ops);