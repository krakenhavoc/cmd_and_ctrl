package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aurora Champion — Creature — Elf Warrior {2}{W}, 3/2 (Battlebond, #5529):
//
//	"Whenever this creature attacks, if your team controls another
//	 Warrior, tap target creature."
//
// "Your team" is "you" (CR 102.4). The intervening "if" is the shared
// Battlebond Warrior check (bbdWheneverAttacksWithAnotherWarrior), made as
// the creature attacks and again as the trigger resolves. The target is any
// creature, chosen as the trigger goes on the stack, and tapped if it is
// still legal.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a4499cf9-04a0-45c5-9c73-19f8e7e95da6",
		Name:         "Aurora Champion",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.BbdWheneverAttacksWithAnotherWarrior("Aurora Champion — tap target creature", effects.TapEachLegalTargetBody),
				effects.TargetCreature("target creature")),
		},
	})
}
