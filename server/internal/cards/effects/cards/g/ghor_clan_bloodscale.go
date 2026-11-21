package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ghor-Clan Bloodscale — Creature — Lizard Warrior {3}{R}, 2/1
// (Guildpact, #5599):
//
//	"First strike
//	 {3}{G}: This creature gets +2/+2 until end of turn. Activate only
//	 once each turn."
//
// Basking Rootwalla's once-a-turn pump (OncePerTurnActivation keyed on
// the label) at {3}{G}, on a first striker.
//
// No simplification.
func init() {
	const label = "{3}{G}: This creature gets +2/+2 until end of turn. Activate only once each turn."
	effects.Register(effects.Spec{
		OracleID:        "e3367ed8-16a3-4b04-bc6a-e0a3c794e6d3",
		Name:            "Ghor-Clan Bloodscale",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Activated: []effects.ActivatedAbility{{
			Label:     label,
			Purpose:   game.Purpose{Answers: game.AnswerPump},
			Cost:      effects.ManaCost("{3}{G}"),
			Condition: effects.OncePerTurnActivation(label),
			Effect:    effects.ThisGetsUntilEndOfTurn(2, 2, "Ghor-Clan Bloodscale — +2/+2 until end of turn"),
		}},
	})
}
