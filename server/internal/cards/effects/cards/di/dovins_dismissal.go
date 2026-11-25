package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dovin's Dismissal — Instant {2}{W}{U} (Ravnica Allegiance):
//
//	"Put up to one target tapped creature on top of its owner's library. You may
//	 search your library and/or graveyard for a card named Dovin, Architect of Law,
//	 reveal it, and put it into your hand. If you search your library this way,
//	 shuffle."
//
// With no target (it is "up to one") or with the target gone, the search still
// happens: the two sentences are independent (CR 608.2b). The creature goes on top
// first, so a shuffle that follows a library search takes it with the rest.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6db5f1a7-4cd8-42e2-a5d8-1603206ad0cc",
		Name:         "Dovin's Dismissal",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("up to one target tapped creature", effects.TappedPermanent()).WithCount(0, 1),
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(ctx.LegalTargets()) > 0 {
				if err := effects.PutTargetOnTopOfOwnersLibrary(item, ctx); err != nil {
					return err
				}
			}
			return effects.S82OldB105MayFetchNamed("Dovin's Dismissal", "Dovin, Architect of Law")(ctx)
		},
	})
}
