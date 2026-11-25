package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Enduring Ideal — Sorcery {5}{W}{W} (Saviors of Kamigawa; OLD b004, #6141):
//
//	"Search your library for an enchantment card, put it onto the
//	 battlefield, then shuffle.
//	 Epic (For the rest of the game, you can't cast spells. At the
//	 beginning of each of your upkeeps, copy this spell except for its
//	 epic ability.)"
//
// The search may find nothing (CR 701.23b); the library is shuffled
// either way. Epic is effects.Epic (CR 702.50a): the caster's cast ban for
// the rest of the game, and a copy of the spell at each of their upkeeps,
// which searches again and has no epic of its own.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d2045055-197c-436a-9874-199db5fe1c96",
		Name:         "Enduring Ideal",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		OnResolve: effects.Epic(func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.SearchLibrary{
				Player:    ctx.Controller(),
				Predicate: func(c game.Card) bool { return c.IsEnchantment() },
				Dest:      game.ZoneBattlefield,
				Limit:     1,
				Shuffle:   true,
				Reason:    "Enduring Ideal — an enchantment card, onto the battlefield",
				Source:    ctx.Source(),
			}.Apply(ctx)
		}),
	})
}
