package cardsg

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Giant Opportunity — Sorcery {2}{G} (Foundations Jumpstart):
//
//	"You may sacrifice two Foods. If you do, create a 7/7 green Giant
//	 creature token. Otherwise, create three Food tokens."
//
// The choice is made as the spell resolves (CR 608.2c): with two or more
// Foods the controller is asked for exactly two or none (a set rule over
// the pick), and the Foods are sacrificed before the rest runs. Two
// sacrificed makes the Giant; none makes three Foods. With fewer than two
// Foods there is nothing to sacrifice, so the spell makes the Foods
// without asking.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "395aa62e-0786-4bd8-a88f-998e4700742e",
		Name:         "Giant Opportunity",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			foods := effects.PermanentsControlledByMatching(ctx.Game, ctx.Controller(), effects.HasSubtype("Food"))
			if len(foods) < 2 {
				return effects.CreateToken{Template: effects.FoodToken(), N: 3}.Apply(ctx)
			}
			return effects.ChoosePermanents{
				Question: "Giant Opportunity — you may sacrifice two Foods",
				Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
					return effects.PermanentsControlledByMatching(g, of, effects.HasSubtype("Food")), 0, 2
				},
				Keyed: giantOpportunityPick,
			}.Apply(effects.NewContext(ctx.Game, item))
		},
	})
}

// giantOpportunityPick is the sacrifice and what follows it: two Foods
// gone makes the Giant, otherwise three Foods. An on-disk key.
var giantOpportunityPick = effects.RegisterPermanentsPick("giant-opportunity/sacrifice-two-foods", effects.PermanentsDef{
	Validate:  func(picked []game.Card) bool { return len(picked) == 0 || len(picked) == 2 },
	Sacrifice: true,
	Then: func(ctx *effects.Context, picked game.PromptedPicks) error {
		if picked.Count() == 2 {
			return effects.CreateToken{Template: effects.TokenCard("7/7 green Giant"), N: 1}.Apply(ctx)
		}
		return effects.CreateToken{Template: effects.FoodToken(), N: 3}.Apply(ctx)
	},
})
