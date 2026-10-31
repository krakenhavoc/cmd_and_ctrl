package effects

// Marvelous Melee — Instant {4}{R} (Marvel Super Heroes Commander,
// #4455): "Marvelous Melee deals 6 damage to target creature. Basic
// landcycling {2}". Concentrated Fire's damage body at six damage, with
// the cycling family's shared BasicLandcycling.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5ee9940a-0569-45b6-b28c-0ef3914013a9",
		Name:         "Marvelous Melee",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      ForTargets(DamageToTarget(0, 6)),
		OnResolve:    damageToFirstTarget(6),
		Activated:    []ActivatedAbility{BasicLandcycling("{2}")},
	})
}
