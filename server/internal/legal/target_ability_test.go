package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// target_ability_test.go — #1211 through the bot's eyes.
//
// The enumerator is id-agnostic by construction (it walks
// LegalTargets.Cards and emits a TargetCard ref for each entry), so
// the point of this file is to PROVE that rather than assume it: a
// seat holding Stifle with an opponent's trigger on the stack must be
// offered the trigger, and the move it produces must be one the real
// router accepts (#544, in both directions).

const oracleStifle = "b3b00911-ece7-4484-bc36-f211ce72b6cc"

func stifle() game.Card {
	return game.Card{Name: "Stifle", TypeLine: "Instant", ManaCost: "{U}", OracleID: oracleStifle}
}

// announceTriggerForTest puts an opponent's triggered ability on the
// stack and returns its item id.
func announceTriggerForTest(t *testing.T, g *game.Game, controller, source uuid.UUID, label string) uuid.UUID {
	t.Helper()
	if err := g.AnnounceTrigger(controller, source, game.AbilityParams{Label: label}); err != nil {
		t.Fatalf("AnnounceTrigger: %v", err)
	}
	var out uuid.UUID
	g.WithWriteLock(func() {
		for id, it := range g.StackMeta {
			if it != nil && it.Label == label {
				out = id
			}
		}
	})
	if out == uuid.Nil {
		t.Fatalf("the trigger %q is not on the stack", label)
	}
	return out
}

// A trigger on the stack is an offer for a seat holding Stifle, and
// the offer carries the STACK ITEM's id in an ordinary card-kind ref.
func TestABotIsOfferedAnAbilityAsATarget(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	other := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	basicLands(g, seat, 1, "Island")
	rock := battlefieldCard(g, other, game.Card{Name: "Their Rock", TypeLine: "Artifact"})
	ability := announceTriggerForTest(t, g, other.ID, rock, "Their Rock — do a thing")

	stifleID := handCard(seat, stifle())
	moves := legal.EnumerateFor(g, seat.ID)
	ids := targetIDsOf(t, moves, stifleID)
	if len(ids) != 1 || ids[0] != ability.String() {
		t.Fatalf("Stifle's offers = %v, want the one ability item %s", ids, ability)
	}

	// #544: the move the enumerator produced is a move the engine
	// accepts. An ability id that only the bot believed in would be a
	// move the router refuses.
	var found *legal.Move
	for i := range moves {
		if moves[i].Source == stifleID && moves[i].Kind == legal.KindCast {
			found = &moves[i]
		}
	}
	if found == nil {
		t.Fatal("no cast move for Stifle")
	}
	if err := g.CastSpell(seat.ID, stifleID, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: ability}},
	}); err != nil {
		t.Fatalf("the enumerated target was refused by the engine: %v", err)
	}
}

// And the narrow half: with NOTHING but a spell on the stack, Stifle
// is not offered at all — an unfillable clause produces no move
// rather than a move with no target.
func TestABotIsNotOfferedStifleWithNoAbilityOnTheStack(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	other := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	basicLands(g, seat, 1, "Island")
	theirs := handCard(other, game.Card{Name: "Their Trick", TypeLine: "Instant", ManaCost: "{0}"})
	castSpellForTest(t, g, other.ID, theirs, seat.ID)

	stifleID := handCard(seat, stifle())
	moves := legal.EnumerateFor(g, seat.ID)
	if ids := targetIDsOf(t, moves, stifleID); len(ids) != 0 {
		t.Errorf("Stifle was offered a spell: %v", ids)
	}
}
