package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mist Raven — Creature — Bird {2}{U}{U}, 2/2 (Innistrad Remastered, S82
// GAP, #5948):
//
//	"Flying
//	 When this creature enters, return target creature to its owner's
//	 hand."
//
// Man-o'-War's enters trigger on a flier: any creature, the Raven itself
// included, chosen as the trigger goes on the stack; a target that has left
// does nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "12654457-3aae-4046-a976-8e41c3f78927",
		Name:            "Mist Raven",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Mist Raven — return target creature to its owner's hand",
				bounceTheTargetToHand), TargetCreature("target creature")),
		},
	})
}
