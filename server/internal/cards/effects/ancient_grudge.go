package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ancient Grudge — Instant {1}{R}:
//
//	"Destroy target artifact.
//	 Flashback {G}"
//
// Faithless Looting's flashback declaration around a plain destroy.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "306593b0-6ea8-476f-a3f2-e17876c1bab4",
		Name:             "Ancient Grudge",
		Completeness:     CompletenessFull,
		Purpose:          game.Purpose{Answers: game.AnswerRemove},
		Targets:          TargetPermanent("target artifact", Artifact()),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{G}")},
		OnResolve:        destroyTheTargetPermanent,
	})
}
