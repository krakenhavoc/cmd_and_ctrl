package effects

// Chandler — Legendary Creature — Human Rogue {4}{R}, 3/3 (Homelands, #5635):
//
//	"{R}{R}{R}, {T}: Destroy target artifact creature."
//
// The {T} keeps him off the turn he arrives (CR 602.5a). The target
// must be both an artifact and a creature, judged at announce and
// again at resolution (CR 608.2b); a regeneration shield still saves it
// (CR 701.19a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "520abf49-7631-4e7f-889e-547ccc8c3b95",
		Name:         "Chandler",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{R}{R}{R}, {T}: Destroy target artifact creature.",
			Cost:    Plus(ManaCost("{R}{R}{R}"), TapCost()),
			Targets: TargetCreature("target artifact creature", Artifact()),
			Effect:  destroyFirstLegalTarget,
		}},
	})
}
