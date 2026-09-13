package effects

// Elves of Deep Shadow — Creature — Elf Druid {G}, 1/1 (EDHREC rank
// 731):
//
//	"{T}: Add {B}. This creature deals 1 damage to you."
//
// A painland on legs: the one-drop dork that fixes into black at a
// life a tap. The damage is a RIDER on the mana ability, not a cost
// (PainRider — see mana_rider_helpers.go for why that distinction is
// load-bearing: the Elf stays activatable at 1 life, the damage is
// damage rather than life loss, and its source is the creature). CR
// 302.1 summoning sickness applies, engine-side.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "20347559-95a9-4689-bb79-c5bb3809b719",
		Name:         "Elves of Deep Shadow",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{B}",
			Label:    "Add {B}. This creature deals 1 damage to you.",
			Rider:    PainRider(1),
		}},
	})
}
