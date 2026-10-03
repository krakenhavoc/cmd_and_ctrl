package effects

// Pillar of the Paruns — Land:
//
//	"{T}: Add one mana of any color. Spend this mana only to cast a
//	 multicolored spell."
//
// A five-colour land for gold decks, paid for by everything it cannot
// cast. The restriction is two tags on the token (#1600): "cast" and
// "multicolored" (CR 105.2b — two or more colours). A hybrid spell is
// every colour of its hybrid symbols (CR 202.2d), so a {R/G} spell is
// multicolored and the Pillar pays for it; a monocolored or colourless
// spell, and every activated ability, is refused.
//
// "Any color", not "any color in your commander's color identity", so
// the pipe keeps all five (NarrowToCommanderIdentity off). The
// auto-tapper does not plan a restricted source (ADR 0040 §7), so the
// Pillar is tapped by hand, as on paper.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "677b8ce7-f922-4ee3-b311-f199da9b352b",
		Name:         "Pillar of the Paruns",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:         ManaAbilityCost{Tap: true},
			Produced:     "{W|U|B|R|G}",
			Restrictions: []string{ManaRestrictCast, ManaRestrictMulticolored},
			Label:        "Add one mana of any color. Spend this mana only to cast a multicolored spell",
		}},
	})
}
