package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Healer of the Pride — Creature — Cat Cleric {3}{W}, 2/3 (Magic 2013, #5568):
//
//	"Whenever another creature you control enters, you gain 2 life."
//
// Dazzling Angel's trigger at two life: a creature entering under your
// control other than the Healer itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "69b9fa68-b409-4de1-9a40-5262386a0180",
		Name:         "Healer of the Pride",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.CreatureEnterPaying(effects.WheneverAnotherCreatureEntersUnderYourControl("Healer of the Pride — you gain 2 life", effects.Do(effects.GainLife{Amount: 2})),
				game.CreatureEnterPayoff{LifeGain: 2}),
		},
	})
}
