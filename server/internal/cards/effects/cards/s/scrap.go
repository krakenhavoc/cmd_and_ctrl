package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Scrap — Instant {2}{R} (Urza's Saga, S82 OLD b722):
//
//	Destroy target artifact.
//	Cycling {2} ({2}, Discard this card: Draw a card.)
//
// Shatter's destroy with cycling, the shared constructor (the cycle draws through the ordinary activation from hand).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0ad3b788-eb3d-4876-9c89-893429595e0e",
		Name:         "Scrap",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target artifact", effects.Artifact()),
		OnResolve:    effects.DestroyTheTargetPermanent,
		Activated:    []effects.ActivatedAbility{effects.Cycling("{2}")},
	})
}
