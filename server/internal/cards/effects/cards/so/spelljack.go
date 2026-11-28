package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spelljack — Instant {3}{U}{U}{U} (Judgment):
//
//	"Counter target spell. If that spell is countered this way, exile it
//	 instead of putting it into its owner's graveyard. You may play it without
//	 paying its mana cost for as long as it remains exiled. (If it has X in its
//	 mana cost, X is 0.)"
//
// Kheru Spellsnatcher's counter-and-exile, as a spell: a spell that can't be
// countered stays on the stack and is not exiled or offered, and the standing
// {0} permission is the caster's own (S82OldB627CounterExileAndMayPlay).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7687b2a7-816d-4416-979b-675e35e235fc",
		Name:         "Spelljack",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetSpell("target spell"),
		OnResolve:    effects.S82OldB627CounterExileAndMayPlay("Spelljack"),
	})
}
