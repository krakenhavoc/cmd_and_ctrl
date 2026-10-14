package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// granted_cap_test.go — #3503, the enumerator half of a granted offer
// whose mana value cap is computed from the board (Omnipresence): the
// moves walk game.CastOffersForLocked, so the free cast is a move
// exactly while the creature count covers the spell, and every one is a
// move CastSpell accepts.

const oracleTestOmnipresence = "test-3503-legal-omnipresence"

func withLegalCappedFreeCast(t *testing.T) {
	t.Helper()
	prev := game.CatalogGrantedAlternativeCosts
	game.CatalogGrantedAlternativeCosts = func(key string) []game.GrantedAlternativeCost {
		if key != oracleTestOmnipresence {
			return nil
		}
		return []game.GrantedAlternativeCost{{
			Offer: game.AlternativeCost{Key: "granted-free-up-to-creature-count", Label: "Cast it without paying its mana cost"},
			Zones: []game.ZoneKind{game.ZoneHand},
			MaxManaValueFunc: func(g *game.Game, caster uuid.UUID) int {
				n := 0
				for i := range g.Battlefield.Cards {
					if c := &g.Battlefield.Cards[i]; c.Controller == caster && c.IsCreature() {
						n++
					}
				}
				return n
			},
			MaxManaValueText: "the number of creatures you control",
		}}
	}
	t.Cleanup(func() { game.CatalogGrantedAlternativeCosts = prev })
}

func TestEnumeratorOffersTheCappedFreeCastOnlyWhileTheCountCoversIt(t *testing.T) {
	withLegalCappedFreeCast(t)
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	battlefieldCard(g, seat, game.Card{Name: "Test Omnipresence", TypeLine: "Enchantment", OracleID: oracleTestOmnipresence})
	spell := handCard(seat, game.Card{Name: "Test Opportunity", TypeLine: "Sorcery", ManaCost: "{8}{U}"})

	// Nine mana value, no creatures, no mana: no cast of any kind.
	if casts := castPayloadsOf(t, legal.EnumerateFor(g, seat.ID), spell); len(casts) != 0 {
		t.Fatalf("a cast was offered with no creatures and no mana: %+v", casts)
	}
	for i := 0; i < 8; i++ {
		battlefieldCard(g, seat, creature("Test Beast", "{1}", 1, 1))
	}
	if casts := castPayloadsOf(t, legal.EnumerateFor(g, seat.ID), spell); len(casts) != 0 {
		t.Fatalf("a mana value 9 spell was offered for free with eight creatures: %+v", casts)
	}
	battlefieldCard(g, seat, creature("Test Beast", "{1}", 1, 1))
	moves := legal.EnumerateFor(g, seat.ID)
	casts := castPayloadsOf(t, moves, spell)
	if len(casts) != 1 || casts[0].AlternativeCost != "granted-free-up-to-creature-count" {
		t.Fatalf("casts with nine creatures = %+v, want one claiming the capped free offer: %v", casts, labels(moves))
	}
	dispatchAll(t, g, seat.ID, castMovesFor(moves, spell))
}
