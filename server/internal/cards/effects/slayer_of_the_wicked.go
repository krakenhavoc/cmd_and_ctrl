package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Slayer of the Wicked — Creature — Human Soldier {3}{W}, 3/2 (Innistrad
// Remastered):
//
//	"When this creature enters, you may destroy target Vampire, Werewolf,
//	 or Zombie."
//
// The target is chosen as the trigger goes on the stack and the "may" is
// asked on resolution. The creature types are read through the layered view.
func init() {
	enter := WhenThisEnters("Slayer of the Wicked — you may destroy target Vampire, Werewolf, or Zombie",
		destroyFirstLegalTarget)
	enter.Targets = TargetCreature("target Vampire, Werewolf, or Zombie", AnySubtype("Vampire", "Werewolf", "Zombie"))
	Register(Spec{
		OracleID:     "d93299d4-3aaf-48fb-8227-26203963c8b9",
		Name:         "Slayer of the Wicked",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(enter, "Slayer of the Wicked — destroy target Vampire, Werewolf, or Zombie?"),
		},
	})
}
