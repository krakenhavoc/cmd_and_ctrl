package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Frilled Oculus — Creature — Homunculus {1}{U}, 1/3 (Gatecrash, #5565):
//
//	"{1}{G}: This creature gets +2/+2 until end of turn. Activate only once
//	 each turn."
//
// Basking Rootwalla's ability: the once-a-turn limit is the shared
// activation condition, counted per object (CR 602.5b).
//
// No simplification.
func init() {
	const label = "{1}{G}: This creature gets +2/+2 until end of turn. Activate only once each turn."
	effects.Register(effects.Spec{
		OracleID:     "07ba4fb4-f95b-49ca-be6c-6cdda0f1f030",
		Name:         "Frilled Oculus",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:     label,
			Purpose:   game.Purpose{Answers: game.AnswerPump},
			Cost:      effects.ManaCost("{1}{G}"),
			Condition: effects.OncePerTurnActivation(label),
			Effect:    effects.ThisGetsUntilEndOfTurn(2, 2, "Frilled Oculus — +2/+2 until end of turn"),
		}},
	})
}
