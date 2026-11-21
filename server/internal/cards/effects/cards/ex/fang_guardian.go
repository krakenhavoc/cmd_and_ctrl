package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fang Guardian — Creature — Ape Druid {3}{G}, 4/2 (Aetherdrift, #4464):
//
//	"Flash
//	 When this creature enters, another target creature or Vehicle you
//	 control gets +2/+2 until end of turn."
//
// Daybreak Charger's entry pump with the target narrowed to another
// creature or Vehicle you control. "Another" is object identity
// (Another), so the Guardian can't name itself. An uncrewed Vehicle takes
// the boost too: it changes nothing until the Vehicle becomes a creature
// that turn (CR 301.7a, CR 613.4c). The target is re-checked as the
// trigger resolves (CR 608.2b). Flash is the engine's (CR 702.8a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1e0b9767-8e51-4021-a8fb-22259e2f5347",
		Name:            "Fang Guardian",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(effects.WhenThisEnters("Fang Guardian — another target creature or Vehicle you control gets +2/+2 until end of turn",
				effects.BoostTargetUntilEOT("Fang Guardian — +2/+2", 2, 2)),
				effects.Another(effects.TargetPermanent("another target creature or Vehicle you control",
					effects.Or(effects.Creature(), effects.HasSubtype("Vehicle")), effects.YouControl()))),
				effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2))),
		},
	})
}
