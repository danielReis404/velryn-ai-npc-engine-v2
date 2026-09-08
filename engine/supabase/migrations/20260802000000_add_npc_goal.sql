-- Lets an NPC's "current errand" (e.g. heading to a tavern) survive a
-- restart, and lets no-AI ticks keep walking toward it (see
-- engine.World.followGoal). NULL means the NPC has no standing goal.

ALTER TABLE entities
    ADD COLUMN IF NOT EXISTS goal_x INT,
    ADD COLUMN IF NOT EXISTS goal_y INT;
