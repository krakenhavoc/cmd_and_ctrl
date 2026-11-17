package effects

// Disruptive Student — Creature — Human Wizard {2}{U}, 1/1 (Urza's Saga):
//
//	"{T}: Counter target spell unless its controller pays {1}."
//
// Spiketail Hatchling's tax with a tap cost in place of the sacrifice.
// The spell's controller is asked as the ability resolves; a spell that
// has left the stack by then makes it do nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "74315869-46e3-43a2-abc0-d86d0a70d343",
		Name:         "Disruptive Student",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: Counter target spell unless its controller pays {1}.",
			Cost:    TapCost(),
			Targets: TargetSpell("target spell"),
			Effect:  counterTheTargetSpellUnlessPaid("Disruptive Student", "{1}"),
		}},
	})
}
