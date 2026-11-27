package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Faerie Trickery — Kindred Instant — Faerie {1}{U}{U} (Lorwyn, S82 OLD b557):
//
//	"Counter target non-Faerie spell. If that spell is countered this way,
//	 exile it instead of putting it into its owner's graveyard."
//
// Horribly Awry's counter-and-exile with a non-Faerie target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bbe370ba-412b-4052-89d9-2d0ae4928118",
		Name:         "Faerie Trickery",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetSpell("target non-Faerie spell", effects.Not(effects.HasSubtype("Faerie"))),
		OnResolve:    effects.CounterFirstTargetAndExile,
	})
}
