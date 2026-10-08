package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Burly Breaker // Dire-Strain Demolisher — {3}{G}{G} Creature — Human
// Werewolf 6/5 // Creature — Werewolf 8/7 (#2586, ADR 0132):
//
//	Front: "Ward {1}
//	        Daybound"
//	Back:  "Ward {3}
//	        Nightbound"
//
// Ward is the triggered ability ward.go builds, one face to each cost.
//
// No simplification.
func init() {
	const oracle = "4cbfb898-e1ca-4e34-9cb2-11ba50272984"
	Register(Spec{
		OracleID:        oracle,
		Name:            "Burly Breaker",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"daybound"},
		Triggered:       []game.TriggeredAbility{Ward(WardMana("{1}"), "Burly Breaker — ward {1}")},
	})
	Register(Spec{
		OracleID:        oracle + "#1",
		Name:            "Dire-Strain Demolisher",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"nightbound"},
		Triggered:       []game.TriggeredAbility{Ward(WardMana("{3}"), "Dire-Strain Demolisher — ward {3}")},
	})
}
