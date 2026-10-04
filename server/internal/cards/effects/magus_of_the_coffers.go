package effects

// Magus of the Coffers — Creature — Human Wizard {4}{B}, 4/4:
//
//	"{2}, {T}: Add {B} for each Swamp you control."
//
// Cabal Coffers on a creature: the same mana ability, the same
// effective-subtype Swamp count (so Urborg, Tomb of Yawgmoth grows
// it), the same {2} mana component. The {T} makes it subject to
// summoning sickness, which the engine applies to a creature's own
// tap-cost abilities. Not auto-tappable, like every ability with a
// mana component.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "23dd895c-92bd-4af0-8b1a-7d76ca49178f",
		Name:         "Magus of the Coffers",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:         ManaAbilityCost{Tap: true, Mana: "{2}"},
			ProducedFunc: ProducedPerPermanent("B", MatchLandSubtype("Swamp")),
			Label:        "{2}, {T}: Add {B} for each Swamp you control",
		}},
	})
}
