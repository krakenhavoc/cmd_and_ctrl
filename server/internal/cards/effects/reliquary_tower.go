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
// No simplifications. Shipped unreviewed and audited in #74's
// theme-deck pass, where an eighteen-card hand walks through cleanup
// with the Tower out and parks the cursor there the moment it leaves
// (blue_draw_deck_smoke_test.go) — which is the whole card.
func init() {
	Register(Spec{
		OracleID:      "c23e5b80-08d2-4e24-9908-fe2aa4f30f6f",
		Name:          "Reliquary Tower",
		Completeness:  CompletenessFull,
		NoMaxHandSize: true,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
	})
}
