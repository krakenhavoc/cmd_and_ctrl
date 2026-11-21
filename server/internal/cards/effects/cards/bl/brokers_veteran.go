package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Broker's Veteran — Creature — Human Soldier {1}{U}, 2/1 (Streets of
// New Capenna):
//
//	"When this creature dies, put a shield counter on target creature
//	 you control."
//
// The target is chosen as the trigger goes on the stack; with it gone
// by resolution nothing happens (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "feee1926-0b63-404f-9b01-df965de26b5f",
		Name:         "Brokers Veteran",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisDies("Brokers Veteran — put a shield counter on target creature you control",
				effects.Do(effects.Instruction(effects.ShieldCounterOnFirstTarget))),
				effects.TargetCreature("target creature you control", effects.YouControl())),
		},
	})
}
