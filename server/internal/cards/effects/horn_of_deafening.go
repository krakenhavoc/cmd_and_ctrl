package effects

// Horn of Deafening — Artifact {4}:
//
//	"{2}, {T}: Prevent all combat damage that would be dealt by target creature this turn."
//
// ADR 0108 §7 (#1904, Delivery PR 7b): the shield's source is the
// targeted creature, pinned as the ability resolves (CR 400.7).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "50a1c14a-003f-424b-bb8e-2e2d51465a90",
		Name:         "Horn of Deafening",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{shieldAgainstTargetsRow(
			"{2}, {T}: Prevent all combat damage that would be dealt by target creature this turn.",
			Plus(ManaCost("{2}"), TapCost()), TargetCreature("target creature"), true)},
	})
}
