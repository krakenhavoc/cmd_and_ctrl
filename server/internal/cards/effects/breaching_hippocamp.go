package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Breaching Hippocamp — Creature — Horse Fish {3}{U}, 3/2 (Foundations
// Jumpstart):
//
//	"Flash
//	 When this creature enters, untap another target creature you
//	 control."
//
// Same untap body as Little Bear, without the Bear rider.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5ba04e0d-a3eb-448d-8a97-08c155ef9694",
		Name:            "Breaching Hippocamp",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Breaching Hippocamp — untap another target creature you control",
					untapFirstLegalTarget),
				Another(TargetCreature("another target creature you control", YouControl()))),
		},
	})
}
