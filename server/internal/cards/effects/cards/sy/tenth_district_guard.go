package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tenth District Guard — Creature — Human Soldier {1}{W}, 2/2 (Guilds of Ravnica, #5525):
//
//	"When this creature enters, target creature gets +0/+1 until end of turn."
//
// Daybreak Charger's targeted enter trigger with a +0/+1 pump.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "83ce0859-cb88-45dd-8485-3317792235de",
		Name:         "Tenth District Guard",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Tenth District Guard — target creature gets +0/+1 until end of turn",
				effects.BoostTargetUntilEOT("Tenth District Guard — +0/+1", 0, 1)),
				effects.TargetCreature("target creature")),
		},
	})
}
