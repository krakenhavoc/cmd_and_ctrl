package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// autotap_restricted_test.go — #3874, ADR 0040 amendment 2026-10-16:
// auto-pay plans Cavern of Souls' and Delighted Halfling's coloured
// halves for a spell their restriction admits, and the spell can't be
// countered; for any other spell they are left alone. The engine half
// (the candidate rule, its ranking, the payability bound) is
// game/autotap_restricted_test.go; the enumerator's half is
// legal/restricted_autopay_test.go.

var AutoPay = game.CastSpellParams{Strict: true, AutoTap: true, FromZone: "hand"}

// restrictedBasic puts a basic land on the battlefield.
func RestrictedBasic(t *testing.T, g *game.Game, me *game.Player, name string) uuid.UUID {
	t.Helper()
	ids, err := g.SpawnCards(me.ID, me.ID, game.ZoneBattlefield, game.Card{Name: name, TypeLine: "Basic Land — " + name}, 1)
	if err != nil || len(ids) != 1 {
		t.Fatalf("spawn %s: %v", name, err)
	}
	return ids[0]
}
