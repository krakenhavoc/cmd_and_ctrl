package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ray of Distortion — Instant {3}{W} (Odyssey, #5613):
//
//	"Destroy target artifact or enchantment.
//	 Flashback {4}{W}{W} (You may cast this card from your graveyard for its
//	 flashback cost. Then exile it.)"
//
// Disenchant's sentence (CR 701.8a) on the shared destroyTheTargetPermanent
// body; a target gone by resolution does nothing (CR 608.2b).
//
// Flashback is the shared constructor: the graveyard cast, the price,
// and CR 702.34a's exile as the spell leaves the stack (resolved,
// countered, or with its target gone).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "1c838441-4713-49dd-97e6-4fb7155c36c8",
		Name:             "Ray of Distortion",
		Completeness:     effects.CompletenessFull,
		Targets:          effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
		Purpose:          game.Purpose{Answers: game.AnswerRemove},
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Flashback("{4}{W}{W}")},
		OnResolve:        effects.DestroyTheTargetPermanent,
	})
}
