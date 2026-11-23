package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Twinblade Slasher — Creature — Elf Warrior {G}, 1/1 (Eventide, #5590):
//
//	"Wither (This deals damage to creatures in the form of -1/-1 counters.)
//	 {1}{G}: This creature gets +2/+2 until end of turn. Activate only once
//	 each turn."
//
// Basking Rootwalla's once-each-turn pump on a wither body.
//
// No simplification.
func init() {
	const label = "{1}{G}: This creature gets +2/+2 until end of turn. Activate only once each turn."
	effects.Register(effects.Spec{
		OracleID:        "2375a9c6-28ef-4c8f-909c-0aa7e5fc085c",
		Name:            "Twinblade Slasher",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordWither},
		Activated: []effects.ActivatedAbility{{
			Label:     label,
			Purpose:   game.Purpose{Answers: game.AnswerPump},
			Cost:      effects.ManaCost("{1}{G}"),
			Condition: effects.OncePerTurnActivation(label),
			Effect:    effects.ThisGetsUntilEndOfTurn(2, 2, "Twinblade Slasher — +2/+2 until end of turn"),
		}},
	})
}
