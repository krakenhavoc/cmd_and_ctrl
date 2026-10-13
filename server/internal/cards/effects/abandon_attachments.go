package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Abandon Attachments — Instant — Lesson {1}{U/R} (EDHREC rank 3697):
//
//	"You may discard a card. If you do, draw two cards."
//
// A two-mana loot-into-two, and a Lesson for the learn family (the
// type line is what "Lesson spell" watchers read, so nothing on the
// card is needed for that).
//
// "You may discard a card" is the discard prompt with a floor of zero
// (UpTo), asked as the spell resolves; the draw is its continuation and
// runs only for a player who did discard ("if you do"). So a countered
// spell keeps the card, and a player with an empty hand can cast it to
// do nothing.
func init() {
	Register(Spec{
		OracleID:     "82333385-631f-4abf-b159-bb367f1c6fd9",
		Name:         "Abandon Attachments",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			player := item.Controller
			return ctx.Game.PlayerDiscardsThenForEffect(game.DiscardPrompt{
				Player:   player,
				Source:   item.SourceCardID,
				N:        1,
				UpTo:     true,
				Question: "Abandon Attachments — you may discard a card; if you do, draw two",
			}, func(g *game.Game, done game.PromptedDiscards) error {
				if !done.Discarded(player) {
					return nil
				}
				return g.DrawNForEffect(player, 2)
			})
		},
	})
}
