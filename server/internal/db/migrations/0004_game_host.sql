-- Migration 0004: the table host (ADR 0075 §2.1, S35, #1032).
--
--   host_player_id  the engine Player.ID of the seat that hosts the
--                   table (may manage it alongside the server admin).
--                   NULL: nobody hosts — no human has sat down yet, or
--                   none is left in the game.
--   host_discord_id a named host still waiting to claim a seat: the
--                   Discord snowflake POST /games was given (the
--                   /cc-invite invoker). NULL once that identity sits
--                   down, after an explicit transfer, or when no host
--                   was named. Never served over the API.
--
-- Plain ADD COLUMNs: neither is a foreign key. host_player_id names an
-- engine player, which has no table (seats.player_id is not unique on
-- its own), and host_discord_id is a pending pointer exactly like
-- seats.pending_discord_id. When games.created_by starts seeding the
-- host (ADR 0075 §2.1 "S34 migration"), it does so at claim time and
-- needs nothing from this schema.
--
-- Every existing row gets NULL in both, which is what "no host yet"
-- means; the first human join or a transfer fills it in.

ALTER TABLE games ADD COLUMN host_player_id TEXT;
ALTER TABLE games ADD COLUMN host_discord_id TEXT;
