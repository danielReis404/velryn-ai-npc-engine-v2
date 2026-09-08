-- Adds the fields needed for an NPC to keep its own story across restarts.
-- Without these, a restart brought back correct HP/position but the NPC
-- "forgot" who it was, since background/objective only lived in main.go.

ALTER TABLE entities
    ADD COLUMN IF NOT EXISTS background    TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS objective     TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS xp            INT  NOT NULL DEFAULT 0 CHECK (xp >= 0),
    ADD COLUMN IF NOT EXISTS base_attack   INT  NOT NULL DEFAULT 1 CHECK (base_attack >= 0),
    ADD COLUMN IF NOT EXISTS vision_radius INT  NOT NULL DEFAULT 3 CHECK (vision_radius >= 0);
