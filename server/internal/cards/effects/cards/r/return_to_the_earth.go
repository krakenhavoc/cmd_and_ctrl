package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Return to the Earth — Instant {3}{G} (Fate Reforged,
// #5552):
//
//	"Destroy target artifact, enchantment, or creature with flying."
//
// Broken Wings' text and body: the creature clause is a creature with flying
// as it is now, re-checked at resolution (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "722f2965-a1ec-4b97-a4bd-b99364f58b1e",
		Name:         "Return to the Earth",
		Completeness: effects.CompletenessFull,
		Targets: effects.TargetPermanent("target artifact, enchantment, or creature with flying",
			effects.Or(effects.Artifact(), effects.Enchantment(), effects.And(effects.Creature(), effects.HasKeyword("flying")))),
		OnResolve: effects.DestroyTheTargetPermanent,
	})
}
