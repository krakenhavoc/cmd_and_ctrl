package effects

// Dismissive Denial — Instant {2}{U}{U} (Marvel Super Heroes Commander,
// #4455): "Counter target spell. Basic landcycling {2}". Cancel's
// resolution (counterTheTargetSpell) plus the cycling family's
// BasicLandcycling.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "99eff6e9-5911-4f89-bad1-0888067a7783",
		Name:         "Dismissive Denial",
		Completeness: CompletenessFull,
		Targets:      TargetSpell("target spell"),
		OnResolve:    counterTheTargetSpell,
		Activated:    []ActivatedAbility{BasicLandcycling("{2}")},
	})
}
