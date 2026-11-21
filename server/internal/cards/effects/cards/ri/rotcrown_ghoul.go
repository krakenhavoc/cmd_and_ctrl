package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rotcrown Ghoul — {4}{U} Creature — Zombie 3/3 (Avacyn Restored, #5570):
//
//	"When this creature dies, target player mills five cards."
//
// A dies trigger aimed at a player, Drownyard Amalgam's mill for
// five.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e9a420d4-a834-4a14-b874-1db96624d8ef",
		Name:         "Rotcrown Ghoul",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisDies("Rotcrown Ghoul — target player mills five cards", effects.TargetPlayerMillsEffect(5)), effects.TargetPlayer("target player")),
		},
	})
}
