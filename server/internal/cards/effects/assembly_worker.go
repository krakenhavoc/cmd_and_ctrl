package effects

// Assembly-Worker — Artifact Creature — Assembly-Worker {3}, 2/2 (Time
// Spiral, #5596):
//
//	"{T}: Target Assembly-Worker creature gets +1/+1 until end of turn."
//
// Mishra's Factory's pump row on a creature: the target must be an
// Assembly-Worker creature at announce and still be one as the ability
// resolves (CR 608.2b). The {T} is a creature's, so it waits out
// summoning sickness (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "751321ac-af07-4b7e-976f-5bf4e4281470",
		Name:         "Assembly-Worker",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: Target Assembly-Worker creature gets +1/+1 until end of turn.",
			Purpose: ForTargets(BenefitUntilEOT(0, 1, 1)),
			Cost:    TapCost(),
			Targets: TargetCreature("target Assembly-Worker creature", Subtype("Assembly-Worker")),
			Effect:  boostTargetUntilEOT("Assembly-Worker — +1/+1", 1, 1),
		}},
	})
}
