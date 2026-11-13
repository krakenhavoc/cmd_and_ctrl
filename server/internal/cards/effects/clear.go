package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Clear — Instant {1}{W} (Urza's Saga, #5624):
//
//	"Destroy target enchantment.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// Airship Crash's shape on a narrower target: the destroy applies only if
// the enchantment is still a legal target as the spell resolves
// (CR 608.2b). Cycling is the shared CR 702.29a ability from the hand.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d0783be9-e518-431a-8e71-121610da50e1",
		Name:         "Clear",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetPermanent("target enchantment", Enchantment()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return destroyFirstLegalCardTarget(ctx.Game, item)
		},
		Activated: []ActivatedAbility{Cycling("{2}")},
	})
}
