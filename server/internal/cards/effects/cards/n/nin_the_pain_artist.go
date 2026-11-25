package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nin, the Pain Artist — Legendary Creature — Vedalken Wizard {U}{R}, 1/1
// (Commander 2017, S82 older sets):
//
//	"{X}{U}{R}, {T}: Nin deals X damage to target creature. That creature's
//	 controller draws X cards."
//
// The creature's controller is read before the damage (CR 608.2h: it is the
// controller as the ability resolves, and a creature the damage destroys is
// still read as it last was), and draws X cards whether or not it survived.
// If the target is gone the ability does nothing at all (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ec439630-e18b-45f6-8034-2ecbea27772f",
		Name:         "Nin, the Pain Artist",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Activated: []effects.ActivatedAbility{{
			Label:   "{X}{U}{R}, {T}: Nin deals X damage to target creature. That creature's controller draws X cards.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.Plus(effects.ManaCost("{X}{U}{R}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				victim, ok := g.LookupCardForEffect(id)
				if !ok {
					return nil
				}
				x := ctx.X()
				if err := (effects.DealDamage{Source: ctx.Source(), Target: id, Amount: x}).Apply(ctx); err != nil {
					return err
				}
				return effects.DrawCards{Player: victim.Controller, N: x}.Apply(ctx)
			},
		}},
	})
}
