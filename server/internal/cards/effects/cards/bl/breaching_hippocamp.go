package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "5ba04e0d-a3eb-448d-8a97-08c155ef9694",
		Name:            "Breaching Hippocamp",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Breaching Hippocamp — untap another target creature you control",
					effects.UntapFirstLegalTarget),
				effects.Another(effects.TargetCreature("another target creature you control", effects.YouControl()))),
		},
	})
}
