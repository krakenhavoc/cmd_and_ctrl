package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Introduction to Annihilation — Sorcery — Lesson {5} (Strixhaven: School of
// Mages, #5502):
//
//	"Exile target nonland permanent. Its controller draws a card."
//
// "Its controller" is whoever controlled the permanent as the spell began to
// resolve (CR 608.2h); that player draws once the exile has settled, whether
// the card reached exile or a replacement sent it elsewhere, and a target that
// is no longer legal makes the whole spell do nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "60e36f86-e284-468e-8073-07f19a3ed65e",
		Name:         "Introduction to Annihilation",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target nonland permanent", effects.Nonland()),
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			controller, ok := effects.ControllerOfTarget(ctx, t.ID)
			if !ok {
				return nil
			}
			tail := ctx.Tail(introductionToAnnihilationDraw)
			tail.Carry = []uuid.UUID{controller}
			return effects.ExileTarget{Target: t.ID, ThenTail: tail}.Apply(ctx)
		},
	})
}

// introductionToAnnihilationDraw is "its controller draws a card", once the
// exile is done. An on-disk key.
var introductionToAnnihilationDraw = effects.CardTail("old-b042/introduction-to-annihilation-draw", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.Tail.Carry) == 0 {
		return nil
	}
	return effects.DrawCards{Player: o.Tail.Carry[0], N: 1}.Apply(ctx)
})
