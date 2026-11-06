package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "d7aaffed-a143-49d6-a516-b9bc8b503aee",
		Name:         "Peerless Ropemaster",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Peerless Ropemaster — return up to one target tapped creature to its owner's hand",
				bounceTheTargetToHand), UpToOneTargetCreature("up to one target tapped creature", b751Tapped())),
		},
	})
}
