package legal_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// imposed_sacrifice_test.go — #6205: a sacrifice the board adds to a
// spell's cost (Drought's "Sacrifice a Swamp" per black symbol, CR
// 601.2f) is folded into the enumerator's payment exactly as CastSpell
// folds it, so every offered cast names the Swamps and is one the engine
// accepts (#544).

func TestEnumeratorNamesTheSwampsDroughtImposes(t *testing.T) {
	const drought = "test-drought"
	prev := game.CatalogImposedSacrifices
	game.CatalogImposedSacrifices = func(key string) []game.ImposedSacrifice {
		if key == drought {
			return []game.ImposedSacrifice{{
				Label:  "Spells cost an additional \"Sacrifice a Swamp\" to cast for each black mana symbol in their mana costs.",
				Kind:   game.SacrificeKind{Label: "a Swamp", Subtypes: []string{"Swamp"}},
				Spells: true, PerBlackSymbol: true,
			}}
		}
		if prev != nil {
			return prev(key)
		}
		return nil
	}
	t.Cleanup(func() { game.CatalogImposedSacrifices = prev })

	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	battlefieldCard(g, g.Seats[1], game.Card{Name: "Drought", TypeLine: "Enchantment", OracleID: drought})
	basicLands(g, seat, 4, "Swamp")
	shade := handCard(seat, creature("Twin Shade", "{B}{B}", 2, 2))

	moves := legal.EnumerateFor(g, seat.ID)
	pays := varSacMovesOf(t, moves, shade)
	if len(pays) == 0 {
		t.Fatal("the {B}{B} creature is not offered under Drought with four Swamps")
	}
	for _, p := range pays {
		if len(p.SacrificeIDs) != 2 {
			t.Fatalf("sacrifice_ids = %v, want two Swamps", p.SacrificeIDs)
		}
	}
	dispatchAll(t, g, seat.ID, moves)

	// Two Swamps can't both pay {B}{B} and be sacrificed: no cast.
	g2 := newTable(t)
	seat2 := g2.Seats[g2.Turn.ActiveSeat]
	clearHand(seat2)
	advanceTo(t, g2, game.StepPrecombatMain)
	battlefieldCard(g2, g2.Seats[1], game.Card{Name: "Drought", TypeLine: "Enchantment", OracleID: drought})
	basicLands(g2, seat2, 1, "Swamp")
	shade2 := handCard(seat2, creature("Twin Shade", "{B}{B}", 2, 2))
	seat2.ManaPool.AddMana(game.ManaToken{Color: "B"}, game.ManaToken{Color: "B"})
	if pays := varSacMovesOf(t, legal.EnumerateFor(g2, seat2.ID), shade2); len(pays) != 0 {
		t.Fatalf("offered %d casts with one Swamp for \"Sacrifice a Swamp\" twice", len(pays))
	}
}
