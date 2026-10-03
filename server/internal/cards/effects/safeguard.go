package effects

// Safeguard — Enchantment {3}{W}{W}:
//
//	"{2}{W}: Prevent all combat damage that would be dealt by target creature this turn."
//
// ADR 0108 §7 (#1904, Delivery PR 7b): the shield's source is the
// targeted creature, pinned as the ability resolves (CR 400.7).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "e310c3ab-a729-404d-944f-9b477258495c",
		Name:         "Safeguard",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{shieldAgainstTargetsRow(
			"{2}{W}: Prevent all combat damage that would be dealt by target creature this turn.",
			ManaCost("{2}{W}"), TargetCreature("target creature"), true)},
	})
}
