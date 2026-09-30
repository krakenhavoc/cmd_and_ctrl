package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Demand Answers — Instant {1}{R}:
//
//	"As an additional cost to cast this spell, sacrifice an artifact or
//	 discard a card. Draw two cards."
//
// The either/or cost (ADR 0100 §2): the caster announces which branch
// on cast_spell's cost_branch and pays it with the spell already on the
// stack, so a discard payoff or an artifact-dies trigger resolves before
// the cards are drawn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c11e84a1-dbda-429b-8cd6-fd0deaefc689",
		Name:         "Demand Answers",
		Completeness: CompletenessFull,
		AdditionalCost: EitherCost(
			SacrificeCost("an artifact", Artifact()).Keyed("sacrifice"),
			DiscardCost(1).Keyed("discard"),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return DrawCards{Player: item.Controller, N: 2}.Apply(ctx)
		},
	})
}
