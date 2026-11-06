package effects

// Sterling Keykeeper — Creature — Human Mercenary {1}{W}, 2/2 (Outlaws of
// Thunder Junction, #4468):
//
//	"{2}, {T}: Tap target non-Mount creature."
//
// Arashin Sunshield's tapper (tapChosenPermanent) with the target clause
// narrowed to creatures that aren't Mounts. {T} is a creature's tap cost,
// so the Keykeeper can't activate the turn it arrives.
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
			Targets: TargetCreature("target non-Mount creature", Not(OfSubtype("Mount"))),
			Effect:  tapChosenPermanent,
		}},
	})
}
