package effects

// Orcish Lumberjack — Creature — Orc {R}, 1/1 (EDHREC rank 2914):
//
//	"{T}, Sacrifice a Forest: Add three mana in any combination of
//	 {R} and/or {G}."
//
// The one-drop that turns a Forest into three mana. A mana ability
// with a tap and a sacrifice-another cost (Phyrexian Tower's shape)
// narrowed to a Forest — any permanent with the subtype, effective
// so a Dryad Arbor or a Forest-typed dual qualifies — producing
// three independent pipe slots: "{R|G}{R|G}{R|G}" resolves to any
// combination of red and green, one colour pick per slot (Mystic
// Gate's three-way output). Printed colours, so no commander-
// identity narrowing. A tap on a creature: summoning sickness
// applies, and the engine enforces it inside ActivateManaAbility.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "383e3fe4-8558-4561-8632-6eadb5d5963c",
		Name:         "Orcish Lumberjack",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost: ManaAbilityCost{
				Tap:            true,
				SacrificeOther: b27SacrificeAForest(),
			},
			Produced: "{R|G}{R|G}{R|G}",
			Label:    "{T}, Sacrifice a Forest: Add three mana in any combination of {R} and/or {G}",
		}},
	})
}
