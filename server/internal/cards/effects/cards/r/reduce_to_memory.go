package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Reduce to Memory — Sorcery — Lesson {1}{W}{W} (Strixhaven: School of Mages,
// #5502):
//
//	"Exile target nonland permanent. Its controller creates a 3/2 red and white
//	 Spirit creature token."
//
// Swords to Plowshares' shape with a token for the permanent's controller. That
// controller is read before the exile (CR 608.2h), and the Spirit is made for
// them even if the permanent could not be exiled. Being a Lesson is its type
// line, which learn reads.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "661736eb-97d7-4b3c-ab18-708479a5fb29",
		Name:         "Reduce to Memory",
		Completeness: effects.CompletenessFull,
		// #2679: what the target's controller is given back.
		Purpose: effects.ForTargets(effects.RemovalReturning(0, game.TargetReturn{CreatureTokens: 1, TokenPower: 3, TokenToughness: 2})),
		Targets: effects.TargetPermanent("target nonland permanent", effects.Nonland()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			victim, ok := effects.ControllerOfTarget(ctx, target)
			if !ok {
				return nil
			}
			if err := (effects.ExileTarget{Target: target}).Apply(ctx); err != nil {
				return err
			}
			return effects.CreateToken{Controller: victim, Template: effects.TokenCard("3/2 red and white Spirit"), N: 1}.Apply(ctx)
		},
	})
}
