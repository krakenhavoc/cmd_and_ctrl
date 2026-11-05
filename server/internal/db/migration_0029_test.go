package db

import (
	"context"
	"testing"
)

// Tests for migration 0029 (game statistics version 2, ADR 0217 §1): a
// v28 database keeps its version-1 records, which read NULL in every new
// column; the new tables take rows, refuse values outside their CHECKs
// and go with their game; and the hand rollback in docs/environments.md.

const statsV2Version = 29

func TestMigration0029KeepsVersionOneRecords(t *testing.T) {
	d := openAtVersion(t, statsV2Version-1)
	seedStats(t, d)
	if err := migrate(context.Background(), d); err != nil {
		t.Fatalf("migrate v28 -> latest: %v", err)
	}
	var (
		frames, winKind any
		deckHash, own   any
	)
	if err := d.QueryRow(`SELECT frames_complete, win_kind FROM game_stats WHERE game_id = 'g1'`).Scan(&frames, &winKind); err != nil {
		t.Fatal(err)
	}
	if err := d.QueryRow(`SELECT deck_hash, own_turns FROM seat_stats WHERE game_id = 'g1'`).Scan(&deckHash, &own); err != nil {
		t.Fatal(err)
	}
	if frames != nil || winKind != nil || deckHash != nil || own != nil {
		t.Errorf("a version-1 record reads %v %v %v %v, want NULL in every new column", frames, winKind, deckHash, own)
	}
	for _, table := range []string{"seat_turn_stats", "seat_target_stats", "seat_card_seen"} {
		if n := countRows(t, d, table); n != 0 {
			t.Errorf("%s has %d rows after the migration, want 0", table, n)
		}
	}
}

func TestMigration0029TablesGoWithTheirGame(t *testing.T) {
	d := openAtVersion(t, statsV2Version)
	seedStats(t, d)
	mustExec(t, d, `UPDATE game_stats SET frames_complete = 1, win_kind = 'combo', win_turn = 9, combo_rule = 'loop'`)
	mustExec(t, d, `UPDATE seat_stats SET out_kind = 'combat', deck_hash = 'ed05af03e6000a4c', screw_game = 1`)
	mustExec(t, d, `INSERT INTO seat_turn_stats (game_id, seat, own_turn, turn, land_played, land_drops, lands_controlled,
		mana_sources, hand_size, lands_in_hand, spells_in_hand, missed, screw, flood, colour_screw, attackers, eligible,
		damage_to_players, spells_cast) VALUES ('g1', 0, 1, 1, 1, 1, 1, 1, 6, 2, 4, 0, 0, 0, 0, 0, 0, 0, 0)`)
	mustExec(t, d, `INSERT INTO seat_target_stats (game_id, seat, target_seat, attacks, damage, combat_damage, commander_damage)
		VALUES ('g1', 0, 1, 3, 12, 10, 7)`)
	mustExec(t, d, `INSERT INTO seat_card_seen (game_id, seat, name, copies, in_opening_hand, first_seen_turn, held_turns,
		affordable_turns, "cast") VALUES ('g1', 0, 'Sol Ring', 1, 1, 0, 2, 2, 1)`)
	for _, q := range []string{
		`UPDATE game_stats SET win_kind = 'luck'`,
		`UPDATE game_stats SET combo_rule = 'vibes'`,
		`UPDATE game_stats SET frames_complete = 2`,
		`UPDATE seat_stats SET out_kind = 'luck'`,
		`UPDATE seat_stats SET flood_game = 5`,
		`UPDATE seat_turn_stats SET screw = 2`,
		`UPDATE seat_card_seen SET "cast" = 3`,
		// A row for a seat with no seat line.
		`INSERT INTO seat_target_stats (game_id, seat, target_seat, attacks, damage, combat_damage, commander_damage) VALUES ('g1', 3, 0, 0, 0, 0, 0)`,
	} {
		if _, err := d.Exec(q); err == nil {
			t.Errorf("accepted: %s", q)
		}
	}
	mustExec(t, d, `DELETE FROM games WHERE id = 'g1'`)
	for _, table := range []string{"seat_turn_stats", "seat_target_stats", "seat_card_seen"} {
		if n := countRows(t, d, table); n != 0 {
			t.Errorf("%s kept %d rows after its game was deleted", table, n)
		}
	}
}

func TestMigration0029RollbackByHand(t *testing.T) {
	ctx := context.Background()
	d := openAtVersion(t, statsV2Version)
	seedStats(t, d)
	for _, q := range []string{
		`DROP TABLE seat_card_seen`,
		`DROP TABLE seat_target_stats`,
		`DROP TABLE seat_turn_stats`,
		`ALTER TABLE game_stats DROP COLUMN frames_complete`,
		`ALTER TABLE game_stats DROP COLUMN win_kind`,
		`ALTER TABLE game_stats DROP COLUMN win_turn`,
		`ALTER TABLE game_stats DROP COLUMN combo_rule`,
	} {
		mustExec(t, d, q)
	}
	for _, c := range []string{"turn_position", "deck_hash", "out_kind", "out_by_seat", "own_turns", "land_drops_hit",
		"land_drops_missed", "screw_turns", "flood_turns", "colour_screw_turns", "screw_game", "flood_game",
		"colour_screw_game", "creatures_eligible", "creatures_attacked", "commander_casts", "commander_first_cast_turn",
		"life_lost_to_combat"} {
		mustExec(t, d, `ALTER TABLE seat_stats DROP COLUMN `+c)
	}
	mustExec(t, d, `DELETE FROM schema_migrations WHERE version = 29`)
	if v, _ := currentSchemaVersion(ctx, d); v != statsV2Version-1 {
		t.Fatalf("after the rollback the version is %d, want %d", v, statsV2Version-1)
	}
	if n := countRows(t, d, "seat_stats"); n != 1 {
		t.Errorf("the rollback lost the version-1 record: %d seat lines", n)
	}
	if err := migrate(ctx, d); err != nil {
		t.Fatalf("rolling forward again: %v", err)
	}
	if n := countRows(t, d, "seat_card_seen"); n != 0 {
		t.Errorf("after the round trip seat_card_seen has %d rows, want 0", n)
	}
}
