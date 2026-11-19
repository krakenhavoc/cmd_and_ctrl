package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Collected Company — Instant {3}{G} (Special Guests):
//
//	"Look at the top six cards of your library. Put up to two creature
//	 cards with mana value 3 or less from among them onto the battlefield.
//	 Put the rest on the bottom of your library in any order."
//
// A look, not a reveal: only its controller sees the six. The rest are
// ordered by the player (bnsB06PutRestOnBottomInAnyOrder), unlike the
// random-order tail of the cards that print "a random order".
func init() {
	Register(Spec{
		OracleID:     "0bbcb563-09b3-499d-a9cc-3a6d7501dd81",
		Name:         "Collected Company",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			cards := ctx.Game.LookAtTopOfLibraryForEffect(item.Controller, 6)
			if len(cards) == 0 {
				return nil
			}
			return PutFromLibraryOntoBattlefield{
				Player:   item.Controller,
				Cards:    cards,
				Match:    And(Creature(), ManaValueLE(3)),
				Max:      2,
				Optional: true,
				Label:    "Collected Company — put up to two creature cards with mana value 3 or less onto the battlefield",
				ThenTail: bnsB06PutRestOnBottomInAnyOrder,
			}.Apply(ctx)
		},
	})
}
