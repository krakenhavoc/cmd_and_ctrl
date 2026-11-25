package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Channeled Force — Instant {2}{U}{R} (Ikoria, #6145):
//
//	"As an additional cost to cast this spell, discard X cards.
//	 Target player draws X cards. Channeled Force deals X damage to up to
//	 one target creature or planeswalker."
//
// X is announced with the cast (CR 107.3a) and is the number of cards
// discarded as the cost (effects.DiscardXCost): the discard happens while
// the spell is on the stack (CR 601.2f), and the stack item keeps the X
// for the draw and the damage. Two target clauses (Unstable Experiment's
// shape): slot 0 the player, slot 1 an optional creature or planeswalker,
// each re-checked as the spell resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "e12ed186-01c1-4da5-b90f-cfbb77c4e8d0",
		Name:           "Channeled Force",
		Completeness:   effects.CompletenessFull,
		XMatters:       true,
		AdditionalCost: effects.DiscardXCost(),
		Targets: effects.Clauses(
			effects.TargetPlayer("target player"),
			effects.TargetPermanent("up to one target creature or planeswalker",
				effects.Or(effects.Creature(), effects.Planeswalker())).WithCount(0, 1),
		),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			x := ctx.X()
			if x <= 0 {
				return nil
			}
			if p, ok := ctx.ClauseTarget(0); ok && p.Kind == game.TargetPlayer {
				if err := (effects.DrawCards{Player: p.ID, N: x}).Apply(ctx); err != nil {
					return err
				}
			}
			if c, ok := ctx.ClauseTarget(1); ok && c.Kind == game.TargetCard {
				return effects.DealDamage{Source: ctx.Source(), Target: c.ID, Amount: x}.Apply(ctx)
			}
			return nil
		},
	})
}
