package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Magus of the Abyss — Creature — Human Wizard {3}{B}, 4/3:
//
//	"At the beginning of each player's upkeep, destroy target
//	 nonartifact creature that player controls of their choice. It
//	 can't be regenerated."
//
// #1997: The Abyss's trigger on a creature (the_abyss.go has the long
// form). The player whose upkeep it is chooses, the controller targets.
// On its controller's own upkeep the Magus is a legal choice, and when
// it is their only nonartifact creature they must choose it, as the
// card does.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "66958c48-3ac7-4f41-aa3e-adacaa24ea4f",
		Name:         "Magus of the Abyss",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{theAbyssTrigger("Magus of the Abyss")},
	})
}
