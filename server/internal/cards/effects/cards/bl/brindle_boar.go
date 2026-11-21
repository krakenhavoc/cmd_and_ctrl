package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Brindle Boar — Creature — Boar {2}{G}, 2/2 (Magic 2011, #5579):
//
//	"Sacrifice this creature: You gain 4 life."
//
// Dedicated Martyr's life row with no mana: the sacrifice is the whole
// cost (CR 602.2b), so the Boar is gone before the ability resolves and
// the life comes anyway. It answers nothing (ADR 0142).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1c4f6893-4e83-4756-b324-1de0d06ea2c4",
		Name:         "Brindle Boar",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice this creature: You gain 4 life.",
			Purpose: game.Purpose{Answers: game.AnswerValue, LifeGain: 4},
			Cost:    effects.SacrificeThis(),
			Effect:  effects.Do(effects.GainLife{Amount: 4}),
		}},
	})
}
