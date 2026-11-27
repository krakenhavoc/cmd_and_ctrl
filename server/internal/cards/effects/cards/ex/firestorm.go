package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Firestorm — {R} Instant (Weatherlight, S82 OLD batch 771, #5629):
//
//	"As an additional cost to cast this spell, discard X cards.
//	 Firestorm deals X damage to each of X targets."
//
// Channeled Force's cost (DiscardXCost: the X announced with the cast is the
// number of cards discarded while the spell is on the stack, CR 601.2f) and
// Immoral Bargain's X targets, here any target (creatures, players and
// planeswalkers, each picked once). The damage is dealt as one event to each
// target that is still legal as the spell resolves (CR 608.2b); at X = 0 the
// spell is castable with no targets and does nothing.
//
// No simplification.
func init() {
	targets := effects.TargetAny()
	targets.CountFromX = true
	effects.Register(effects.Spec{
		OracleID:       "157813c8-af83-41b5-9ffb-4abd167c1bfc",
		Name:           "Firestorm",
		Completeness:   effects.CompletenessFull,
		XMatters:       true,
		AdditionalCost: effects.DiscardXCost(),
		Targets:        targets,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			x := ctx.X()
			if x <= 0 {
				return nil
			}
			return ctx.Game.DamageInstanceForEffect(func() error {
				for _, t := range ctx.LegalTargets() {
					if err := (effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: x}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			})
		},
	})
}
