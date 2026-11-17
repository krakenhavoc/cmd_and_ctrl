package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shivan Phoenix — Creature — Phoenix {4}{R}{R}, 3/4 (Urza's Legacy,
// #5623):
//
//	"Flying
//	 When this creature dies, return it to its owner's hand."
//
// Immortal Phoenix's death trigger through the shared ulgB1DiesReturnToHand
// row: only a death (CR 700.4) sets it off, and it goes to its OWNER's hand
// while it is still the graveyard object that died (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "49ca8516-624b-4aa8-8e13-d145590cdfda",
		Name:            "Shivan Phoenix",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			ulgB1DiesReturnToHand("Shivan Phoenix"),
		},
	})
}
