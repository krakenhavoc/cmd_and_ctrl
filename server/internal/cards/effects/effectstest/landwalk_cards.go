package effectstest

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// blockRefusal declares `blocker` against `attacker` and returns the
// engine's reason, "" for an accepted block. It runs on a clone, so a
// test can ask several times about the same pair.
func BlockRefusal(t *testing.T, g *game.Game, blocker, attacker uuid.UUID) game.BlockReason {
	t.Helper()
	err := g.Clone().DeclareBlocker(blocker, attacker)
	if err == nil {
		return ""
	}
	var br *game.BlockRefusedError
	if !errors.As(err, &br) || !errors.Is(err, game.ErrIllegalBlock) {
		t.Fatalf("DeclareBlocker = %v, want nil or a block refusal", err)
	}
	return br.Reason
}

// seatsForCombat returns the active seat and the next one, which the
// tests attack.
func SeatsForCombat(g *game.Game) (me, them *game.Player) {
	return g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

func AttackIntoBlocks(t *testing.T, g *game.Game, defender uuid.UUID, attackers ...uuid.UUID) {
	t.Helper()
	AdvanceToStepInTurn(t, g, game.StepDeclareAttackers)
	for _, a := range attackers {
		if err := g.DeclareAttacker(a, defender); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	AdvanceToStepInTurn(t, g, game.StepDeclareBlockers)
}
