-- Adds the "person, not quest-giver" fields: personality, fixed narrative
-- relationships, faction, likes/dislikes, secrets, named skills, and the
-- two fields that actually change during play — Trust and Fear per other
-- character (see engine.adjustTrust / adjustFear).
--
-- Personality/Relationships/Faction/CombatStyle/Likes/Dislikes/
-- KnownSecrets/Skills/CurrentMood/CurrentGoal are saved here for
-- completeness and so you can inspect/query them in Supabase, but — same
-- as background/objective already — the game only re-hydrates the fields
-- that actually mutate over time (Trust, Fear) back into a running NPC on
-- restart. The others stay defined in main.go, which is their source of
-- truth. See engine.World.HydrateFromDB.

ALTER TABLE entities
    ADD COLUMN IF NOT EXISTS personality    TEXT     NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS faction        TEXT     NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS combat_style   TEXT     NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS current_mood   TEXT     NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS current_goal   TEXT     NOT NULL DEFAULT '',

    ADD COLUMN IF NOT EXISTS relationships  TEXT[]   NOT NULL DEFAULT '{}',
    ADD COLUMN IF NOT EXISTS likes          TEXT[]   NOT NULL DEFAULT '{}',
    ADD COLUMN IF NOT EXISTS dislikes       TEXT[]   NOT NULL DEFAULT '{}',
    ADD COLUMN IF NOT EXISTS known_secrets  TEXT[]   NOT NULL DEFAULT '{}',
    ADD COLUMN IF NOT EXISTS skills         TEXT[]   NOT NULL DEFAULT '{}',

    -- Per-other-character opinion, e.g. {"Lyra": 90, "Kael": 25}. jsonb (not
    -- a separate rows-per-relationship table) because the access pattern is
    -- always "load this NPC's whole opinion map at once", never "query
    -- across NPCs by trust value" — a normalized table would just add joins
    -- for no benefit here.
    ADD COLUMN IF NOT EXISTS trust          JSONB    NOT NULL DEFAULT '{}'::jsonb,
    ADD COLUMN IF NOT EXISTS fear           JSONB    NOT NULL DEFAULT '{}'::jsonb;