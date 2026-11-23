package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goldfury Strider — Artifact Creature — Golem {4}{R}, 3/5 (The Lost
// Caverns of Ixalan, #4470):
//
//	"Trample
//	 Tap two untapped artifacts and/or creatures you control: Target
//	 creature gets +2/+0 until end of turn. Activate only as a sorcery."
//
// Adaptive Gemguard's cost (TapOthersCost: two untapped artifacts and/or
// creatures you control, the Strider itself allowed, since the cost has
// no {T} symbol and CR 302.6 does not reach it) with Avatar Roku's
// targeted +2/+0 body. A target that has gone by resolution gets nothing
// (CR 608.2b).
//
// No simplification.
func init() {
	const tapTwo = "two untapped artifacts and/or creatures you control"
	effects.Register(effects.Spec{
		OracleID:        "7534d0b3-4211-4add-8e8d-370c033518d3",
		Name:            "Goldfury Strider",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Activated: []effects.ActivatedAbility{{
			Label: "Tap two untapped artifacts and/or creatures you control: Target creature gets +2/+0 until end of turn. Activate only as a sorcery.",
			Cost: game.AbilityCost{TapOthers: &game.TapOthersCost{
				Count:  2,
				Filter: effects.TargetPermanent(tapTwo, effects.Or(effects.Artifact(), effects.Creature())),
				Label:  tapTwo,
			}},
			Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 2, 0)),
			Targets:      effects.TargetCreature("target creature"),
			SorcerySpeed: true,
			Effect:       effects.BoostTargetUntilEOT("Goldfury Strider — +2/+0", 2, 0),
		}},
	})
}
