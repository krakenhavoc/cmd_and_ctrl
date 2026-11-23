package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kor Celebrant — Creature — Kor Cleric {2}{W}, 1/4 (Zendikar Rising,
// #5508):
//
//	"Whenever this creature or another creature you control enters, you
//	 gain 1 life."
//
// Bogwater Lumaret's trigger: one per creature entering under your
// control, the Celebrant itself included, read off the entering object's
// layered type, so an opponent's creature does not count.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3f8f0ce9-d2ce-45cc-9fbf-0e2819b07a3e",
		Name:         "Kor Celebrant",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.CreatureEnteredUnderYourControl,
				"Kor Celebrant — you gain 1 life", effects.Do(effects.GainLife{Amount: 1})),
		},
	})
}
