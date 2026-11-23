package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Unstoppable Ogre — Artifact Creature — Ogre Warrior {2}{R}, 4/1
// (Kamigawa: Neon Dynasty, S82):
//
//	"When this creature enters, target creature can't block this turn."
//
// Ahn-Crop Crasher's can't-block-this-turn effect on an enter trigger.
//
// No simplification.
func init() {
	const label = "Unstoppable Ogre — target creature can't block this turn"
	effects.Register(effects.Spec{
		OracleID:     "1733777d-87fe-4975-b099-1ff803a721e8",
		Name:         "Unstoppable Ogre",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters(label, effects.TargetCreatureCantBlockThisTurn(label)),
				effects.TargetCreature("target creature")),
		},
	})
}
