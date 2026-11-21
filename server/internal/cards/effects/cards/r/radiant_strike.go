package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Radiant Strike — Instant {3}{W} (Edge of Eternities, #4461):
//
//	"Destroy target artifact or tapped creature. You gain 3 life."
//
// Aerial Assault's destroy-then-gain-life with Royal Assassin's tapped
// creature (tappedPermanent) or any artifact as the target. The life is
// gained even if the target survives (an indestructible one); a creature
// untapped in response is no longer a legal target, so the spell doesn't
// resolve and no life is gained (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "186efd85-83d4-40cc-b617-76f32dbe99ce",
		Name:         "Radiant Strike",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target artifact or tapped creature", effects.Or(effects.Artifact(), effects.And(effects.Creature(), effects.TappedPermanent()))),
		OnResolve:    effects.DestroyTargetThenGainLife(3),
	})
}
