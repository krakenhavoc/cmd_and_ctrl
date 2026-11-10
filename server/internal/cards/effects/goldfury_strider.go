package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "7534d0b3-4211-4add-8e8d-370c033518d3",
		Name:            "Goldfury Strider",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Activated: []ActivatedAbility{{
			Label: "Tap two untapped artifacts and/or creatures you control: Target creature gets +2/+0 until end of turn. Activate only as a sorcery.",
			Cost: game.AbilityCost{TapOthers: &game.TapOthersCost{
				Count:  2,
				Filter: TargetPermanent(tapTwo, Or(Artifact(), Creature())),
				Label:  tapTwo,
			}},
			Purpose:      ForTargets(BenefitUntilEOT(0, 2, 0)),
			Targets:      TargetCreature("target creature"),
			SorcerySpeed: true,
			Effect:       boostTargetUntilEOT("Goldfury Strider — +2/+0", 2, 0),
		}},
	})
}
