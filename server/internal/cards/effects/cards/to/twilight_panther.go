package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Twilight Panther — Creature — Cat Spirit {W}, 1/2 (Ravnica: Clue Edition,
// S82):
//
//	"{B}: This creature gains deathtouch until end of turn."
//
// Battlefly Swarm's deathtouch pump with an off-colour cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5e19aa52-5b76-40d7-ba9c-0e5d64f2b773",
		Name:         "Twilight Panther",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}: This creature gains deathtouch until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{B}"),
			Effect:  effects.ThisGainsKeywordUntilEndOfTurn("deathtouch", "Twilight Panther — deathtouch until end of turn"),
		}},
	})
}
