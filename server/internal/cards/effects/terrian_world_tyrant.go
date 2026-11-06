package effects

// Terrian, World Tyrant — Legendary Creature — Dinosaur Ooze {2}{G}{G}{G},
// 9/7 (Aetherdrift, #4464):
//
//	(no rules text)
//
// A legendary vanilla, registered for the reason Sundial, Dawn Tyrant
// gives: so the spell resolves and the creature enters without the
// manual sandbox path. Everything printed rides the card from Scryfall.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "856245a6-c697-4f03-90be-77d8eb2d274e",
		Name:         "Terrian, World Tyrant",
		Completeness: CompletenessFull,
	})
}
