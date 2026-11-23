package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hunting Pack — Instant {5}{G}{G} (Scourge):
//
//	"Create a 4/4 green Beast creature token.
//	 Storm (When you cast this spell, copy it for each spell cast before
//	 it this turn.)"
//
// Beast Attack's token, plus the storm keyword (CR 702.40a). Each copy is a
// separate resolution that creates one Beast, so a storm count of three is
// four Beasts made one at a time. The spell has no target, so no copy asks
// anything.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d3e1848a-c236-4352-a7be-b0b4699af968",
		Name:         "Hunting Pack",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1, Answers: game.AnswerMakesBlocker},
		Triggered:    []game.TriggeredAbility{effects.Storm()},
		OnResolve:    effects.ResolveSteps(effects.CreateToken{Template: effects.TokenCard("4/4 green Beast"), N: 1}),
	})
}
