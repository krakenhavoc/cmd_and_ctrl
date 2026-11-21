package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "d0783be9-e518-431a-8e71-121610da50e1",
		Name:         "Clear",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target enchantment", effects.Enchantment()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.DestroyFirstLegalCardTarget(ctx.Game, item)
		},
		Activated: []effects.ActivatedAbility{effects.Cycling("{2}")},
	})
}
