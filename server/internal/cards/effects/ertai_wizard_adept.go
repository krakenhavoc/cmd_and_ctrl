package effects

// Ertai, Wizard Adept — Legendary Creature — Human Wizard {2}{U}, 1/1
// (Exodus, #5626):
//
//	"{2}{U}{U}, {T}: Counter target spell."
//
// Kozilek, the Great Distortion's activated counter with a mana-and-tap
// cost: counterTheChosenSpell counters the first target still legal as
// the ability resolves (CR 608.2b, CR 701.6a). The tap cost means it
// can't be activated the turn Ertai comes under your control
// (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f2902dfd-888f-4b02-b8ad-1056f0059fc3",
		Name:         "Ertai, Wizard Adept",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}{U}{U}, {T}: Counter target spell.",
			Cost:    Plus(ManaCost("{2}{U}{U}"), TapCost()),
			Targets: TargetSpell("target spell"),
			Effect:  counterTheChosenSpell,
		}},
	})
}
