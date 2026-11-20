package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fang Skulkin — Artifact Creature — Scarecrow {2}, 2/1 (Eventide, #5590):
//
//	"{2}: Target black creature gains wither until end of turn.
//	 (It deals damage to creatures in the form of -1/-1 counters.)"
//
// The colour is read off the target as the ability is announced and again
// as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "54bed864-fe2e-4359-b0ac-c27a2c225487",
		Name:         "Fang Skulkin",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}: Target black creature gains wither until end of turn.",
			Purpose: ForTargets(BenefitUntilEOT(0, 0, 0, game.KeywordWither)),
			Cost:    ManaCost("{2}"),
			Targets: TargetCreature("target black creature", OfColor("B")),
			Effect:  ebTargetCreatureGainsUntilEOT(game.KeywordWither, "Fang Skulkin — wither until end of turn"),
		}},
	})
}
