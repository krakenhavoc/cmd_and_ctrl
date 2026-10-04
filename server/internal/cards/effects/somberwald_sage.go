package effects

// Somberwald Sage — Creature — Human Druid {2}{G}, 0/1:
//
//	"{T}: Add three mana of any one color. Spend this mana only to cast
//	 creature spells."
//
// Gilded Lotus's one pick of three (OneColorOfAmount) with Food
// Chain's restriction: a cast, of a creature spell. The tag pays any
// part of that spell's total cost, kicker and alternative costs
// included, and never an activated ability — unearth and ninjutsu
// abilities, or a noncreature spell that makes creature tokens (the
// 2012-05-01 rulings). It is a creature's {T} ability, so it can't be
// activated the turn the Sage arrives (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f249db15-09fa-4a1c-9461-9ae803c219b6",
		Name:         "Somberwald Sage",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:         ManaAbilityCost{Tap: true},
			Produced:     OneColorOfAmount(3),
			Label:        "Add three mana of any one color. Spend this mana only to cast creature spells",
			Restrictions: []string{ManaRestrictCast, ManaRestrictType("Creature")},
		}},
	})
}
