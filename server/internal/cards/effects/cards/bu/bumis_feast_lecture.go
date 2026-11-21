package cardsbu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bumi's Feast Lecture — Sorcery — Lesson {1}{G} (Avatar: The Last
// Airbender Eternal, #5075):
//
//	"Create a Food token. Then earthbend X, where X is twice the number
//	 of Foods you control."
//
// The Food is made first and X is counted after it, so the new Food
// counts. "Foods" is any Food artifact you control, token or not, read by
// its effective subtype (the ruling). The land is the spell's only target
// (EarthbendTargets, CR 701.66a): if it has become illegal the spell does
// not resolve and no Food is made, as the ruling says.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "41f53bb5-8a62-4e6b-b9c3-5aa530a036bc",
		Name:         "Bumi's Feast Lecture",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EarthbendTargets(),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			land := effects.FirstLegalBattlefieldTarget(ctx)
			if land == uuid.Nil {
				return nil
			}
			if err := (effects.CreateToken{Template: effects.FoodToken(), N: 1}).Apply(ctx); err != nil {
				return err
			}
			foods := effects.CountControlled(ctx.Game, ctx.Controller(), func(c game.Card) bool { return c.HasSubtype("Food") })
			return effects.Earthbend{Target: land, N: 2 * foods}.Apply(ctx)
		},
	})
}
