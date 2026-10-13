package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Oakshade Stalker // Moonlit Ambusher — {2}{G} Creature — Human Ranger
// Werewolf 3/3 // Creature — Werewolf 6/3 (#2586, ADR 0132):
//
//	Front: "You may cast this spell as though it had flash if you pay {2}
//	        more to cast it.
//	        Daybound"
//	Back:  "Nightbound"
//
// The flash clause is FlashForMore (flash_for_more.go): an offer that
// costs the printed {2}{G} plus {2} and opens the instant-speed window
// for that cast only. Daybound is the keyword token; the back face is
// its own entry.
//
// No simplification.
func init() {
	const oracle = "1b639537-a8fe-4615-b937-afa478fdec0f"
	Register(Spec{
		OracleID:         oracle,
		Name:             "Oakshade Stalker",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{FlashForMore("{4}{G}", "{2}")},
		PrintedKeywords:  []string{"daybound"},
	})
	Register(Spec{
		OracleID:        oracle + "#1",
		Name:            "Moonlit Ambusher",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"nightbound"},
	})
}
