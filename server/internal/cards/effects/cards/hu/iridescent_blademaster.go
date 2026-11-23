package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Iridescent Blademaster — Creature — Elf Warrior 2/2 {1}{G} (March of the Machine, S82):
//
//	"{3}{G}: This creature gets +2/+2 until end of turn."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5e1fd76b-46c8-41b5-87b3-e078f7a47535",
		Name:         "Iridescent Blademaster",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{G}: This creature gets +2/+2 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 2, Toughness: 2}},
			Cost:    effects.ManaCost("{3}{G}"),
			Effect:  effects.ThisGetsUntilEndOfTurn(2, 2, "Iridescent Blademaster — +2/+2 until end of turn"),
		}},
	})
}
