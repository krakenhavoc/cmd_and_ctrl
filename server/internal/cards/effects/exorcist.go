package effects

// Exorcist — Creature — Human Cleric {W}{W}, 1/1 (The Dark, #5638):
//
//	"{1}{W}, {T}: Destroy target black creature."
//
// Royal Assassin's row with a mana component and a colour clause in
// place of "tapped". OfColor reads the creature's current colours, at
// announce and again at resolution, so a creature that stops being
// black in response is no longer a legal target and nothing happens.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2a1ca3e2-e1d1-4525-a07e-13d5ba1ac431",
		Name:         "Exorcist",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}{W}, {T}: Destroy target black creature.",
			Cost:    Plus(ManaCost("{1}{W}"), TapCost()),
			Targets: TargetCreature("target black creature", OfColor("B")),
			Effect:  destroyFirstLegalTarget,
		}},
	})
}
