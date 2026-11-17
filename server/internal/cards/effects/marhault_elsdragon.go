package effects

// Marhault Elsdragon — Legendary Creature — Elf Warrior {3}{R}{R}{G}, 4/6 (Legends, #5769):
//
//	"Rampage 1 (Whenever this creature becomes blocked, it gets +1/+1
//	 until end of turn for each creature blocking it beyond the first.)"
//
// Rampage is the engine's keyword trigger (CR 702.23, game/rampage.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c202f124-4283-48f6-aef1-470ad9ded22b",
		Name:            "Marhault Elsdragon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"rampage 1"},
	})
}
