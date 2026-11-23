package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shattered Wings — Sorcery {2}{G} (Edge of Eternities, #4461):
//
//	"Destroy target artifact, enchantment, or creature with flying.
//	 Surveil 1."
//
// Broken Wings' target with Hour of Defeat's destroy-then-surveil
// (destroyTargetThenSurveil): the surveil happens even if the target
// survives (an indestructible one), and a target gone by resolution
// leaves the spell with no legal target, so it doesn't resolve
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "709b9f76-0ea5-456c-b1db-2cf517370fde",
		Name:         "Shattered Wings",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets: effects.TargetPermanent("target artifact, enchantment, or creature with flying",
			effects.Or(effects.Artifact(), effects.Enchantment(), effects.And(effects.Creature(), effects.HasKeyword("flying")))),
		OnResolve: effects.DestroyTargetThenSurveil(1),
	})
}
