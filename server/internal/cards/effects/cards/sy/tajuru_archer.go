package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tajuru Archer — Creature — Elf Archer Ally {2}{G}, 1/2 (Duel Decks:
// Zendikar vs. Eldrazi, #5548):
//
//	"Whenever this creature or another Ally you control enters, you may have
//	 this creature deal damage to target creature with flying equal to the
//	 number of Allies you control."
//
// Rally, one trigger per entering Ally. The target is chosen as the trigger
// goes on the stack; the Allies are counted as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5ef498fa-31d5-4eca-873d-6b2b1f7f9a3c",
		Name:         "Tajuru Archer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(effects.S82OldB316Rally("Tajuru Archer — this creature deals damage equal to the number of Allies you control to target creature with flying", effects.S82OldB316ArcherShot),
				effects.TargetCreature("target creature with flying", effects.HasKeyword("flying"))),
				"Tajuru Archer — have it deal damage to target creature with flying?"),
		},
	})
}
