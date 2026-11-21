package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Acolyte Hybrid — Creature — Tyranid Human {2}{R}, 2/2 (Warhammer 40,000
// Commander, S82 GAP batch 193, #5948):
//
//	"Heavy Rock Cutter — Whenever this creature attacks, destroy up to one
//	 target artifact. If an artifact is destroyed this way, its controller draws
//	 a card."
//
// The artifact is destroyed through the keyed destruction (Boseiju's door), so
// the draw reads what was really destroyed: an indestructible artifact, or one
// a replacement kept from the graveyard, draws nothing. The draw is its
// controller's as it was before it died.
//
// No simplification.
func init() {
	attack := effects.WheneverThisAttacks("Acolyte Hybrid — destroy up to one target artifact; if destroyed, its controller draws a card",
		effects.GapB193DestroyTargetArtifactControllerDraws)
	attack.Targets = effects.TargetPermanent("up to one target artifact", effects.Artifact()).WithCount(0, 1)
	effects.Register(effects.Spec{
		OracleID:     "998daf11-c0ed-4801-b013-d8b3136f16d1",
		Name:         "Acolyte Hybrid",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{attack},
	})
}
