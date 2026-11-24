package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Specimen Collector — Creature — Vedalken Wizard {4}{U}, 2/1 (Modern Horizons 2):
//
//	"When this creature enters, create a 1/1 green Squirrel creature token and a
//	 0/3 blue Crab creature token.
//	 When this creature dies, create a token that's a copy of target token you
//	 control."
//
// The death trigger targets a token you control as it goes on the stack and
// copies it as it resolves (CR 608.2b); the copy is made from the token as it
// last existed if it has since left.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fdb51cc1-ce0e-4920-b123-1cdd570fc38b",
		Name:         "Specimen Collector",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Specimen Collector — create a 1/1 green Squirrel creature token and a 0/3 blue Crab creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 green Squirrel"), N: 1}, effects.CreateToken{Template: effects.TokenCard("0/3 blue Crab"), N: 1})),
			effects.Targeting(
				effects.WhenThisDies("Specimen Collector — create a token that's a copy of target token you control",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						for _, id := range effects.LegalTargetIDs(ctx) {
							return effects.CreateTokenCopy{Controller: item.Controller, Copy: id, N: 1}.Apply(ctx)
						}
						return nil
					}),
				effects.TargetPermanent("target token you control", effects.IsTokenPredicate(), effects.YouControl()),
			),
		},
	})
}
