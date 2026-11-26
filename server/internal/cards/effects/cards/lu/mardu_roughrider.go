package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mardu Roughrider — Creature — Orc Warrior {2}{R}{W}{B}, 5/4 (Khans of Tarkir):
//
//	"Whenever this creature attacks, target creature can't block this turn."
//
// Ahn-Crop Crasher's can't-block on a plain attack trigger: the target is
// chosen as the trigger goes on the stack and re-checked as it resolves
// (CR 608.2b), and it can't block until the end of the turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "10177654-c342-4537-a141-887a3fef13e8",
		Name:         "Mardu Roughrider",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WheneverThisAttacks("Mardu Roughrider — target creature can't block this turn",
					effects.TargetCreatureCantBlockThisTurn("Mardu Roughrider — it can't block this turn")),
				effects.TargetCreature("target creature")),
		},
	})
}
