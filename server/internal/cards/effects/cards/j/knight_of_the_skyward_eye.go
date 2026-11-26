package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Knight of the Skyward Eye — Creature — Human Knight {1}{W}, 2/2 (Shards of
// Alara, S82 OLD batch 511, #5588):
//
//	"{3}{G}: This creature gets +3/+3 until end of turn. Activate only once
//	 each turn."
func init() {
	const label = "{3}{G}: This creature gets +3/+3 until end of turn. Activate only once each turn."
	effects.Register(effects.Spec{
		OracleID:     "633e91ea-84d4-42ca-a8eb-603db340ebe7",
		Name:         "Knight of the Skyward Eye",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:     label,
			Purpose:   game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 3, Toughness: 3}},
			Cost:      effects.ManaCost("{3}{G}"),
			Condition: effects.OncePerTurnActivation(label),
			Effect:    effects.ThisGetsUntilEndOfTurn(3, 3, "Knight of the Skyward Eye — +3/+3 until end of turn"),
		}},
	})
}
