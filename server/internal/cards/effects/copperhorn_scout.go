package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "3d2ea1a3-72af-48bc-a8f1-98facc2eebd0",
		Name:         "Copperhorn Scout",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Copperhorn Scout — untap each other creature you control",
				somB1UntapEachOther(func(_ *Context, c game.Card) bool { return c.IsCreature() })),
		},
	})
}
