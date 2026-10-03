package effects

// Dark Sphere — Artifact {0}:
//
//	"{T}, Sacrifice this artifact: The next time a source of your choice would deal damage to you this turn, prevent half that damage, rounded down."
//
// ADR 0108 §7 decision 4 (#1904): the next-damage shield (CR 615.8) with
// Half: it prevents half of that instance's damage, rounded down as
// printed (CR 107.1a), and is spent by it. Half of 1 damage is none, so
// a single point is dealt and the shield waits for the source's next
// damage — a shield that prevents nothing is not used up (CR 609.7b).
//
// No simplifications.
func init() {
	shield := PreventNextDamageFromChosenSource(ShieldYou)
	shield.Half = true
	Register(Spec{
		OracleID:     "397f53f7-f801-4442-a778-2f26ac246b62",
		Name:         "Dark Sphere",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{nextDamageShieldRow(
			"{T}, Sacrifice this artifact: The next time a source of your choice would deal damage to you this turn, prevent half that damage, rounded down.",
			Plus(TapCost(), SacrificeThis()), nil, shield)},
	})
}
