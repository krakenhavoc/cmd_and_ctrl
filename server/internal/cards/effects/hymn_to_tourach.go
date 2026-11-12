package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hymn to Tourach — Sorcery {B}{B} (Fallen Empires, #5637):
//
//	"Target player discards two cards at random."
//
// The random discard (CR 701.9b) through DiscardCards, the primitive
// for cards that print "at random". Any player may be targeted, the
// caster included. A hand of one card loses that card, and an empty
// hand loses nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "992d9a72-d34d-47d7-96fb-37f0efb515ef",
		Name:         "Hymn to Tourach",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			player := TargetedPlayer(ctx)
			if player == uuid.Nil {
				return nil
			}
			return DiscardCards{Player: player, N: 2}.Apply(ctx)
		},
	})
}
