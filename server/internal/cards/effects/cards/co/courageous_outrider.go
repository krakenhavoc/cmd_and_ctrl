package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Courageous Outrider — Creature — Human Scout {3}{W}, 3/4 (Eldritch Moon,
// S82 OLD, #5541):
//
//	"When this creature enters, look at the top four cards of your library.
//	 You may reveal a Human card from among them and put it into your hand.
//	 Put the rest on the bottom of your library in any order."
//
// Beastrider Vanguard's dig with "Human card" and four cards: a look, so
// only the controller sees them; the card taken is revealed; the rest go to
// the bottom in the order the controller picks.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "00915480-ffa8-4930-a158-b1c1f75b7876",
		Name:         "Courageous Outrider",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Courageous Outrider — look at the top four cards of your library; you may reveal a Human card from among them and put it into your hand; put the rest on the bottom in any order",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					player := ctx.Controller()
					return effects.TakeFromLibraryToHand{
						Player:   player,
						Cards:    g.LookAtTopOfLibraryForEffect(player, 4),
						Match:    effects.HasSubtype("Human"),
						Max:      1,
						Optional: true,
						Label:    "Courageous Outrider — you may reveal a Human card from among them and put it into your hand",
						Key:      effects.TakeRevealRestBottomAny,
					}.Apply(ctx)
				}),
		},
	})
}
