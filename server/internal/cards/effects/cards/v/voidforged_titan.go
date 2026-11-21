package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Voidforged Titan — Artifact Creature — Robot Warrior {4}{B}, 5/4 (Edge
// of Eternities):
//
//	"Void — At the beginning of your end step, if a nonland permanent
//	 left the battlefield this turn or a spell was warped this turn, you
//	 draw a card and lose 1 life."
//
// Void is an intervening if (void.go): checked as the end step begins and
// again as the ability resolves. The draw and the life loss are two
// steps, so the loss waits for a prompt the draw asks (CR 608.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f7afe534-fbb0-4f8b-9e98-aa49d4005465",
		Name:         "Voidforged Titan",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStepIfVoid("Voidforged Titan — you draw a card and lose 1 life",
				effects.Do(effects.DrawCards{N: 1}, effects.YouLoseLife(1))),
		},
	})
}
