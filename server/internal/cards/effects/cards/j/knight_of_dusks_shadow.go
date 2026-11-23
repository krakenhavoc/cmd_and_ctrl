package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Knight of Dusk's Shadow — Creature — Human Knight {1}{B}, 2/2:
//
//	"Menace
//	 Your opponents can't gain life.
//	 {1}{B}: This creature gets +1/+1 until end of turn."
//
// Menace rides PrintedKeywords; "Your opponents can't gain life" is
// ADR 0107 §5's battlefield static (CR 119.7, #1880); the pump is the
// shared until-end-of-turn boost on the Knight itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "815ae30f-f455-4795-96eb-4bcc41d415e8",
		Name:            "Knight of Dusk's Shadow",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		CantGainLife:    effects.OpponentsCantGainLife(),
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{B}: This creature gets +1/+1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1, Toughness: 1}},
			Cost:    effects.ManaCost("{1}{B}"),
			Effect:  effects.ThisGetsUntilEndOfTurn(1, 1, "Knight of Dusk's Shadow — +1/+1 until end of turn"),
		}},
	})
}
