package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pestilent Wolf — Creature — Wolf {1}{G}, 2/2 (Innistrad: Midnight Hunt, S82):
//
//	"{2}{G}: This creature gains deathtouch until end of turn."
//
// Battlefly Swarm's shape with a different cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "39940633-ef64-408e-ae8f-ade201fe7f06",
		Name:         "Pestilent Wolf",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{G}: This creature gains deathtouch until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{2}{G}"),
			Effect:  effects.ThisGainsKeywordUntilEndOfTurn("deathtouch", "Pestilent Wolf — deathtouch until end of turn"),
		}},
	})
}
