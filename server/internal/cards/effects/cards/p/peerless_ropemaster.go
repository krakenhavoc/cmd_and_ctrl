package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Peerless Ropemaster — Creature — Human Rogue {4}{U}, 4/4 (Outlaws of
// Thunder Junction, #4468):
//
//	"When this creature enters, return up to one target tapped creature
//	 to its owner's hand."
//
// Pelia's enters bounce with "up to one" and a tapped-creature target.
// Any player's creature may be chosen, the Ropemaster's own included.
// The target must still be a tapped creature as the trigger resolves
// (CR 608.2b), so one that untaps in response stays.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d7aaffed-a143-49d6-a516-b9bc8b503aee",
		Name:         "Peerless Ropemaster",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Peerless Ropemaster — return up to one target tapped creature to its owner's hand",
				effects.BounceTheTargetToHand), effects.UpToOneTargetCreature("up to one target tapped creature", effects.B751Tapped())),
		},
	})
}
