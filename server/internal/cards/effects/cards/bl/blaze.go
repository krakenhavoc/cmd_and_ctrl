package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blaze — Sorcery for {X}{R}:
//
//	"Blaze deals X damage to any target."
//
// S20 sub-PR 3: the first X spell. The client's X prompt announces
// X, the S15 cost gate charges {X}{R}, and the effect reads ctx.X()
// at resolution. X = 0 is a legal (if pointless) cast: zero damage.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0596920f-9946-42f4-a03b-24aab67f9f1b",
		Name:         "Blaze",
		XMatters:     true,
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			return effects.DealDamage{
				Source: ctx.Source(),
				Target: item.Targets[0].ID,
				Amount: ctx.X(),
			}.Apply(ctx)
		},
	})
}
