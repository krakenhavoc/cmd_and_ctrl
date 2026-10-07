package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dig Through Time — Instant {6}{U}{U}:
//
//	"Delve. Look at the top seven cards of your library. Put two of them into your hand and the rest on the bottom of your library in any order."
//
// Delve (CR 702.66, ADR 0100) is `Delve: true` and nothing else: the
// engine prices it, validates the graveyard cards named on
// `delve_ids` and exiles them at CR 601.2h with the spell on the
// stack.
//
// Impulse's shape with seven and two: a private look, a mandatory
// take of two (fewer if the library holds fewer), and the ordered
// bottom (ADR 0088). No simplification.
//
// Its purpose (ADR 0126 §6) is Tutors: 2. Taking two of seven is not
// a draw, and the bot should price it as two chosen cards in hand,
// which is what a tutor's purpose says.
func init() {
	Register(Spec{
		OracleID:     "f8b17b89-26ce-4208-874a-9e1d66514640",
		Name:         "Dig Through Time",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tutors: 2},
		Delve:        true,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			player := ctx.Controller()
			return TakeFromLibraryToHand{
				Player: player,
				Cards:  ctx.Game.LookAtTopOfLibraryForEffect(player, 7),
				Max:    2,
				Label:  "Dig Through Time — put two of them into your hand",
				Then:   TakeRestOnBottomInAnyOrder,
			}.Apply(ctx)
		},
	})
}
