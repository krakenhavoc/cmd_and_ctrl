package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blossom Prancer — Creature — Spirit {3}{G}{G}, 4/4 (Kamigawa: Neon
// Dynasty, S82):
//
//	"Reach
//	 When this creature enters, look at the top five cards of your library.
//	 You may reveal a creature or enchantment card from among them and put
//	 it into your hand. Put the rest on the bottom of your library in a
//	 random order. If you didn't put a card into your hand this way, you
//	 gain 4 life."
//
// Contagious Vorrac's look-reveal-and-take over five cards and creature or
// enchantment cards, with the life gain as the "if you didn't" rider.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "69e92c4f-df05-4297-a453-8e4644c31a90",
		Name:            "Blossom Prancer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Blossom Prancer — look at the top five cards of your library for a creature or enchantment card",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					player := ctx.Controller()
					return effects.TakeFromLibraryToHand{
						Player:   player,
						Cards:    g.LookAtTopOfLibraryForEffect(player, 5),
						Match:    effects.Or(effects.Creature(), effects.Enchantment()),
						Max:      1,
						Optional: true,
						Label:    "Blossom Prancer — you may reveal a creature or enchantment card and put it into your hand",
						Key:      blossomPrancerTake,
					}.Apply(ctx)
				}),
		},
	})
}

// blossomPrancerTake is the take: the rest on the bottom in a random order,
// then 4 life if no card went to the hand. An on-disk key.
var blossomPrancerTake = effects.RegisterTake("choose-cards/blossom-prancer-take",
	effects.TakeDef{Reveal: true, Then: func(g *game.Game, res effects.TakeFromLibraryResult) error {
		if err := effects.TakeRestOnBottomInRandomOrder(g, res); err != nil {
			return err
		}
		if len(res.Taken) > 0 {
			return nil
		}
		return effects.GainLife{Player: res.Player, Amount: 4}.Apply(effects.KeyedContext(g, res.Player, res.Source))
	}})
