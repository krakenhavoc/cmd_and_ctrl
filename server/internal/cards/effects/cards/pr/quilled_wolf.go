package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Quilled Wolf — Creature — Wolf {1}{G}, 2/2 (Shadows over Innistrad, #5543):
//
//	"{5}{G}: This creature gets +4/+4 until end of turn."
//
// Hematite Golem's self-pump at {5}{G} for +4/+4 (thisGetsUntilEndOfTurn).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "607f9177-fea0-464c-a15f-65842865546a",
		Name:         "Quilled Wolf",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{5}{G}: This creature gets +4/+4 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 4, Toughness: 4}},
			Cost:    effects.ManaCost("{5}{G}"),
			Effect:  effects.ThisGetsUntilEndOfTurn(4, 4, "Quilled Wolf — +4/+4 until end of turn"),
		}},
	})
}
