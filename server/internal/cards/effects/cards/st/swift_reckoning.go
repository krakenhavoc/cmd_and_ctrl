package cardsst

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Swift Reckoning — Sorcery {1}{W} (Magic Origins, #5549):
//
//	"Spell mastery — If there are two or more instant and/or sorcery cards in
//	 your graveyard, you may cast this spell as though it had flash.
//	 Destroy target tapped creature."
//
// The timing permission is read as the cast begins (CR 601.3), when the
// spell is still in hand and so cannot count itself. The target is tapped
// by the announcement and by resolution (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "82b0c006-bf58-4f05-8801-e5d2857a6fd6",
		Name:         "Swift Reckoning",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target tapped creature", effects.TappedPermanent()),
		FlashIf: effects.FlashIfYou("if there are two or more instant and/or sorcery cards in your graveyard",
			func(g *game.Game, you uuid.UUID) bool { return effects.SpellMastery(g, you) }),
		OnResolve: effects.DestroyEachLegalTarget,
	})
}
