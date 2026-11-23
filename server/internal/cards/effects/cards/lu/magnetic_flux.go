package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Magnetic Flux — Instant {2}{U} (Darksteel, #5605):
//
//	"Artifact creatures you control gain flying until end of turn."
//
// Crash Through's team grant narrowed to artifact creatures. The set is
// locked in as the spell resolves (CR 611.2c): an artifact creature that
// arrives later doesn't gain flying.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d50f756d-9ab6-4597-86ae-80beb70879ac",
		Name:         "Magnetic Flux",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerCombatGrant},
		OnResolve: effects.ResolveSteps(
			effects.GrantKeywordUntilEOT{
				Match:    effects.And(effects.Artifact(), effects.Creature(), effects.YouControl()),
				Keywords: []string{"flying"},
				Label:    "Magnetic Flux — artifact creatures you control gain flying",
			},
		),
	})
}
