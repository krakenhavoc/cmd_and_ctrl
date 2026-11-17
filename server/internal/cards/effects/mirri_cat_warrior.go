package effects

// Mirri, Cat Warrior — Legendary Creature — Cat Warrior {1}{G}{G}, 2/3
// (Exodus, #5626):
//
//	"First strike, forestwalk, vigilance"
//
// Three printed keywords, each enforced by the engine: first strike in
// the combat damage step (CR 702.7b), forestwalk in block legality — it
// can't be blocked while the defending player controls a Forest
// (CR 702.14c) — and vigilance as attackers are declared (CR 702.20b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f01efb04-ad2b-46fb-852c-dfb7e523e1b7",
		Name:            "Mirri, Cat Warrior",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike", "forestwalk", "vigilance"},
	})
}
