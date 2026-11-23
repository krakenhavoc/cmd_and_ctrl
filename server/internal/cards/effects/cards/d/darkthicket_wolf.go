package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Darkthicket Wolf — Creature — Wolf {1}{G}, 2/2 (Innistrad, #5572):
//
//	"{2}{G}: This creature gets +2/+2 until end of turn. Activate only
//	 once each turn."
//
// Basking Rootwalla's once-a-turn pump at {2}{G}: OncePerTurnActivation
// refuses a second activation in the same turn (CR 602.5b).
//
// No simplification.
func init() {
	const label = "{2}{G}: This creature gets +2/+2 until end of turn. Activate only once each turn."
	effects.Register(effects.Spec{
		OracleID:     "a78ab15e-72e7-4d2b-bc42-b2960147ea7b",
		Name:         "Darkthicket Wolf",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:     label,
			Purpose:   game.Purpose{Answers: game.AnswerPump},
			Cost:      effects.ManaCost("{2}{G}"),
			Condition: effects.OncePerTurnActivation(label),
			Effect:    effects.ThisGetsUntilEndOfTurn(2, 2, "Darkthicket Wolf — +2/+2 until end of turn"),
		}},
	})
}
