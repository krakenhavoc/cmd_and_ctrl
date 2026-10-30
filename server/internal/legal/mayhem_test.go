package legal_test

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	_ "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// mayhem_test.go — #4628, CR 702.187. A bot picks from the enumerator's
// list, so a mayhem cast is offered to it exactly while the engine
// accepts one: the card was discarded this turn and its mayhem cost can
// be paid. The enumerator asks the same predicates the engine does
// (cardOpensZoneLocked, validateCastPathLocked), so this pins that no
// second path was needed.

const oracleSadisticSlash = "0e471b6d-4ba4-46aa-a802-81e401cbfa56"

func mayhemSlashInHand(g *game.Game, seat *game.Player) uuid.UUID {
	clearHand(seat)
	return handCard(seat, game.Card{
		Name: "Sadistic Slash", TypeLine: "Instant", ManaCost: "{3}{B}", OracleID: oracleSadisticSlash,
	})
}

func TestEnumeratorOffersMayhemOnlyForACardDiscardedThisTurn(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	advanceTo(t, g, game.StepPrecombatMain)
	basicLands(g, seat, 2, "Swamp")
	battlefieldCard(g, g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)], creature("Bear", "{1}{G}", 2, 2))

	slash := mayhemSlashInHand(g, seat)
	milled := graveyardCard(seat, game.Card{
		Name: "Sadistic Slash", TypeLine: "Instant", ManaCost: "{3}{B}", OracleID: oracleSadisticSlash,
	})
	if n := len(castMovesFor(legal.EnumerateFor(g, seat.ID), milled)); n != 0 {
		t.Errorf("offered %d casts of a card that was never discarded", n)
	}

	g.WithWriteLock(func() {
		if err := g.DiscardRandomForEffect(seat.ID, 1); err != nil {
			t.Errorf("discard: %v", err)
		}
	})
	if !seat.Graveyard.Contains(slash) {
		t.Fatal("fixture: the Slash was not discarded")
	}

	moves := legal.EnumerateFor(g, seat.ID)
	casts := castMovesFor(moves, slash)
	if len(casts) == 0 {
		t.Fatalf("a card discarded this turn is not offered: %v", labels(moves))
	}
	for _, m := range casts {
		if p := string(m.Params); !strings.Contains(p, `"alternative_cost":"mayhem"`) || !strings.Contains(p, `"from_zone":"graveyard"`) {
			t.Errorf("move %q params = %v, want a mayhem cast from the graveyard", m.Label, m.Params)
		}
	}
	// The contract: every offered move is one the dispatcher accepts.
	dispatchAll(t, g, seat.ID, casts)
	// And the milled copy beside it is still not offered.
	if n := len(castMovesFor(moves, milled)); n != 0 {
		t.Errorf("offered %d casts of the card that was never discarded", n)
	}

	// One mana short of {1}{B}: nothing to offer.
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].Controller == seat.ID && g.Battlefield.Cards[i].Name == "Swamp 1" {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	if n := len(castMovesFor(legal.EnumerateFor(g, seat.ID), slash)); n != 0 {
		t.Errorf("offered %d mayhem casts with one land untapped", n)
	}
}
