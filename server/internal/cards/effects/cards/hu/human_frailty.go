package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Human Frailty — {B} Instant (Shadows over Innistrad via Duel Decks: Blessed
// vs. Cursed, S82 OLD b294):
//
//	"Destroy target Human creature."
//
// The Human bound is a target restriction judged at announce and again as the
// spell resolves (CR 608.2b), on the creature's effective subtypes.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bd95c832-9ae9-404a-a2bc-0760a5b9ff56",
		Name:         "Human Frailty",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target Human creature", effects.Subtype("Human")),
		OnResolve:    effects.DestroyEachLegalTarget,
	})
}
