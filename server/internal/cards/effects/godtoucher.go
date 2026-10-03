package effects

// Godtoucher — Creature — Elf Cleric {3}{G}, 2/2:
//
//	"{1}{W}, {T}: Prevent all damage that would be dealt to target
//	 creature with power 5 or greater this turn."
//
// ADR 0108 §7, Delivery PR 7 (#1904): the not-one-use shield pinned to
// the target. The power is checked as the target is chosen and again as
// the ability resolves (CR 608.2b); the shield then lasts the turn,
// whatever the creature's power becomes.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "3094f214-5aa8-4003-8a1e-035485dfb978",
		Name:         "Godtoucher",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{shieldTargetRow(
			"{1}{W}, {T}: Prevent all damage that would be dealt to target creature with power 5 or greater this turn.",
			Plus(ManaCost("{1}{W}"), TapCost()),
			TargetCreature("target creature with power 5 or greater", PowerGE(5)))},
	})
}
