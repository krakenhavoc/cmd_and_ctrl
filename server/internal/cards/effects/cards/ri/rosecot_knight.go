package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rosecot Knight — Creature — Human Knight {4}{W}, 3/4 (Modern Horizons
// 3):
//
//	"Vigilance
//	 When this creature enters, look at the top six cards of your library.
//	 You may reveal an artifact or enchantment card from among them and put
//	 it into your hand. Put the rest on the bottom of your library in a
//	 random order. If you didn't put a card into your hand this way, put a
//	 +1/+1 counter on this creature."
//
// Contagious Vorrac's look-reveal-and-take over six cards and artifact or
// enchantment cards. The "if you didn't" rider is the take's continuation,
// run after the rest is on the bottom: with no card taken (a declined
// reveal, or none among the six) the Knight gets the counter, if it is
// still on the battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b81a9cf5-4059-45fb-a4e1-f2fe6aa3c841",
		Name:            "Rosecot Knight",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Purpose:         game.Purpose{Tutors: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Rosecot Knight — look at the top six cards of your library for an artifact or enchantment card",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					player := ctx.Controller()
					return effects.TakeFromLibraryToHand{
						Player:   player,
						Cards:    g.LookAtTopOfLibraryForEffect(player, 6),
						Match:    effects.Or(effects.Artifact(), effects.Enchantment()),
						Max:      1,
						Optional: true,
						Label:    "Rosecot Knight — you may reveal an artifact or enchantment card and put it into your hand",
						Key:      rosecotKnightTake,
					}.Apply(ctx)
				}),
		},
	})
}

// rosecotKnightTake is the take: the rest on the bottom in a random order,
// then a +1/+1 counter on the Knight if no card went to the hand. An
// on-disk key.
var rosecotKnightTake = effects.RegisterTake("choose-cards/rosecot-knight-take",
	effects.TakeDef{Reveal: true, Then: func(g *game.Game, res effects.TakeFromLibraryResult) error {
		if err := effects.TakeRestOnBottomInRandomOrder(g, res); err != nil {
			return err
		}
		if len(res.Taken) > 0 {
			return nil
		}
		return g.AddCounterForEffect(res.Source, game.CounterPlusOne, 1)
	}})
