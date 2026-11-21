package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wizard Mentor — Creature — Human Wizard {2}{U}, 2/2 (Jumpstart 2022, #5948):
//
//	"{T}: Return this creature and target creature you control to their
//	 owner's hand."
//
// Both go back as the ability resolves. The target may be the Mentor
// itself (it is a creature you control), which returns once. A Mentor
// that left and came back while the ability waited is a new object and
// stays (CR 400.7); a target that is gone returns nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a1fb1b77-a49f-4447-bbe0-156407f192e2",
		Name:         "Wizard Mentor",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Return this creature and target creature you control to their owner's hand.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target creature you control", effects.YouControl()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if err := (effects.BounceToHand{Target: item.SourceCardID}).Apply(ctx); err != nil {
					return err
				}
				if !ok || id == item.SourceCardID {
					return nil
				}
				return effects.BounceToHand{Target: id}.Apply(ctx)
			},
		}},
	})
}
