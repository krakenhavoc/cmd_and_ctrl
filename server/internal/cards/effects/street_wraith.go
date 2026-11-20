package effects

// Street Wraith — Creature — Wraith {3}{B}{B}, 3/4 (Future Sight, #5594):
//
//	"Swampwalk (This creature can't be blocked as long as defending
//	 player controls a Swamp.)
//	 Cycling—Pay 2 life. (Pay 2 life, Discard this card: Draw a card.)"
//
// Cycling whose cost is life instead of mana (CR 702.29a): the keyword's
// own ability, hand-only and emitting EventCycle, with its cost swapped
// for "Pay 2 life, Discard this card". A player with less than 2 life
// can't pay it (CR 119.4).
//
// No simplification.
func init() {
	cycling := Cycling("{0}")
	cycling.Label = "Cycling—Pay 2 life. (Pay 2 life, Discard this card: Draw a card.)"
	cycling.Cost = Plus(PayLife(2), DiscardThis())
	Register(Spec{
		OracleID:        "bd7867f4-4f58-49c2-9c24-a124dfdb506d",
		Name:            "Street Wraith",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"swampwalk"},
		Activated:       []ActivatedAbility{cycling},
	})
}
