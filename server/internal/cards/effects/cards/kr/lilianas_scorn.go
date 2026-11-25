package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Liliana's Scorn — Sorcery {3}{B}{B} (Core Set 2021, #5511):
//
//	"Destroy target creature. You may search your library and/or graveyard
//	 for a card named Liliana, Death Mage, reveal it, and put it into your
//	 hand. If you search your library this way, shuffle."
//
// The search is the planeswalker-fetch cycle's keyed "you may" answer
// (zones asked first, shuffle only after a library search). It is its own
// sentence, so it runs even when the destroy did nothing (a regenerated or
// indestructible creature); a creature that is an illegal target on
// resolution fizzles the whole spell (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2359db1c-6029-4748-9f09-d57bb0ae92e9",
		Name:         "Liliana's Scorn",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Tutors: 1},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error { return effects.S82MidDestroyTheTarget(nil, ctx) }),
			effects.Instruction(effects.S82OldB105MayFetchNamed("Liliana's Scorn", "Liliana, Death Mage")),
		),
	})
}
