package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scourge of the Undercity — Creature — Gorgon Vampire {1}{B}, 2/1
// (Foundations Jumpstart, #5464):
//
//	"Lifelink (Damage dealt by this creature also causes you to gain
//	 that much life.)
//	 When this creature enters, another target creature you control
//	 gains lifelink until end of turn."
//
// Sanguine Savior's grant on an enters trigger: "another" is object
// identity (Another), so the Scourge can't name itself, and a target
// gone by resolution gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	const label = "Scourge of the Undercity — another target creature you control gains lifelink until end of turn"
	effects.Register(effects.Spec{
		OracleID:        "541ba65b-9b4e-45bb-9055-4f7560563425",
		Name:            "Scourge of the Undercity",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(effects.WhenThisEnters(label, effects.JmpB7FirstLegalTargetGains(label, "lifelink")),
				effects.Another(effects.TargetCreature("another target creature you control", effects.YouControl()))),
				effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "lifelink"))),
		},
	})
}
