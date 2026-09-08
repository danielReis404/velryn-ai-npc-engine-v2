-- Adds gold and inventory as persisted, mutable state.
--
-- What was actually happening without this: BUY/SELL/TRADE (see
-- engine.resolveIntent) updated npc.Gold and npc.Inventory correctly in
-- memory for the lifetime of one process run — the trade genuinely
-- happened, gold and items really moved between characters — but neither
-- field was ever written to Postgres. Every SaveNPC call silently dropped
-- them. On the NEXT run, HydrateFromDB restored Level/HP/Position/etc. from
-- the database, but Gold and Inventory came back from main.go's hardcoded
-- starting values instead — so a completed trade (e.g. Goliath selling his
-- one Echo-touched Shard to Cedric for 150 gold) would silently reset:
-- Goliath has the shard again, Cedric has his starting gold again, and
-- their own persisted memories/dialogue history (which DO survive) say the
-- deal is already done, contradicting their actual restored state — which
-- is exactly the kind of mismatch that made the negotiation loop forever
-- across sessions.
--
-- inventory is JSONB (an array of the same shape as models.Item — Name,
-- Power, Durability, MaxDurability, Kind, HealAmount, Price) for the same
-- reason trust/fear are JSONB: the access pattern is always "load this
-- NPC's whole inventory at once", never a cross-entity item query.

ALTER TABLE entities
    ADD COLUMN IF NOT EXISTS gold      INT   NOT NULL DEFAULT 0 CHECK (gold >= 0),
    ADD COLUMN IF NOT EXISTS inventory JSONB NOT NULL DEFAULT '[]'::jsonb;
