package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Immortal Phoenix — Creature — Phoenix {4}{R}{R}, 5/3 (M19 Gift Pack,
// #5524):
//
//	"Flying
//	 When this creature dies, return it to its owner's hand."
//
// Endless Cockroaches' death trigger (WhenThisDies with
// returnSelfFromGraveyardToHand): only a death sets it off, it moves the
// card only while it is still the graveyard object the trigger stamped
// (CR 400.7), and it goes to its OWNER's hand whoever controlled it when
// it died.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1d8bd391-b133-4c8b-95ad-7f7ab99137c7",
		Name:            "Immortal Phoenix",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Immortal Phoenix — return it to its owner's hand", returnSelfFromGraveyardToHand),
		},
	})
}
