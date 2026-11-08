package effects

// Sterling Keykeeper — Creature — Human Mercenary {1}{W}, 2/2 (Outlaws
// of Thunder Junction, #4468):
//
//	"{2}, {T}: Tap target non-Mount creature."
//
// Frostbridge Guard's tap ability at {2} plus {T}, its target narrowed
// to creatures without the Mount subtype. The {T} in the cost means a
// summoning-sick Keykeeper can't activate it (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f893d3d6-efef-4394-8e15-e01deed72b4f",
		Name:         "Sterling Keykeeper",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}, {T}: Tap target non-Mount creature.",
			Cost:    Plus(ManaCost("{2}"), TapCost()),
			Targets: TargetCreature("target non-Mount creature", Not(Subtype("Mount"))),
			Effect:  tapChosenPermanent,
		}},
	})
}
