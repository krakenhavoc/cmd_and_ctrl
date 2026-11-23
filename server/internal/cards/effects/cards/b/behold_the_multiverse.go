package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Behold the Multiverse — {3}{U} Instant: "Scry 2, then draw two
// cards. Foretell {1}{U}"
//
// Preordain's shape with a bigger draw and foretell on top (#658,
// CR 702.143): {2} on a turn with spare mana, then {1}{U} at the end
// of somebody else's, which is what makes a four-mana instant a
// staple.
//
// The draw goes in Scry's `Then` rather than on the line after it,
// for the reason Preordain's comment gives: Apply only QUEUES the
// prompt, so a draw written afterwards happens before the player has
// put anything back — a different card, off a prompt that is then
// unanswerable because the drawn card is no longer in the library.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d7d2f701-77df-4169-bf98-0d51d6886e9b",
		Name:         "Behold the Multiverse",
		Purpose:      game.Purpose{Scry: 2},
		Completeness: effects.CompletenessFull,
		SpecialActions: []game.SpecialAction{
			effects.Foretell("{1}{U}"),
		},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.Scry{
				Player: ctx.Controller(),
				N:      2,
				// #4253 slice 4: "then draw two cards" is a shared tail.
				ThenTail: ctx.TailN(effects.ThenDrawCards, 2),
			}.Apply(ctx)
		},
	})
}
