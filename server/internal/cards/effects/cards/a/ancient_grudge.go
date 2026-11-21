package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ancient Grudge — Instant {1}{R}:
//
//	"Destroy target artifact.
//	 Flashback {G}"
//
// Faithless Looting's flashback declaration around a plain destroy.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "306593b0-6ea8-476f-a3f2-e17876c1bab4",
		Name:             "Ancient Grudge",
		Completeness:     effects.CompletenessFull,
		Purpose:          game.Purpose{Answers: game.AnswerRemove},
		Targets:          effects.TargetPermanent("target artifact", effects.Artifact()),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Flashback("{G}")},
		OnResolve:        effects.DestroyTheTargetPermanent,
	})
}
