package effects

// Hoof Skulkin — Artifact Creature — Scarecrow {3}, 2/2 (Eventide, #5590):
//
//	"{3}: Target green creature gets +1/+1 until end of turn."
//
// The colour is read off the target as the ability is announced and again
// as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7f60549f-7a0a-4009-a51a-d3f51214be11",
		Name:         "Hoof Skulkin",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{3}: Target green creature gets +1/+1 until end of turn.",
			Purpose: ForTargets(BenefitUntilEOT(0, 1, 1)),
			Cost:    ManaCost("{3}"),
			Targets: TargetCreature("target green creature", OfColor("G")),
			Effect:  boostTargetUntilEOT("Hoof Skulkin — +1/+1 until end of turn", 1, 1),
		}},
	})
}
