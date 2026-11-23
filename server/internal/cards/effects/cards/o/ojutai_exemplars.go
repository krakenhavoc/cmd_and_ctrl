package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ojutai Exemplars — Creature — Human Monk {2}{W}{W}, 4/4 (Foundations
// Jumpstart):
//
//	"Whenever you cast a noncreature spell, choose one —
//	 • Tap target creature.
//	 • This creature gains first strike and lifelink until end of turn.
//	 • Exile this creature, then return it to the battlefield tapped
//	   under its owner's control."
//
// A modal cast trigger (Gandalf the Grey's shape): the bullet is chosen
// as the trigger goes on the stack, and only the tap bullet targets. The
// third bullet is a flicker of the Exemplars themselves: it comes back
// as a new object, tapped, under its owner's control.
//
// No simplification.
func init() {
	t := effects.WheneverYouCast(effects.Noncreature(), "Ojutai Exemplars — choose one",
		func(_ *game.Game, _ *game.StackItem) error { return nil })
	t.Modes = effects.ChooseOne(
		effects.ModeDoing("Tap target creature.", effects.TargetCreature("target creature"), effects.TapTheModesTarget),
		effects.ModeDoing("This creature gains first strike and lifelink until end of turn.", nil,
			func(item *game.StackItem, ctx *effects.Context, _ int) error {
				return effects.GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{"first strike", "lifelink"}, Label: "Ojutai Exemplars — first strike and lifelink"}.Apply(ctx)
			}),
		effects.ModeDoing("Exile this creature, then return it to the battlefield tapped under its owner's control.", nil,
			func(item *game.StackItem, ctx *effects.Context, _ int) error {
				c, ok := ctx.Game.LookupCardForEffect(item.SourceCardID)
				if !ok {
					return nil
				}
				return effects.Flicker{Target: item.SourceCardID, Controller: c.Owner, Tapped: true}.Apply(ctx)
			}),
	)
	effects.Register(effects.Spec{
		OracleID:     "4405db13-7b36-4dc2-9a5d-1d31d568abf5",
		Name:         "Ojutai Exemplars",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{t},
	})
}
