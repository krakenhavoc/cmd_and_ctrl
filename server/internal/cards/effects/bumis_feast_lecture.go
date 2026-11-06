package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:     "41f53bb5-8a62-4e6b-b9c3-5aa530a036bc",
		Name:         "Bumi's Feast Lecture",
		Completeness: CompletenessFull,
		Targets:      EarthbendTargets(),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			land := FirstLegalBattlefieldTarget(ctx)
			if land == uuid.Nil {
				return nil
			}
			if err := (CreateToken{Template: FoodToken(), N: 1}).Apply(ctx); err != nil {
				return err
			}
			foods := countControlled(ctx.Game, ctx.Controller(), func(c game.Card) bool { return c.HasSubtype("Food") })
			return Earthbend{Target: land, N: 2 * foods}.Apply(ctx)
		},
	})
}
