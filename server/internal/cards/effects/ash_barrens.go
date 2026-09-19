package effects

// Ash Barrens — Land:
//
//	"{T}: Add {C}.
//	 Basic landcycling {1} ({1}, Discard this card: Search your
//	 library for a basic land card, reveal it, put it into your hand,
//	 then shuffle.)"
//
// The batch-01 triage filed this under "casting and playing from
// zones other than hand", which is the right shape and the wrong
// verb: cycling is not a cast, it is an ACTIVATED ability that
// functions from the hand (CR 113.6, CR 702.29a), and #660 gave
// `ActivatedAbility.Zones` the dimension it needs. `BasicLandcycling`
// is the constructor, and it stamps the three things a hand-written
// version would forget — the hand zone, the `DiscardThis()` cost
// component, and the `Cycling` bit that makes EventCycle fire so
// Astral Slide and Drake Haven see it.
//
// A land that is a spell-less permanent with an ability from hand is
// the whole point of the card: an Ash Barrens is an untapped
// colourless land when you have the mana and any basic when you do
// not, and it never enters tapped either way.
//
// The colourless ability IS declared, because the type line is a bare
// "Land" with no basic land type for game.ManaAbilitiesForCard to
// derive one from (CR 305.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "58257464-278e-45fa-8e0b-bcd9a7500bc1",
		Name:         "Ash Barrens",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{BasicLandcycling("{1}")},
	})
}
