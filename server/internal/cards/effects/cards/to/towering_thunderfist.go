package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Towering Thunderfist — Creature — Giant Soldier {4}{R}, 4/4 (Ravnica:
// Clue Edition, S82):
//
//	"{W}: This creature gains vigilance until end of turn."
//
// Battlefly Swarm's self-pump shape with vigilance; the {W} is paid from
// any mana source, whatever the Giant's own colour.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d68adf9b-57b1-4c14-97b1-c5906dc5d29b",
		Name:         "Towering Thunderfist",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}: This creature gains vigilance until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{W}"),
			Effect:  effects.ThisGainsKeywordUntilEndOfTurn("vigilance", "Towering Thunderfist — vigilance until end of turn"),
		}},
	})
}
