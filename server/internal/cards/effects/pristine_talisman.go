package effects

// Pristine Talisman — Artifact {3} (EDHREC rank 1353):
//
//	"{T}: Add {C}. You gain 1 life."
//
// A Mind Stone that gains a life instead of cycling itself: the
// lifegain deck's mana rock. The "You gain 1 life" is a RIDER on the
// mana ability — the painland shape with the sign flipped — so it
// happens inside the same atomic activation, after the mana lands in
// the pool, and goes through the ordinary life-change path so every
// "whenever you gain life" trigger (Archangel of Thune, Sanguine
// Bond) sees it. Not a cost: the ability is activatable regardless.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1b3d7fce-e9fe-4176-9d9a-472415826cdd",
		Name:         "Pristine Talisman",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}. You gain 1 life",
			Rider:    b12GainLifeRider(1),
		}},
	})
}
