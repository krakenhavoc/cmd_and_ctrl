package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// additional_cost_test.go — S21 sub-PR 5: the "As an additional cost
// to cast this spell, discard a card" trio. The interesting cases
// aren't the card draw, they're the ordering: the discard is part of
// casting, so the payoffs trigger above the spell and resolve first.

const (
	ThrillOfPossibilityOracle = "1cb0610b-a731-42c2-b93f-0a29f63cebf4"
	BigScoreOracle            = "a5cbd257-c836-493e-bb1a-76242619dea2"
	UnexpectedWindfallOracle  = "498c10c9-253d-4b15-b48c-1509381b17e8"
)

// castWithDiscard seeds a card into the active seat's hand along
// with `discards` cards of the given type line, walks to a main
// phase, and casts the spell paying those cards. Returns the spell
// and the discarded IDs.
func CastWithDiscard(t *testing.T, g *game.Game, name, oracle, discardType string, discards int) (uuid.UUID, []uuid.UUID) {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	paid := make([]uuid.UUID, 0, discards)
	for i := 0; i < discards; i++ {
		paid = append(paid, HandCard(active, "Fodder", discardType))
	}
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: "Instant", OracleID: oracle,
		Owner: active.ID, Controller: active.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{DiscardIDs: paid}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id, paid
}
