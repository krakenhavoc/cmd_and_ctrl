package effects

// Dark Fortress — Land:
//
//	"{T}: Add {C}.
//	 {T}: Add {B} or {R}. Activate only if this land entered this turn or
//	 if you control a basic land."
//
// The second ability is an ordinary mana ability with an activation
// Condition, so the click path, the enumerator and the auto-tapper all
// refuse it when the clause is false (mshB3EnteredOrControlsBasic).
//
// No simplification. Marvel Super Heroes (#4455).
func init() {
	Register(Spec{
		OracleID:     "40760bfa-a423-487c-ba29-043b2d00c736",
		Name:         "Dark Fortress",
		Completeness: CompletenessFull,
		ManaAbilities: mshB3ConditionalDualLand("{B|R}",
			"Add {B} or {R}. Activate only if this land entered this turn or if you control a basic land."),
	})
}
