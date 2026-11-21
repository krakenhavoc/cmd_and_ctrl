package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Terminal Agony — Sorcery {2}{B}{R} (Breaking News, #5929):
//
//	"Destroy target creature.
//	 Madness {B}{R} (If you discard this card, discard it into exile. When
//	 you do, cast it for its madness cost or put it into your graveyard.)"
//
// Murder at sorcery speed. Madness is the keyword's own (game/madness.go),
// the engine owns the discard trigger and the cast; the card is one
// target and one string, as Big Game Hunter and Fiery Temper are.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b69e68e7-0612-40f8-a40a-9c4e77653442",
		Name:         "Terminal Agony",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Madness:      "{B}{R}",
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.DestroyTheTargetPermanent,
	})
}
