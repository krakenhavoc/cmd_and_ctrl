package effects

// Rograkh, Son of Rohgahh — Legendary Creature — Kobold Warrior {0},
// 0/1 (EDHREC rank 2346):
//
//	"First strike, menace, trample
//	 Partner (You can have two commanders if both have partner.)"
//
// The free partner. Three printed combat keywords the engine already
// honours, declared through PrintedKeywords; Partner is a
// deck-construction rule, not a game action, and needs nothing on the
// card (Kediss's posture). The body and cost are printed card data.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "584cee10-f18c-4633-95cc-f2e7a11841ac",
		Name:            "Rograkh, Son of Rohgahh",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike", "menace", "trample"},
	})
}
