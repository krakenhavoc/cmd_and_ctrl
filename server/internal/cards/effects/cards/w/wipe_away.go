package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wipe Away — Instant {1}{U}{U} (Time Spiral, #5596):
//
//	"Split second (As long as this spell is on the stack, players can't
//	 cast spells or activate abilities that aren't mana abilities.)
//	 Return target permanent to its owner's hand."
//
// Boomerang with split second (CR 702.61a): while it is on the stack
// nobody can cast a spell or activate an ability that isn't a mana
// ability, so the target's controller can't answer the bounce with an
// activation.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ad9b6737-22c4-45cc-976b-32743ed48462",
		Name:            "Wipe Away",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordSplitSecond},
		Targets:         effects.TargetPermanent("target permanent"),
		Purpose:         game.Purpose{Answers: game.AnswerRemove},
		OnResolve:       effects.BounceTheTarget,
	})
}
