package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "c11e84a1-dbda-429b-8cd6-fd0deaefc689",
		Name:         "Demand Answers",
		Completeness: effects.CompletenessFull,
		AdditionalCost: effects.EitherCost(
			effects.SacrificeCost("an artifact", effects.Artifact()).Keyed("sacrifice"),
			effects.DiscardCost(1).Keyed("discard"),
		),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.DrawCards{Player: item.Controller, N: 2}.Apply(ctx)
		},
	})
}
