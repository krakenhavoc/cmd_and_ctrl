package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sootstoke Kindler — Creature — Elemental Shaman {1}{B/R}, 1/1 (Shadowmoor, S82
// OLD b531, #5591):
//
//	"Haste
//	 {T}: Target black or red creature gains haste until end of turn."
//
// Battle Rampart's tap ability narrowed to a black or red target. Colour is read
// through the target's changed words (HasColorWord, CR 612.2). The haste ends at
// cleanup (CR 514.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "16a96727-c53a-4563-8151-f8b7abb7e534",
		Name:            "Sootstoke Kindler",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Target black or red creature gains haste until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "haste")),
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target black or red creature", effects.Or(effects.OfColor("B"), effects.OfColor("R"))),
			Effect:  effects.NeoB07TargetGetsUntilEOT("Sootstoke Kindler — haste until end of turn", 0, 0, "haste"),
		}},
	})
}
