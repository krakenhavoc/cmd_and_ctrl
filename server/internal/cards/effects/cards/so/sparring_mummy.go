package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sparring Mummy — Creature — Zombie {3}{W}, 3/3 (Battlebond, #5529):
//
//	When this creature enters, untap target creature.
//
// The target is chosen as the trigger goes on the stack and re-checked on resolution.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "414c97a3-cf31-4145-a431-3b1ad4343947",
		Name:         "Sparring Mummy",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Sparring Mummy — untap target creature", effects.UntapEachLegalTarget),
				effects.TargetCreature("target creature")),
		},
	})
}
