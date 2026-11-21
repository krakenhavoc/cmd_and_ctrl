package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Reset — Instant {U}{U}:
//
//	"Cast this spell only during an opponent's turn after their upkeep step.
//	 Untap all lands you control."
//
// The timing is a cast condition: the active player is someone else and the
// turn has moved past its untap and upkeep steps. The untap is the one
// Zacama and Unstoppable Plan share.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4faa5a70-ac0c-4a29-8bec-218a35a29fdf",
		Name:         "Reset",
		Completeness: CompletenessFull,
		CastCondition: func(g *game.Game, controller uuid.UUID, _ game.Card) bool {
			active := activePlayerIDOf(g)
			return active != uuid.Nil && active != controller &&
				g.Turn.Step != game.StepUntap && g.Turn.Step != game.StepUpkeep
		},
		CastConditionLabel: "Cast this spell only during an opponent's turn after their upkeep step.",
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return b16UntapAllYouControlMatching(ctx, ctx.Controller(), func(c game.Card) bool { return c.IsLand() })
		},
	})
}
