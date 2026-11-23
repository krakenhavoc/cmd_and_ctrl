package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// temples_test.go — the scry-land cycle, and the CR 614
// self-replacement it needed.
//
// The assertion that earns its keep is the difference between a
// replacement and the OnETB-tap workaround it replaces. Both leave the
// permanent tapped, so "is it tapped?" cannot tell them apart. What can:
// an OnETB tap emits EventTapCard, because it really does tap an
// untapped permanent. A replacement emits none — the permanent was
// never untapped on the battlefield.

const TempleOfSilenceOracle = "e6e6fce8-0f6a-4b84-865e-d4e4a4182f9f"

// playLandFromHand puts a land in the active player's hand and plays it
// through the real cast path, so the replacement pipeline runs. Pushing
// straight onto the battlefield would skip exactly what's under test.
func PlayLandFromHand(t *testing.T, g *game.Game, name, oracleID string) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id,
		Name:       name,
		TypeLine:   "Land",
		OracleID:   oracleID,
		Owner:      active.ID,
		Controller: active.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	// #500: the engine now enforces CR 305.2's land-play allowance.
	// These are CARD-behaviour fixtures — a fastland wants three
	// lands on the board, Field of the Dead wants eight — not tests
	// of the land-drop rule, and they build their board inside a
	// single main phase. Raising this seat's base allowance buys each
	// helper call its own land play, which is the fixture saying "and
	// another turn's land drop" without spending real turns.
	active.LandDropsPerTurn++
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("play %s: %v", name, err)
	}
	return id
}

// tapEventsFor counts EventTapCard entries for one card.
func TapEventsFor(g *game.Game, id uuid.UUID) int {
	n := 0
	for _, ev := range g.Events {
		if ev.Kind == game.EventTapCard && ev.CardID == id {
			n++
		}
	}
	return n
}
