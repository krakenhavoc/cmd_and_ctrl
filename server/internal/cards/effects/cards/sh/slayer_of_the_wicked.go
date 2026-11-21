package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Slayer of the Wicked — Creature — Human Soldier {3}{W}, 3/2 (Innistrad
// Remastered):
//
//	"When this creature enters, you may destroy target Vampire, Werewolf,
//	 or Zombie."
//
// The target is chosen as the trigger goes on the stack and the "may" is
// asked on resolution. The creature types are read through the layered view.
func init() {
	enter := effects.WhenThisEnters("Slayer of the Wicked — you may destroy target Vampire, Werewolf, or Zombie",
		effects.DestroyFirstLegalTarget)
	enter.Targets = effects.TargetCreature("target Vampire, Werewolf, or Zombie", effects.AnySubtype("Vampire", "Werewolf", "Zombie"))
	effects.Register(effects.Spec{
		OracleID:     "d93299d4-3aaf-48fb-8227-26203963c8b9",
		Name:         "Slayer of the Wicked",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(enter, "Slayer of the Wicked — destroy target Vampire, Werewolf, or Zombie?"),
		},
	})
}
