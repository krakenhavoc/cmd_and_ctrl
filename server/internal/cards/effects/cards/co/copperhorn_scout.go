package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Copperhorn Scout — Creature — Elf Scout {G}, 1/1 (Scars of Mirrodin,
// #5577):
//
//	"Whenever this creature attacks, untap each other creature you
//	 control."
//
// An attack trigger whose body untaps every other tapped creature its
// controller controls as it resolves (somB1UntapEachOther). The Scout
// itself stays tapped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3d2ea1a3-72af-48bc-a8f1-98facc2eebd0",
		Name:         "Copperhorn Scout",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Copperhorn Scout — untap each other creature you control",
				effects.SomB1UntapEachOther(func(_ *effects.Context, c game.Card) bool { return c.IsCreature() })),
		},
	})
}
