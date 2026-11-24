package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Brokers Initiate — Creature — Cat Citizen {W}, 0/4 (Streets of New
// Capenna):
//
//	"{4}{G/U}: This creature has base power and toughness 5/5 until end
//	 of turn."
//
// The effect sets base power and toughness in layer 7b (Flexible
// Waterbender's shape), so counters and pumps still apply on top. It does
// nothing to a new object if this one has left.
//
// No simplification.
func init() {
	const label = "Brokers Initiate — base power and toughness 5/5 until end of turn"
	effects.Register(effects.Spec{
		OracleID:     "9bdf033e-87d4-4bc4-ba08-4b497cc1b559",
		Name:         "Brokers Initiate",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}{G/U}: This creature has base power and toughness 5/5 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{4}{G/U}"),
			Effect:  effects.S82OldB097ThisHasBasePTUntilEOT(label, 5, 5),
		}},
	})
}
