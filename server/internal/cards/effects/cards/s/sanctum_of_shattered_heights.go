package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sanctum of Shattered Heights — Legendary Enchantment — Shrine {2}{R} (Core
// Set 2021, #5511):
//
//	"{1}, Discard a land card or Shrine card: Sanctum of Shattered Heights
//	 deals X damage to target creature or planeswalker, where X is the number
//	 of Shrines you control."
//
// The discard is a cost: only a land card or a Shrine card in hand can pay
// it. X is counted as the ability resolves (the Sanctum itself counts while
// it is on the battlefield), so a Shrine that left in response is not
// counted; a target that is gone makes the ability do nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5e39d311-efb5-4f51-82cd-e8079e32907a",
		Name:         "Sanctum of Shattered Heights",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label: "{1}, Discard a land card or Shrine card: Sanctum of Shattered Heights deals X damage to target creature or planeswalker, where X is the number of Shrines you control.",
			Cost: effects.Plus(effects.ManaCost("{1}"),
				effects.DiscardCardsMatching(1, "a land card or Shrine card", func(c game.Card) bool { return c.IsLand() || c.HasSubtype("Shrine") })),
			Targets: effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				x := effects.S82OldB106ShrinesYouControl(g, ctx.Controller())
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetCard {
						return effects.DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: x}.Apply(ctx)
					}
				}
				return nil
			},
		}},
	})
}
