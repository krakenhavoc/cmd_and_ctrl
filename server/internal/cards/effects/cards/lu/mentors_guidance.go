package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mentor's Guidance — Sorcery {2}{U} (Strixhaven: School of Mages, #5502):
//
//	"When you cast this spell, copy it if you control a planeswalker, Cleric,
//	 Druid, Shaman, Warlock, or Wizard.
//	 Scry 1, then draw a card."
//
// A cast trigger that resolves above the spell (Ancestral Communion's shape):
// the copy is made under the caster's control when the trigger resolves, and
// only if the caster then controls a planeswalker or a creature that is a
// Cleric, Druid, Shaman, Warlock or Wizard (the condition is part of the
// instruction, not an intervening "if", so it is checked as the trigger
// resolves, CR 608.2c). The copy is not cast, so it triggers no cast trigger
// and cannot copy itself (CR 707.10). The spell and each copy scry 1 and then
// draw: the draw waits for the scry prompt (Opt's registered tail).
//
// No simplification.
func init() {
	mentor := effects.Or(effects.Planeswalker(), effects.HasSubtype("Cleric"), effects.HasSubtype("Druid"), effects.HasSubtype("Shaman"), effects.HasSubtype("Warlock"), effects.HasSubtype("Wizard"))
	copyOnCast := effects.WhenYouCastThisSpell("Mentor's Guidance — copy this spell if you control a planeswalker, Cleric, Druid, Shaman, Warlock, or Wizard",
		func(g *game.Game, item *game.StackItem) error {
			ctx := effects.NewContext(g, item)
			if effects.CountYouControl(ctx, mentor) == 0 {
				return nil
			}
			return effects.CopySpell{StackID: item.SourceCardID}.Apply(ctx)
		})
	effects.Register(effects.Spec{
		OracleID:     "ba961e01-315a-4195-a6d5-694a97a70b5a",
		Name:         "Mentor's Guidance",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue, Scry: 1, Draws: 1},
		Triggered:    []game.TriggeredAbility{copyOnCast},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.Scry{
				Player:   ctx.Controller(),
				N:        1,
				ThenTail: game.KeyedTail{Ref: effects.ThenDrawACard, Player: ctx.Controller()},
			}.Apply(ctx)
		},
	})
}
