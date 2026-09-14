package effects

// Karametra's Acolyte — Creature — Human Druid {3}{G}, 1/4 (EDHREC
// rank 3024):
//
//	"{T}: Add an amount of {G} equal to your devotion to green. (Each
//	 {G} in the mana costs of permanents you control counts toward
//	 your devotion to green.)"
//
// The green devotion dork. One mana ability whose amount is derived
// at activation (ProducedFunc — the Cabal Coffers shape): devotion to
// green over the permanents the controller controls, read by
// devotionTo (CR 700.5 — hybrid symbols count for every colour they
// offer, the Acolyte's own {G} included), repeated as {G} slots.
// Zero devotion adds nothing and still taps; summoning sickness
// applies, as the engine enforces for any tap mana ability on a
// creature.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e8cf80a1-6908-4adb-88e7-2015672d4905",
		Name:         "Karametra's Acolyte",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:         ManaAbilityCost{Tap: true},
			ProducedFunc: b28ProducedDevotion("G"),
			Label:        "Add {G} equal to your devotion to green",
		}},
	})
}
