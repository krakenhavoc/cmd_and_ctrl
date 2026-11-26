package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Searing Light — Instant {W} (Oath of the Gatewatch, #5545):
//
//	"Destroy target attacking or blocking creature with power 2 or less."
//
// The target is an attacking or blocking creature whose power is 2 or less,
// checked as it is announced and again as the spell resolves (CR 608.2b), so a
// creature pumped past 2 in response is safe. It is destroyed (it may
// regenerate or be indestructible).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2bad07d2-6284-4f87-9e2e-2e9ec5d2dd57",
		Name:         "Searing Light",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets: effects.TargetCreature("target attacking or blocking creature with power 2 or less",
			effects.AttackingOrBlocking(), effects.PowerLE(2)),
		OnResolve: effects.DestroyTheTargetPermanent,
	})
}
