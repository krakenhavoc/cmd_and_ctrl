package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pestilent Wolf — Creature — Wolf {1}{G}, 2/2 (Innistrad: Midnight Hunt, S82):
//
//	"{2}{G}: This creature gains deathtouch until end of turn."
//
// Battlefly Swarm's shape with a different cost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "39940633-ef64-408e-ae8f-ade201fe7f06",
		Name:         "Pestilent Wolf",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}{G}: This creature gains deathtouch until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{2}{G}"),
			Effect:  thisGainsKeywordUntilEndOfTurn("deathtouch", "Pestilent Wolf — deathtouch until end of turn"),
		}},
	})
}
