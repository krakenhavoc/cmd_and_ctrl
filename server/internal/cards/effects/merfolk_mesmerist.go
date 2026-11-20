package effects

// Merfolk Mesmerist — Creature — Merfolk Wizard {1}{U}, 1/2 (Magic 2012,
// #5573):
//
//	"{U}, {T}: Target player mills two cards."
//
// Millstone's row at {U}, {T}: the target player mills two cards, or what
// is left of their library (CR 701.17b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "590d524f-278d-4cf4-ae3f-efcced4517b1",
		Name:         "Merfolk Mesmerist",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{U}, {T}: Target player mills two cards.",
			Cost:    Plus(ManaCost("{U}"), TapCost()),
			Targets: TargetPlayer("target player"),
			Effect:  b31MillChosenPlayer(2),
		}},
	})
}
