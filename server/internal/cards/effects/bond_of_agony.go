package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bond of Agony — Sorcery {X}{B} (Dissension, #5598):
//
//	"As an additional cost to cast this spell, pay X life.
//	 Each other player loses X life."
//
// Hatred's PayXLifeCost; "each other player" is every opponent at this
// table (eachOpponentLosesLife).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "8e7bd5ae-005a-47d8-90bc-69f48cd8793d",
		Name:           "Bond of Agony",
		Completeness:   CompletenessFull,
		XMatters:       true,
		AdditionalCost: PayXLifeCost(),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return eachOpponentLosesLife(ctx.Game, item, ctx.X())
		},
	})
}
