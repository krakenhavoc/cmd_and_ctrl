package legal_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// trigger_mana_cast_test.go — ADR 0074's amendment of 2026-10-30 (#4716):
// the enumerator's affordability probe is the auto-tap planner, and the
// planner now counts triggered mana, so a cast only that mana makes
// payable is offered, and the offered cast is accepted.

const oracleCryptGhast = "a3c8d817-7949-4dae-b9f5-f9d952479270"

// Crypt Ghast and three Swamps make six mana, so a {4}{B} spell is
// offered under strict payment and casts; a seven-mana one is not.
func TestCastPayableOnlyWithTriggerManaIsOffered(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	battlefieldCard(g, active, game.Card{
		Name: "Crypt Ghast", TypeLine: "Creature — Spirit", OracleID: oracleCryptGhast, Power: 2, Toughness: 2,
	})
	lands(g, active, "Swamp", "Swamp", 3)
	five := handCard(active, creature("Five Drop", "{4}{B}", 4, 4))
	seven := handCard(active, creature("Seven Drop", "{5}{B}{B}", 7, 7))
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	offered := castMovesFor(moves, five)
	if len(offered) == 0 {
		t.Fatalf("{4}{B} not offered off three Swamps and Crypt Ghast (six mana): %v", labels(moves))
	}
	if got := castMovesFor(moves, seven); len(got) != 0 {
		t.Errorf("{5}{B}{B} offered off six mana: %v", labels(got))
	}
	dispatchAll(t, g, active.ID, offered)
}
