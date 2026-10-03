package effects

// Diamond Lion — Artifact Creature — Cat {2}, 2/2:
//
//	"{T}, Discard your hand, Sacrifice this creature: Add three mana of
//	 any one color. Activate only as an instant."
//
// Lion's Eye Diamond on a body, with a {T} — and the {T} is on a
// CREATURE, so summoning sickness applies (CR 302.6): the Lion has to
// have been under its controller's control since their most recent
// turn began. The engine reads that off the type line for every mana
// ability with a tap cost; the card declares nothing for it.
//
// Everything else is Lion's Eye Diamond's, declared the same way: the
// "Discard your hand" clause (#1600), one colour pick for three mana,
// and the OnlyAsAnInstant Condition. The 2021-06-18 ruling is the
// reason the timing matters: "you can't activate the ability intending
// to use the mana to cast a spell from your hand" — the auto-tapper
// never plans it, and a hand-click is refused in any prompt window.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7ea9bb3b-76e9-4150-9e4f-1cd3abfe8ad7",
		Name:         "Diamond Lion",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost: ManaAbilityCost{
				Tap:          true,
				Sacrifice:    true,
				DiscardCards: DiscardYourHand().DiscardCards,
			},
			Produced:  OneColorOfAmount(3),
			Label:     "{T}, Discard your hand, Sacrifice this creature: Add three mana of any one color. Activate only as an instant.",
			Condition: OnlyAsAnInstant(),
		}},
	})
}
