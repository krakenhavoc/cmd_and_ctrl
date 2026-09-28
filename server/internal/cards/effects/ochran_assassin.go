package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ochran Assassin — Creature — Elf Assassin, {1}{B}{G}, 1/1:
//
//	"Deathtouch
//	 All creatures able to block this creature do so."
//
// #1684 leftover: deathtouch rides PrintedKeywords, the Lure clause
// is the Prized Unicorn shape (AllAbleToBlockDoSo). No new machinery.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ef6ecb67-7e49-4d40-84f2-1e7c02960a0d",
		Name:            "Ochran Assassin",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Static:          []game.StaticAbility{AllAbleToBlockDoSo()},
	})
}
