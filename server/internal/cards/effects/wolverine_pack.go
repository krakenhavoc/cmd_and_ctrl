package effects

// Wolverine Pack — Creature — Wolverine {2}{G}{G}, 2/4 (Legends, #5769):
//
//	"Rampage 2 (Whenever this creature becomes blocked, it gets +2/+2
//	 until end of turn for each creature blocking it beyond the first.)"
//
// Rampage is the engine's keyword trigger (CR 702.23, game/rampage.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d67d9d82-2786-48fc-8107-11d12c5813d0",
		Name:            "Wolverine Pack",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"rampage 2"},
	})
}
