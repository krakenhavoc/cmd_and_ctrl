package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ray of Revelation — Instant {1}{W} (Judgment, #5611):
//
//	"Destroy target enchantment.
//	 Flashback {G} (You may cast this card from your graveyard for its
//	 flashback cost. Then exile it.)"
//
// Clear's destroy (destroyFirstLegalCardTarget, nothing when the target
// has gone, CR 608.2b) with a flashback at the off-colour {G}, the
// shared constructor, exiled as it leaves the stack (CR 702.34a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "4620bfa8-7198-44ae-a162-51a19ca0f555",
		Name:             "Ray of Revelation",
		Completeness:     CompletenessFull,
		Purpose:          game.Purpose{Answers: game.AnswerRemove},
		Targets:          TargetPermanent("target enchantment", Enchantment()),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{G}")},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return destroyFirstLegalCardTarget(ctx.Game, item)
		},
	})
}
