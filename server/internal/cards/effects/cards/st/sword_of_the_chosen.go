package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sword of the Chosen — Legendary Artifact {2} (Stronghold, #5627):
//
//	"{T}: Target legendary creature gets +2/+2 until end of turn."
//
// Willow Satyr's "target legendary creature" clause and Angelic Page's
// shared boostTargetUntilEOT body at +2/+2, ending in the cleanup step
// (CR 514.2). It is an artifact that isn't a creature, so the tap can be
// paid the turn it arrives (CR 302.6 applies to creatures only). It is
// not Equipment and attaches to nothing. A target that stops being
// legendary in response is no longer legal and gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3c756eda-4fc1-4766-99be-32d8c9f35262",
		Name:         "Sword of the Chosen",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Target legendary creature gets +2/+2 until end of turn.",
			Purpose: effects.WithAnswers(effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2)), game.AnswerPump),
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target legendary creature", effects.Legendary()),
			Effect:  effects.BoostTargetUntilEOT("Sword of the Chosen — +2/+2", 2, 2),
		}},
	})
}
