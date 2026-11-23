package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Broken Wings — Instant {2}{G} (Aetherdrift, #4464):
//
//	"Destroy target artifact, enchantment, or creature with flying."
//
// Make Your Move's destroy shape with the creature clause narrowed to a
// creature with flying, read off the permanent as it is now (a creature
// that only gained flying until end of turn is a legal target). The
// target is re-checked at resolution (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5e316864-d55c-496f-8f46-773567896864",
		Name:         "Broken Wings",
		Completeness: effects.CompletenessFull,
		Targets: effects.TargetPermanent("target artifact, enchantment, or creature with flying",
			effects.Or(effects.Artifact(), effects.Enchantment(), effects.And(effects.Creature(), effects.HasKeyword("flying")))),
		OnResolve: effects.DestroyTheTargetPermanent,
	})
}
