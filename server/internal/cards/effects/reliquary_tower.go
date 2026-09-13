package effects

// Reliquary Tower — Land:
//
//	"You have no maximum hand size."
//	"{T}: Add {C}."
//
// The land every draw deck plays, and the reason it is in THIS
// sprint: a deck built to draw four cards a turn spends the cleanup
// step throwing them away without one.
//
// Both halves already have machinery. Spec.NoMaxHandSize is the
// derived player-level static Thought Vessel and Decanter of Endless
// Water use — the engine asks the battlefield at cleanup rather than
// writing to Player.MaxHandSize, so two of them, and one of two
// leaving, come out right with no bookkeeping. The mana is a plain
// colourless tap ability.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:      "c23e5b80-08d2-4e24-9908-fe2aa4f30f6f",
		Name:          "Reliquary Tower",
		NoMaxHandSize: true,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
	})
}
