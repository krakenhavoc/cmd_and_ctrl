package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fencer Clique — Creature — Faerie Soldier {2}{U}{U}, 3/2 (Morningtide,
// #5592):
//
//	"Flying
//	 {U}: Put this creature on top of its owner's library."
//
// Wayward Soul's text on a different body: the activation puts the
// creature on top of its owner's library, an escape from removal.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "58374ee4-0497-4d12-847c-0f67555f8399",
		Name:            "Fencer Clique",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:   "{U}: Put this creature on top of its owner's library.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    ManaCost("{U}"),
			Effect:  putThisPermanentOnTopOfOwnersLibrary,
		}},
	})
}
