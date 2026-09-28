package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Brontotherium — Creature — Beast, {4}{G}{G}, 5/3:
//
//	"Trample
//	 Provoke (Whenever this creature attacks, you may have target
//	 creature defending player controls untap and block it if able.)"
//
// #1684 leftover: trample rides PrintedKeywords, Provoke rides the
// Provoke() trigger constructor built for Goblin Grappler. No new
// machinery.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6e3301a5-a99b-4b17-aa40-34ccc5904977",
		Name:            "Brontotherium",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered:       []game.TriggeredAbility{Provoke()},
	})
}
