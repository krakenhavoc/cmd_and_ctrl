package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Brave-Kin Duo — Creature — Rabbit Mouse {W}, 1/1 (Bloomburrow, #4467):
//
//	"{1}, {T}: Target creature gets +1/+1 until end of turn. Activate
//	 only as a sorcery."
//
// Alaborn Veteran's pump body (boostTargetUntilEOT) at sorcery speed: a
// target that has left in response is skipped (CR 608.2b). The tap
// cost keeps it from being activated the turn it arrives (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1f5b7f4c-a3d2-4472-80f5-8504562fc7fd",
		Name:         "Brave-Kin Duo",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:        "{1}, {T}: Target creature gets +1/+1 until end of turn. Activate only as a sorcery.",
			Purpose:      ForTargets(BenefitUntilEOT(0, 1, 1)),
			Cost:         game.AbilityCost{Mana: "{1}", Tap: true},
			Targets:      TargetCreature("target creature"),
			SorcerySpeed: true,
			Effect:       boostTargetUntilEOT("Brave-Kin Duo — +1/+1", 1, 1),
		}},
	})
}
