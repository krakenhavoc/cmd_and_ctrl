package effects

// Indomitable Ancients — Creature — Treefolk Warrior {2}{W}{W}, 2/10
// (EDHREC rank 4526):
//
//	(no rules text)
//
// Ten toughness for four mana and nothing else. Commander plays it
// for exactly one reason: toughness-matters decks — Assault Formation,
// Doran, the Siege Tower, High Alert — turn that 10 into ten points
// of attack, and Arcades, the Strategist decks want a wall that is
// not a Wall. The card contributes nothing on its own; it contributes
// everything to the deck that counts its toughness.
//
// Registered for the same reason Gigantosaurus is: an unregistered
// card falls back to the manual sandbox, and a vanilla body should
// never need a human in the loop.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2764741c-1f3e-459a-a487-54cbac2c84a6",
		Name:         "Indomitable Ancients",
		Completeness: CompletenessFull,
	})
}
