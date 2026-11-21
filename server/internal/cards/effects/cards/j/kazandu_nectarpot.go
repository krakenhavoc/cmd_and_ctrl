package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kazandu Nectarpot — Creature — Insect {1}{G}, 1/3:
//
//	"Landfall — Whenever a land you control enters, you gain 1 life."
//
// Eumidian Terrabotanist's landfall row.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "263b526c-8b81-44ee-a7c4-5c24bf9b42a4",
		Name:         "Kazandu Nectarpot",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.LandfallPaying("Kazandu Nectarpot — you gain 1 life (landfall)", game.LandfallPayoff{LifeGain: 1}, effects.Do(effects.GainLife{Amount: 1})),
		},
	})
}
