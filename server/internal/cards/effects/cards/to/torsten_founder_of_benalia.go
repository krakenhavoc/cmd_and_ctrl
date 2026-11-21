package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Torsten, Founder of Benalia — Legendary Creature — Human Soldier
// {5}{G}{W}, 7/7 (Dominaria United Commander, S82):
//
//	"When Torsten enters, reveal the top seven cards of your library. Put
//	 any number of creature and/or land cards from among them into your hand
//	 and the rest on the bottom of your library in a random order.
//	 When Torsten dies, create seven 1/1 white Soldier creature tokens."
//
// The enter trigger is Marina Vendrell's reveal over seven cards, with a
// choice in place of "all": you name any number (none included) of the
// creature and land cards, and the rest of the seven, the cards you did
// not take as well as the cards that could not be taken, go to the bottom
// in a random order.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3b1c6c1e-e298-4f45-8ab7-e0bb5358e11f",
		Name:         "Torsten, Founder of Benalia",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Torsten, Founder of Benalia — reveal the top seven cards; put any number of creature and/or land cards into your hand, the rest on the bottom in a random order",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.TakeFromLibraryToHand{
						Player:   item.Controller,
						Cards:    g.RevealTopOfLibraryForEffect(item.Controller, ctx.Source(), 7, "Torsten, Founder of Benalia — reveal the top seven cards"),
						Match:    effects.Or(effects.Creature(), effects.Land()),
						Optional: true,
						Label:    "Torsten, Founder of Benalia — any number of creature and/or land cards into your hand",
						Key:      effects.TakeRestBottomRandom,
					}.Apply(ctx)
				}),
			effects.WhenThisDies("Torsten, Founder of Benalia — create seven 1/1 white Soldier tokens",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Soldier"), N: 7})),
		},
	})
}
