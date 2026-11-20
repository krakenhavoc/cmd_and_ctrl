package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Searchlight Geist — {2}{B} Creature — Spirit 2/1 (Avacyn Restored, #5570):
//
//	"Flying
//	 {3}{B}: This creature gains deathtouch until end of turn. (Any
//	 amount of damage it deals to a creature is enough to destroy it.)"
//
// Flying and Battlefly Swarm's deathtouch row at {3}{B}.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f9366a4a-54d0-4c65-920b-d8ddd475db3c",
		Name:            "Searchlight Geist",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:   "{3}{B}: This creature gains deathtouch until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{3}{B}"),
			Effect:  thisGainsKeywordUntilEndOfTurn("deathtouch", "Searchlight Geist — deathtouch until end of turn"),
		}},
	})
}
