package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dwarven Confluencer — Creature — Dwarf {2}{R}, 1/1:
//
//	"{T}: Destroy target nontoken land. Its controller creates a Mana
//	 Confluence token. (It's a land with "{T}, Pay 1 life: Add one mana of any
//	 color.")"
//
// Beast Within's order: the land's controller is read BEFORE the destroy, then
// the land is destroyed, then that player (not the Confluencer's controller)
// creates the token. A target that is no longer a legal nontoken land as the
// ability resolves does nothing (CR 608.2b), so neither half happens. An
// indestructible land is not destroyed, but its controller still gets the
// token: the instruction to create it does not depend on the destruction.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d483f61b-6000-43ca-b5eb-8ab1c924ffd6",
		Name:         "Dwarven Confluencer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Destroy target nontoken land. Its controller creates a Mana Confluence token.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.TapCost(),
			Targets: effects.TargetPermanent("target nontoken land", effects.Land(), effects.Not(effects.IsTokenPredicate())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					controller, ok := effects.ControllerOfTarget(ctx, t.ID)
					if !ok {
						return nil
					}
					if err := (effects.DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
						return err
					}
					return effects.CreateToken{Controller: controller, Template: effects.ManaConfluenceToken(), N: 1}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
