package aiseat_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// jace_token_test.go — the bot half of ADR 0139 (#2796), in
// resolution_pick_test.go's shape: Empower Jace's "which Jace?" question
// is queued on a real table with two Jace tokens on it, and the real
// heuristic plays every seat until the table is two turns past it. The
// Jace tokens' own loyalty abilities are on offer to the seat that owns
// them the whole time, so a refused loyalty activation would wedge the
// run too.
func TestHeuristicAnswersEmpowerJaceAndPlaysPastTheJaceTokens(t *testing.T) {
	resolutionPickBotRun(t, func(t *testing.T, g *game.Game, seat, _ *game.Player, answers *int) {
		t.Helper()
		var jaces []uuid.UUID
		g.WithWriteLock(func() {
			for i := 0; i < 2; i++ {
				ids, err := g.CreateTokensForEffect(seat.ID, effects.JaceToken(), 1, game.TokenEntryOptions{})
				if err != nil || len(ids) != 1 {
					t.Fatalf("create a Jace token: %v %v", ids, err)
				}
				if err := g.AddCounterForEffect(ids[0], game.CounterLoyalty, 3); err != nil {
					t.Fatalf("loyalty: %v", err)
				}
				jaces = append(jaces, ids[0])
			}
			err := effects.EmpowerJace{N: 2, Then: func(*effects.Context) error {
				*answers++
				return nil
			}}.Apply(effects.NewContext(g, &game.StackItem{
				Kind: game.StackItemSpell, Controller: seat.ID, Owner: seat.ID,
			}))
			if err != nil {
				t.Fatalf("EmpowerJace: %v", err)
			}
		})
		if len(jaces) != 2 {
			t.Fatal("setup: two Jace tokens were not made")
		}
	})
}
