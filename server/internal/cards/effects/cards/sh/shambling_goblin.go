package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shambling Goblin — Creature — Zombie Goblin {B}, 1/1 (Jumpstart, #5510):
//
//	"When this creature dies, target creature an opponent controls gets
//	 -1/-1 until end of turn."
//
// Splatter Goblin's trigger. The target is chosen when the trigger goes on
// the stack; one that is gone or no longer an opponent's creature by
// resolution is illegal and the shrink does nothing (CR 608.2b).
//
// No simplification.
func init() {
	const label = "Shambling Goblin — target creature an opponent controls gets -1/-1 until end of turn"
	effects.Register(effects.Spec{
		OracleID:     "9d8c39f2-ddb5-4b79-861a-e1e25f447c29",
		Name:         "Shambling Goblin",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(
				effects.WhenThisDies(label, effects.S82OldB098TargetCreatureGetsUntilEOT(label, -1, -1)),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
				effects.ForTargets(effects.HarmUntilEOT(0, -1, -1))),
		},
	})
}
