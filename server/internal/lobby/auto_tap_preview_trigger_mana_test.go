package lobby

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// auto_tap_preview_trigger_mana_test.go — ADR 0074's amendment of
// 2026-10-30 (#4716), item 7: the preview returns the plan itself, so once
// the planner counts triggered mana the preview names the taps the
// payment will make, before anything is tapped. No wire field changes
// (owner decision 3).

const previewCryptGhastOracle = "a3c8d817-7949-4dae-b9f5-f9d952479270"

// The #4716 board: Crypt Ghast, five Swamps and a Treasure, previewing a
// {3}{B}{B}{B} spell. The plan is three Swamps, and the Treasure is in
// none of the sources.
func TestAutoTapPreviewCountsCryptGhastsMana(t *testing.T) {
	f := newPreviewFixture(t)
	spell := f.spawn(f.alice, game.ZoneHand, game.Card{
		Name: "Black Six-Drop", TypeLine: "Creature — Horror", ManaCost: "{3}{B}{B}{B}",
	}, 1)[0]
	f.spawn(f.alice, game.ZoneBattlefield, game.Card{
		Name: "Crypt Ghast", TypeLine: "Creature — Spirit", OracleID: previewCryptGhastOracle,
	}, 1)
	swamps := f.spawn(f.alice, game.ZoneBattlefield, game.Card{Name: "Swamp", TypeLine: "Basic Land — Swamp"}, 5)
	treasure := f.spawn(f.alice, game.ZoneBattlefield, game.Card{
		Name: "Treasure", TypeLine: "Token Artifact — Treasure", TokenKey: game.TokenKey("treasure"),
	}, 1)[0]

	ok, plan, sources := f.previewSources(spell)
	if !ok || len(plan) != 3 {
		t.Fatalf("preview for {3}{B}{B}{B} under Crypt Ghast: ok=%v plan=%v, want three Swamps", ok, plan)
	}
	isSwamp := map[string]bool{}
	for _, id := range swamps {
		isSwamp[id.String()] = true
	}
	for _, id := range plan {
		if !isSwamp[id] {
			t.Errorf("plan names %s, which is not a Swamp", id)
		}
	}
	for _, s := range sources {
		if s.CardID == treasure.String() {
			t.Errorf("the preview spends the Treasure (%+v), but three Swamps pay six mana", s)
		}
	}
}
