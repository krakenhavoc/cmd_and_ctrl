package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Steelclad Serpent — Artifact Creature — Serpent {5}{U}, 4/5 (Shards of Alara,
// S82 OLD batch 517, #5588):
//
//	"This creature can't attack unless you control another artifact."
//
// Mouser Mark III's attack gate, shared as
// S82OldB517CantAttackUnlessAnotherArtifact: a layer 6 restriction on the
// Serpent itself while its controller controls no OTHER artifact (it is an
// artifact itself and does not count). Checked as attackers are declared
// (CR 508.1c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6410e2bb-5ae6-4190-8b4a-6b2d5c9dff8d",
		Name:         "Steelclad Serpent",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.S82OldB517CantAttackUnlessAnotherArtifact()},
	})
}
