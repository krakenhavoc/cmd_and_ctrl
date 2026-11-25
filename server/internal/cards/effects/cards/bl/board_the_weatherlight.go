package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Board the Weatherlight — Sorcery {1}{W} (Dominaria, S82 OLD, #5530):
//
//	"Look at the top five cards of your library. You may reveal a historic card
//	 from among them and put it into your hand. Put the rest on the bottom of
//	 your library in a random order. (Artifacts, legendaries, and Sagas are
//	 historic.)"
//
// Blossom Prancer's look-reveal-and-take over five cards, with "historic" read
// through the layered view (CR 700.6): an artifact, a legendary or a Saga. The
// pick is optional; the card taken is revealed, and the rest go to the bottom
// in a random order.
//
// No simplification.
func init() {
	const label = "Board the Weatherlight — you may reveal a historic card and put it into your hand"
	effects.Register(effects.Spec{
		OracleID:     "c4ce96d8-3ce4-4ed2-9ad3-73119aa79fbe",
		Name:         "Board the Weatherlight",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			player := ctx.Controller()
			return effects.TakeFromLibraryToHand{
				Player:   player,
				Cards:    ctx.Game.LookAtTopOfLibraryForEffect(player, 5),
				Match:    effects.B14Historic(),
				Max:      1,
				Optional: true,
				Label:    label,
				Key:      boardTheWeatherlightTake,
			}.Apply(ctx)
		},
	})
}

// boardTheWeatherlightTake is the take: the card is revealed and the rest go
// on the bottom in a random order. An on-disk key.
var boardTheWeatherlightTake = effects.RegisterTake("choose-cards/s82-old-b212-board-the-weatherlight-take",
	effects.TakeDef{Reveal: true, Then: effects.TakeRestOnBottomInRandomOrder})
