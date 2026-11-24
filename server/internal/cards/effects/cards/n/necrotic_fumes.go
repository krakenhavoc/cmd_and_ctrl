package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Necrotic Fumes — Sorcery — Lesson {1}{B}{B} (Strixhaven: School of Mages,
// #5502):
//
//	"As an additional cost to cast this spell, exile a creature you control.
//	 Exile target creature or planeswalker."
//
// Fear of Abduction's additional cost (ExileCreatureYouControlCost): the
// caster names the creature on cast, it is exiled with the spell on the stack
// (nothing dies, nothing is sacrificed), and the spell is uncastable with no
// creature to exile. The exile of the target is Baleful Mastery's, done only
// if the target is still legal (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "784108b9-dc85-4344-8c71-d527c4f850c0",
		Name:           "Necrotic Fumes",
		Completeness:   effects.CompletenessFull,
		Purpose:        game.Purpose{Answers: game.AnswerRemove},
		AdditionalCost: effects.ExileCreatureYouControlCost(),
		Targets:        effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.ExileTheFirstTargetIfLegal(ctx)
		},
	})
}
