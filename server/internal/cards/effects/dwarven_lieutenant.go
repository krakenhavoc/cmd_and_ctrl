package effects

// Dwarven Lieutenant — Creature — Dwarf Soldier {R}{R}, 1/2 (Fallen
// Empires, #5637):
//
//	"{1}{R}: Target Dwarf creature gets +1/+0 until end of turn."
//
// The target must be a Dwarf as the ability is activated and again as
// it resolves (CR 608.2b); the Lieutenant is a Dwarf, so it may pump
// itself. No {T} in the cost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "06e18c42-d158-494d-979c-c2ac074a8ee3",
		Name:         "Dwarven Lieutenant",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}{R}: Target Dwarf creature gets +1/+0 until end of turn.",
			Purpose: ForTargets(BenefitUntilEOT(0, 1, 0)),
			Cost:    ManaCost("{1}{R}"),
			Targets: TargetCreature("target Dwarf creature", HasSubtype("Dwarf")),
			Effect:  pumpAndGrantFirstTarget(1, 0, "Dwarven Lieutenant — +1/+0 until end of turn"),
		}},
	})
}
