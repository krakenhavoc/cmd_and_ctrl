package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Llanowar Scout — Creature — Elf Scout {1}{G}, 1/3 (Dominaria, #5530):
//
//	"{T}: You may put a land card from your hand onto the battlefield."
//
// Sakura-Tribe Scout's ability: a land card from your hand onto the
// battlefield, if you choose one.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3bdc9211-8eff-4b30-8d6f-d26f4087f77f",
		Name:         "Llanowar Scout",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: You may put a land card from your hand onto the battlefield.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.TapCost(),
			Effect:  effects.Do(effects.MayPutALandFromHand("Llanowar Scout")),
		}},
	})
}
