package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Battle Rampart — Creature — Wall {2}{R}, 1/3 (Battlebond):
//
//	"Defender
//	 {T}: Target creature gains haste until end of turn."
//
// Akki Drillmaster's tap ability over a defender body; the haste ends at
// cleanup (CR 514.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ec187304-4c05-44da-8ad5-7c56f6b05212",
		Name:            "Battle Rampart",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Target creature gains haste until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "haste")),
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.NeoB07TargetGetsUntilEOT("Battle Rampart — haste until end of turn", 0, 0, "haste"),
		}},
	})
}
