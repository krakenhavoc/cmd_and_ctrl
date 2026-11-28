package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Expunge — Instant {2}{B} (Urza's Saga, S82 OLD b722):
//
//	Destroy target nonartifact, nonblack creature. It can't be regenerated.
//	Cycling {2} ({2}, Discard this card: Draw a card.)
//
// Terror's destroy (regeneration ignored, CR 701.19c) with cycling.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4b1fee86-360d-41d8-9238-eff216376732",
		Name:         "Expunge",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target nonartifact, nonblack creature", effects.Not(effects.Artifact()), effects.NonBlack()),
		OnResolve:    effects.DestroyTheTargetPermanentNoRegen,
		Activated:    []effects.ActivatedAbility{effects.Cycling("{2}")},
	})
}
