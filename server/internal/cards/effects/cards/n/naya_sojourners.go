package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Naya Sojourners — Creature — Elf Shaman {2}{R}{G}{W}, 5/3 (Alara Reborn,
// #5586):
//
//	"When you cycle this card and when this creature dies, you may put a +1/+1
//	 counter on target creature.
//	 Cycling {2}{G}"
//
// One ability with two conditions (S82OldB501CycledOrDies); the creature is
// chosen as the trigger goes on the stack and judged again as it resolves
// (CR 608.2b); the counter is optional.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "703215df-dee2-4b82-9f32-256e1410543c",
		Name:         "Naya Sojourners",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.Cycling("{2}{G}")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.Optional(
					effects.S82OldB501CycledOrDies("Naya Sojourners — you may put a +1/+1 counter on target creature",
						effects.PlusOneCounterOnTheTarget),
					"Naya Sojourners — put a +1/+1 counter on the target creature?"),
				effects.TargetCreature("target creature")),
		},
	})
}
