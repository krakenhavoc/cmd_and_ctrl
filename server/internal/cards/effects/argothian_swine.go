package effects

// Argothian Swine — Creature — Boar {3}{G}, 3/3 (Urza's Saga, #5624):
//
//	"Trample"
//
// Registered so the creature resolves and enters on its own rather than
// through the manual sandbox path. Trample is the printed keyword the
// combat-damage assignment reads (CR 702.19a); the cost, the type and the
// 3/3 ride the card from Scryfall.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "80e08ee2-4272-4783-9a0d-90079600b594",
		Name:            "Argothian Swine",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
	})
}
