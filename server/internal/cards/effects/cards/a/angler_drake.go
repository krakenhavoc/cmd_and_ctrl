package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Angler Drake — {4}{U}{U} Creature — Drake 4/4 (Amonkhet, #5535):
//
//	"Flying
//	 When this creature enters, you may return target creature to its
//	 owner's hand."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "620d7146-b5b3-487c-a24b-f36b10f48df4",
		Name:            "Angler Drake",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(
				effects.WhenThisEnters("Angler Drake — return target creature to its owner's hand", effects.BounceTheTargetToHand),
				effects.TargetCreature("target creature")),
				"Angler Drake — return target creature to its owner's hand?"),
		},
	})
}
