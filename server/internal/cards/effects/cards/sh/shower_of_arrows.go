package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shower of Arrows — Instant {2}{G} (The Lord of the Rings: Tales of
// Middle-earth, #5948):
//
//	"Destroy target artifact, enchantment, or creature with flying. Scry 1."
//
// The target clause is an artifact, an enchantment, or a creature that has
// flying, each read through the layers. The destroy and the scry are one
// resolution, the scry after any prompt the destroy asks (CR 608.2c); with the
// target gone as it resolves the spell is countered and does not scry
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "efcbca5a-743c-4b16-a7cf-b07335b31ced",
		Name:         "Shower of Arrows",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Scry: 1},
		Targets: effects.TargetPermanent("target artifact, enchantment, or creature with flying",
			effects.Or(effects.Artifact(), effects.Enchantment(), effects.And(effects.Creature(), effects.HasKeyword("flying")))),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(effects.KineticOozeDestroy),
			effects.Scry{N: 1},
		),
	})
}
