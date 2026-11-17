package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Frantic Purification — Instant {2}{W} (Torment, #5612):
//
//	"Destroy target enchantment.
//	 Madness {W} (If you discard this card, discard it into exile. When
//	 you do, cast it for its madness cost or put it into your graveyard.)"
//
// Clear's destroy (destroyFirstLegalCardTarget, nothing when the target
// has gone, CR 608.2b) with madness, which the engine grows from the one
// string (CR 702.35a, #657): discarded, it is exiled and offered for {W}
// at instant speed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b4025136-3146-499a-8c17-ed9de548ab8d",
		Name:         "Frantic Purification",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetPermanent("target enchantment", Enchantment()),
		Madness:      "{W}",
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return destroyFirstLegalCardTarget(ctx.Game, item)
		},
	})
}
