package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tezzeret's Betrayal — Sorcery {3}{U}{B} (Aether Revolt, S82 older sets, #5536):
//
//	"Destroy target creature. You may search your library and/or graveyard
//	 for a card named Tezzeret, Master of Metal, reveal it, and put it into
//	 your hand. If you search your library this way, shuffle."
//
// The search is Vraska's Stoneglare's optional keyed fetch by name
// (S82OldB105MayFetchNamed); a target that is gone fizzles the whole spell
// (CR 608.2b), so there is no search then.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a660e7a0-0750-475d-a4bd-de5b37d53e97",
		Name:         "Tezzeret's Betrayal",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1, Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				ids := effects.LegalTargetCardIDs(ctx)
				if len(ids) == 0 {
					return nil
				}
				return ctx.Game.DestroyPermanentsKeyedForEffect(ids[:1], game.KeyedTail{})
			}),
			effects.Instruction(effects.S82OldB105MayFetchNamed("Tezzeret's Betrayal", "Tezzeret, Master of Metal")),
		),
	})
}
