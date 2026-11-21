package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Throne of Geth — Artifact {2} (Scars of Mirrodin, #5577):
//
//	"{T}, Sacrifice an artifact: Proliferate."
//
// Atog's "sacrifice an artifact" cost (the Throne itself may be the one)
// beside the tap, then Thrummingbird's proliferate, which asks its
// controller which permanents and players get a counter (CR 701.34a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5a5f39ab-21b4-415f-aa1f-d8fad420d372",
		Name:         "Throne of Geth",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Sacrifice an artifact: Proliferate.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.TapCost(), effects.SacrificeN(1, "an artifact", effects.Artifact())),
			Effect:  effects.Do(effects.Proliferate{}),
		}},
	})
}
