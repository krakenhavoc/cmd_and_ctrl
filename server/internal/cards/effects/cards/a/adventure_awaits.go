package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Adventure Awaits — Sorcery {1}{G} (Foundations Jumpstart, #5929):
//
//	"Look at the top five cards of your library. You may reveal a creature
//	 card from among them and put it into your hand. Put the rest on the
//	 bottom of your library in a random order. If you didn't put a card into
//	 your hand this way, draw a card."
//
// Blossom Prancer's take shape: the revealed pick, the rest on the bottom in
// a random order, then the consolation draw only when nothing reached the
// hand (a declined "may" or no creature among the five).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "70235727-e330-46fa-b057-ddbb55cd1ac3",
		Name:         "Adventure Awaits",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			player := ctx.Controller()
			return effects.TakeFromLibraryToHand{
				Player:   player,
				Cards:    ctx.Game.LookAtTopOfLibraryForEffect(player, 5),
				Match:    effects.Creature(),
				Max:      1,
				Optional: true,
				Label:    "Adventure Awaits — reveal a creature card and put it into your hand",
				Key:      adventureAwaitsTake,
			}.Apply(ctx)
		},
	})
}

// adventureAwaitsTake is the take: the rest on the bottom in a random order,
// then a card drawn if none went to the hand. An on-disk key.
var adventureAwaitsTake = effects.RegisterTake("choose-cards/adventure-awaits-take",
	effects.TakeDef{Reveal: true, Then: func(g *game.Game, res effects.TakeFromLibraryResult) error {
		if err := effects.TakeRestOnBottomInRandomOrder(g, res); err != nil {
			return err
		}
		if len(res.Taken) > 0 {
			return nil
		}
		return g.DrawNForEffect(res.Player, 1)
	}})
