package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mouser Mark III — {1}{U/R} Artifact Creature — Robot 2/3 (Teenage Mutant
// Ninja Turtles, #4457):
//
//	"This creature can't attack unless you control another artifact."
//
// Chief Warg's Company's shape: a layer 6 static on the creature itself
// that sets CantAttack while its controller controls no OTHER artifact (it
// is an artifact itself and does not count). The other permanents' types
// are read after the layers, as Doctor Doom's check reads them, so an
// artifact made by an effect counts. Restrictions are checked as attackers
// are declared (CR 508.1c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ad0c6d46-a1ca-4a9f-93f0-4a4dc31868e6",
		Name:         "Mouser Mark III",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.S82OldB517CantAttackUnlessAnotherArtifact()},
	})
}
