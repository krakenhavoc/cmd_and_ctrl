package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// autotap_source_wish_test.go — #1212 × #1215, on the card the two
// were written about.
//
// #1212 gave Hired Hexblade a `WantsManaFrom` wish and declared it
// INERT in the card's own header: the auto-tapper refused every
// sacrifice-cost mana ability, so a Treasure was never a candidate and
// the hint could not reach the one family it exists for. #1215 made
// the Treasure plannable and then put it in a LAST-RESORT tier, which
// would have left the hint inert a second way — the planner would
// route around the Treasure to protect it and the Hexblade would
// never draw.
//
// The resolution is one line of ordering: the wish is read before the
// tier. The card asked for Treasure mana, so the planner cracks the
// Treasure. These tests are what stops either half from being quietly
// reordered back.

// hexbladeBoard seats a player at precombat main with one Swamp per
// `swamps` and one real Treasure token, and a Hired Hexblade in hand.
func hexbladeBoard(t *testing.T, swamps int) (*game.Game, *game.Player, uuid.UUID, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)

	if _, err := g.SpawnCards(me.ID, me.ID, game.ZoneBattlefield, game.Card{
		Name: "Swamp", TypeLine: "Basic Land — Swamp",
	}, swamps); err != nil {
		t.Fatalf("spawn Swamps: %v", err)
	}
	tmpl, ok := Tokens().TokenTemplate("Treasure")
	if !ok {
		t.Fatal("no Treasure template")
	}
	ids, err := g.SpawnCards(me.ID, me.ID, game.ZoneBattlefield, tmpl, 1)
	if err != nil || len(ids) != 1 {
		t.Fatalf("spawn Treasure: %v (%d ids)", err, len(ids))
	}

	hexblade := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{
			InstanceID: hexblade, Name: "Hired Hexblade", TypeLine: "Creature — Elf Warlock",
			ManaCost: "{1}{B}", OracleID: hiredHexbladeOracle, Owner: me.ID, Controller: me.ID,
		})
	})
	return g, me, hexblade, ids[0]
}

// The whole composition, end to end and through the real card. Two
// Swamps and a Treasure pay {1}{B} either way, so the generic pip is
// a free choice between the second Swamp and the Treasure — and the
// Hexblade's own text is what decides it. The auto-tapper cracks the
// Treasure, the trigger's intervening if (CR 603.4) sees Treasure
// mana in the record, and the card draws.
//
// Before #1215 this cast did not happen at all: the Treasure was not
// a candidate, and with only one other Swamp on board the strict gate
// reported insufficient mana on a board that pays.
func TestAutoTapCracksTheTreasureHiredHexbladeAsksFor(t *testing.T) {
	g, me, hexblade, treasure := hexbladeBoard(t, 2)
	handBefore, lifeBefore := me.Hand.Size(), me.Life

	if err := g.CastSpell(me.ID, hexblade, game.CastSpellParams{
		Strict: true, AutoTap: true, FromZone: "hand",
	}); err != nil {
		t.Fatalf("auto-tap cast of Hired Hexblade: %v", err)
	}
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(treasure) {
		t.Error("the auto-tapper left the Treasure alone — the wish must outrank the last-resort tier")
	}
	// handBefore counts the Hexblade, which has left; a draw of one
	// puts the count back where it started.
	if me.Hand.Size() != handBefore {
		t.Errorf("hand %d, want %d — the Hexblade left and one card was drawn",
			me.Hand.Size(), handBefore)
	}
	if lifeBefore-me.Life != 1 {
		t.Errorf("lost %d life, want 1 — the draw and the loss are one clause", lifeBefore-me.Life)
	}
}

// The same board and the same auto-tap, with a card that asks for
// nothing: the last-resort tier is back in charge and the Treasure
// survives. This is the half that proves the ordering is a WISH and
// not a new blanket preference for cracking Treasures.
func TestAutoTapSparesTheTreasureForACardThatDoesNotAskForIt(t *testing.T) {
	g, me, _, treasure := hexbladeBoard(t, 2)

	plain := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{
			InstanceID: plain, Name: "Plain Warlock", TypeLine: "Creature — Elf Warlock",
			ManaCost: "{1}{B}", Owner: me.ID, Controller: me.ID,
		})
	})
	if err := g.CastSpell(me.ID, plain, game.CastSpellParams{
		Strict: true, AutoTap: true, FromZone: "hand",
	}); err != nil {
		t.Fatalf("auto-tap cast of a wishless creature: %v", err)
	}
	if !g.Battlefield.Contains(treasure) {
		t.Error("the auto-tapper cracked a Treasure for a card that never asked for one — " +
			"two Swamps could pay {1}{B}")
	}
}
