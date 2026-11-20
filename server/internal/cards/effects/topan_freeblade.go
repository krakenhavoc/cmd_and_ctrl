package effects

// Topan Freeblade — Creature — Human Soldier {1}{W}, 2/2 (Magic Origins, #5549):
//
//	"Vigilance (Attacking doesn't cause this creature to tap.)
//	 Renown 1 (When this creature deals combat damage to a player, if it isn't renowned, put a +1/+1 counter on it and it becomes renowned.)"
//
// Keywords only: vigilance, and renown, the engine's keyword trigger
// (game/renown.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9e2cd0d5-1537-412b-9ecd-989ab4ac2700",
		Name:            "Topan Freeblade",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance", "renown 1"},
	})
}
