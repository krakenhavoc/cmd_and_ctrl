package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Territorial Hammerskull — Creature — Dinosaur {2}{W}, 2/3 (Ixalan, #5532):
//
//	"Whenever this creature attacks, tap target creature an opponent
//	 controls."
//
// The target is chosen as the trigger goes on the stack (CR 603.3d), so
// with no opposing creature it is removed; a target gone or no longer
// legal by resolution is not tapped (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a690c1de-158a-4f33-a916-ed2a130f92eb",
		Name:         "Territorial Hammerskull",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WheneverThisAttacks("Territorial Hammerskull — tap target creature an opponent controls", effects.TapTheFirstLegalTarget),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}
