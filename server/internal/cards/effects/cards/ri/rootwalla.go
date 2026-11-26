package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rootwalla — Creature — Lizard {2}{G}, 2/2 (Magic 2014, S82 OLD batch 401, #5563):
//
//	"{1}{G}: This creature gets +2/+2 until end of turn. Activate only once
//	 each turn."
//
// Astute Terraformer's once-per-turn pump: "Activate only once each turn" gates
// the offer (CR 602.5b), and the +2/+2 is a layer-7 modification until cleanup.
//
// No simplification.
func init() {
	const label = "{1}{G}: This creature gets +2/+2 until end of turn. Activate only once each turn."
	effects.Register(effects.Spec{
		OracleID:     "4dda5f21-11d6-40ce-be24-7c9529272c25",
		Name:         "Rootwalla",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:     label,
			Purpose:   game.Purpose{Answers: game.AnswerPump},
			Cost:      effects.ManaCost("{1}{G}"),
			Condition: effects.OncePerTurnActivation(label),
			Effect:    effects.ThisGetsUntilEndOfTurn(2, 2, "Rootwalla — +2/+2 until end of turn"),
		}},
	})
}
