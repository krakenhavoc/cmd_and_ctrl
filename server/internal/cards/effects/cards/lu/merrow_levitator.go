package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Merrow Levitator — Creature — Merfolk Wizard {3}{U}, 2/3 (Eventide, #5590):
//
//	"{T}: Target creature gains flying until end of turn.
//	 Whenever you cast a blue spell, you may untap this creature."
//
// Flying Carpet's grant as a tap ability, and the optional cast-trigger
// untap (eveB1UntapOnCast).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a4154dac-3f6d-4f8f-a769-9f5b8acecda7",
		Name:         "Merrow Levitator",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Target creature gains flying until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "flying")),
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.EbTargetCreatureGainsUntilEOT("flying", "Merrow Levitator — flying until end of turn"),
		}},
		Triggered: []game.TriggeredAbility{effects.EveB1UntapOnCast("Merrow Levitator", "U")},
	})
}
