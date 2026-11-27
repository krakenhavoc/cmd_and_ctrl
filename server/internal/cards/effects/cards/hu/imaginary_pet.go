package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Imaginary Pet — Creature — Illusion {1}{U}, 4/4 (Urza's Saga, S82 OLD batch
// 731, #5624):
//
//	"At the beginning of your upkeep, if you have a card in hand, return this
//	 creature to its owner's hand."
//
// The "if" is intervening (CR 603.4): a card in its controller's hand is
// checked as the upkeep begins and again as the trigger resolves, so emptying
// the hand in response stops the return. It returns only the object that
// triggered; one that left and came back is a new object (CR 400.7).
//
// No simplification.
func init() {
	hasCard := func(g *game.Game, you game.Card) bool {
		p := g.PlayerByIDForEffect(you.Controller)
		return p != nil && p.Hand != nil && len(p.Hand.Cards) > 0
	}
	effects.Register(effects.Spec{
		OracleID:     "afee3422-5490-411b-a36e-82e447c1bd0f",
		Name:         "Imaginary Pet",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondBoard{
				Name: "you have a card in hand",
				Test: func(ctx *effects.Context) bool {
					p := ctx.Game.PlayerByIDForEffect(ctx.Controller())
					return p != nil && p.Hand != nil && len(p.Hand.Cards) > 0
				},
			}, effects.On(game.EventBeginUpkeep, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return effects.ByYou(ev, source, lki, g) && hasCard(g, *source)
			}, "Imaginary Pet — return this creature to its owner's hand",
				effects.OldB731ReturnThisToOwnersHand)),
		},
	})
}
