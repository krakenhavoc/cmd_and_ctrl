package effects

// Lightbringer — Creature — Kor Rebel {2}{W}, 2/2 (Nemesis, #5618):
//
//	"{T}, Sacrifice this creature: Exile target black creature."
//
// Lawbringer's shape with black: a tap-and-sacrifice cost (a {T} cost,
// so not the turn it came under your control, CR 302.6) and the shared
// exileFirstLegalCardTarget body. A creature that stopped being black in
// response is no longer a legal target and is not exiled (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8656d9c1-d890-4473-9da8-7ec339d7d546",
		Name:         "Lightbringer",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}, Sacrifice this creature: Exile target black creature.",
			Cost:    Plus(TapCost(), SacrificeThis()),
			Targets: TargetCreature("target black creature", OfColor("B")),
			Effect:  exileFirstLegalCardTarget,
		}},
	})
}
