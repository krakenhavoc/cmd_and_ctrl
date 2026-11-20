package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Maximize Velocity — Sorcery {R} (Guilds of Ravnica, #5525):
//
//	"Target creature gets +1/+1 and gains haste until end of turn.
//	 Jump-start"
//
// The shared socB27TargetGetsAndGains (both halves end at cleanup, CR
// 514.2), with jump-start (cmdr2S39JumpStart, CR 702.133a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "96e7e4c8-f66d-4327-aee6-85da648418f5",
		Name:             "Maximize Velocity",
		Completeness:     CompletenessFull,
		Targets:          TargetCreature("target creature"),
		Purpose:          ForTargets(BenefitUntilEOT(0, 1, 1, "haste")),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{cmdr2S39JumpStart("{R}")},
		OnResolve:        socB27TargetGetsAndGains(1, 1, "haste", "Maximize Velocity"),
	})
}
