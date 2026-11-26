package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gallows at Willow Hill — Artifact {3}:
//
//	"{3}, {T}, Tap three untapped Humans you control: Destroy target
//	 creature. Its controller creates a 1/1 white Spirit creature token with
//	 flying."
//
// The creature's controller is read before it is destroyed, and gets the
// token even when the creature was indestructible or regenerated: the
// instruction does not depend on the destruction.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0d80841e-7a33-4b7f-b061-fa2f87f596de",
		Name:         "Gallows at Willow Hill",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, {T}, Tap three untapped Humans you control: Destroy target creature. Its controller creates a 1/1 white Spirit creature token with flying.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost: game.AbilityCost{Mana: "{3}", Tap: true, TapOthers: &game.TapOthersCost{
				Count:  3,
				Filter: effects.TargetPermanent("three untapped Humans you control", effects.HasSubtype("Human")),
				Label:  "three untapped Humans you control",
			}},
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					controller, ok := effects.ControllerOfTarget(ctx, t.ID)
					if !ok {
						return nil
					}
					if err := (effects.DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
						return err
					}
					return effects.CreateToken{Controller: controller, Template: effects.TokenCard("1/1 white Spirit with flying"), N: 1}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
