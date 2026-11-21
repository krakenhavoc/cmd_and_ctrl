package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Floriferous Vinewall — Creature — Plant Wall {1}{G}, 0/2 (Dominaria United,
// S82 DMU batch 4):
//
//	"Defender
//	 When this creature enters, look at the top six cards of your library.
//	 You may reveal a land card from among them and put it into your hand.
//	 Put the rest on the bottom of your library in a random order."
//
// Contagious Vorrac's look-reveal-and-take over six cards, without the
// proliferate rider. Defender is a printed keyword.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "26685121-ab1f-4366-840b-c5a80508cde6",
		Name:            "Floriferous Vinewall",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Floriferous Vinewall — look at the top six cards of your library for a land card",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					player := ctx.Controller()
					return effects.TakeFromLibraryToHand{
						Player:   player,
						Cards:    g.LookAtTopOfLibraryForEffect(player, 6),
						Match:    effects.Land(),
						Max:      1,
						Optional: true,
						Label:    "Floriferous Vinewall — you may reveal a land card and put it into your hand",
						Key:      effects.DmuB04VinewallTake,
					}.Apply(ctx)
				}),
		},
	})
}
