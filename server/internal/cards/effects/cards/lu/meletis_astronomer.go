package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Meletis Astronomer — {1}{U} Creature — Human Wizard 1/3 (Born of the Gods):
//
//	"Heroic — Whenever you cast a spell that targets this creature, look at the top three
//	 cards of your library. You may reveal an enchantment card from among them and put it
//	 into your hand. Put the rest on the bottom of your library in any order."
//
// Bond of Flourishing's dig, over enchantment cards.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6593a7c1-fb48-4e53-8e6a-339ad2cc4489",
		Name:         "Meletis Astronomer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Heroic("Meletis Astronomer — look at the top three cards of your library; you may reveal an enchantment card and put it into your hand; the rest on the bottom in any order (heroic)",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					player := ctx.Controller()
					return effects.TakeFromLibraryToHand{
						Player:   player,
						Cards:    g.LookAtTopOfLibraryForEffect(player, 3),
						Match:    effects.Enchantment(),
						Max:      1,
						Optional: true,
						Label:    "Meletis Astronomer — reveal an enchantment card and put it into your hand",
						Key:      effects.TakeRevealRestBottomAny,
					}.Apply(ctx)
				}),
		},
	})
}
