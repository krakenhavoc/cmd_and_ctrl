package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Viridian Corrupter — Creature — Phyrexian Elf Shaman {1}{G}{G}, 2/2:
//
//	"Infect
//	 When this creature enters, destroy target artifact."
//
// The target is chosen as the trigger is put on the stack; if it is gone or
// no longer an artifact by resolution nothing is destroyed (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8ba5e97b-3eac-4291-8674-4b402a5554bc",
		Name:            "Viridian Corrupter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"infect"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Viridian Corrupter — destroy target artifact", destroyChosenPermanent),
				TargetPermanent("target artifact", Artifact())),
		},
	})
}
