package legal_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// granted_alternative_cost_test.go — ADR 0118 §3, #2163, the
// enumerator half: the moves walk game.CastOffersForLocked, so a
// granted offer is a cast a bot is offered, at the price the engine
// charges, and every one is a move CastSpell accepts (dispatchAll).

const (
	oracleTestJodah       = "test-2163-legal-jodah"
	oracleTestOmniscience = "test-2163-legal-omniscience"
)

// withLegalGrantedAltCosts stubs the two test grantors, spelled as
// effects.PayWUBRGForSpellsYouCast and
// effects.CastFromHandWithoutPayingManaCost build them.
func withLegalGrantedAltCosts(t *testing.T) {
	t.Helper()
	prev := game.CatalogGrantedAlternativeCosts
	game.CatalogGrantedAlternativeCosts = func(key string) []game.GrantedAlternativeCost {
		switch key {
		case oracleTestJodah:
			return []game.GrantedAlternativeCost{{Offer: game.AlternativeCost{
				Key: "granted-wubrg", Label: "Pay {W}{U}{B}{R}{G} rather than pay this spell's mana cost",
				ManaCost: "{W}{U}{B}{R}{G}",
			}}}
		case oracleTestOmniscience:
			return []game.GrantedAlternativeCost{{
				Offer: game.AlternativeCost{Key: "granted-free", Label: "Cast it without paying its mana cost"},
				Zones: []game.ZoneKind{game.ZoneHand},
			}}
		}
		return nil
	}
	t.Cleanup(func() { game.CatalogGrantedAlternativeCosts = prev })
}

// The printed {7}{B}{B} is out of reach of five basics, Jodah's
// {W}{U}{B}{R}{G} is not: the enumerator offers exactly that cast, and
// the auto-tapper pays it.
func TestEnumeratorOffersAGrantedWUBRGCast(t *testing.T) {
	withLegalGrantedAltCosts(t)
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	for _, sub := range []string{"Plains", "Island", "Swamp", "Mountain", "Forest"} {
		basicLands(g, seat, 1, sub)
	}
	battlefieldCard(g, seat, game.Card{Name: "Test Jodah", TypeLine: "Legendary Creature — Human Wizard", OracleID: oracleTestJodah})
	bringer := handCard(seat, creature("Test Bringer", "{7}{B}{B}", 6, 6))

	moves := legal.EnumerateFor(g, seat.ID)
	casts := castPayloadsOf(t, moves, bringer)
	if len(casts) != 1 || casts[0].AlternativeCost != "granted-wubrg" {
		t.Fatalf("casts of the Bringer = %+v, want one claiming granted-wubrg: %v", casts, labels(moves))
	}
	dispatchAll(t, g, seat.ID, castMovesFor(moves, bringer))
}

// Omniscience's free cast is offered with no mana at all, and only at
// the spell's own timing: a sorcery is not offered in the draw step.
func TestEnumeratorOffersAGrantedFreeCastAtTheSpellsTiming(t *testing.T) {
	withLegalGrantedAltCosts(t)
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	// newTable parks the cursor on the active seat's draw step.
	battlefieldCard(g, seat, game.Card{Name: "Test Omniscience", TypeLine: "Enchantment", OracleID: oracleTestOmniscience})
	sorcery := handCard(seat, game.Card{Name: "Test Opportunity", TypeLine: "Sorcery", ManaCost: "{9}{U}"})

	early := legal.EnumerateFor(g, seat.ID)
	if len(early) == 0 {
		t.Fatalf("the seat is offered nothing in its draw step; the timing check below would prove nothing")
	}
	if casts := castPayloadsOf(t, early, sorcery); len(casts) != 0 {
		t.Fatalf("a sorcery was offered in the draw step: %+v", casts)
	}
	advanceTo(t, g, game.StepPrecombatMain)
	moves := legal.EnumerateFor(g, seat.ID)
	casts := castPayloadsOf(t, moves, sorcery)
	if len(casts) != 1 || casts[0].AlternativeCost != "granted-free" {
		t.Fatalf("casts of the sorcery = %+v, want one claiming granted-free: %v", casts, labels(moves))
	}
	dispatchAll(t, g, seat.ID, castMovesFor(moves, sorcery))
}
