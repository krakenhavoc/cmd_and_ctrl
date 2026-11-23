package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Desolation Prowler — Creature — Wolf {1}{B} (The Hobbit), 2/2:
//
//	"Pay 2 life: This creature gets +2/+2 until end of turn. Activate
//	 only once each turn."
//
// #4454, S76. Life is the cost (CR 119.4: refused below 2), and the
// once-a-turn limit counts activations (CR 602.1b), not resolutions.
//
// No simplification.
func init() {
	const label = "Pay 2 life: This creature gets +2/+2 until end of turn. Activate only once each turn."
	effects.Register(effects.Spec{
		OracleID:     "ed9d4bb1-7335-45fd-b2ea-97b1f6491d7a",
		Name:         "Desolation Prowler",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:     label,
			Purpose:   game.Purpose{Answers: game.AnswerPump},
			Cost:      effects.PayLife(2),
			Condition: effects.OncePerTurnActivation(label),
			Effect:    effects.ThisGetsUntilEndOfTurn(2, 2, "Desolation Prowler — +2/+2 until end of turn"),
		}},
	})
}
