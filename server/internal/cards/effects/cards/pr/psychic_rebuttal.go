package cardspr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Psychic Rebuttal — Instant {1}{U} (Magic Origins, #5549):
//
//	"Counter target instant or sorcery spell that targets you.
//	 Spell mastery — If there are two or more instant and/or sorcery cards in
//	 your graveyard, you may copy the spell countered this way. You may choose
//	 new targets for the copy."
//
// The target clause is checked at announcement and again on resolution. The
// copy is offered only when the spell really was countered (an uncounterable
// spell stays on the stack and is not "countered this way"), and is made from
// the countered spell's last-known information (CR 608.2h, 707.10), under the
// caster's control with a new-target choice (CR 707.10c). Spell mastery is
// read as the spell resolves, so Psychic Rebuttal does not count itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "08dd393f-12d8-4801-882c-6d76009b4f12",
		Name:         "Psychic Rebuttal",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets: effects.TargetSpell("target instant or sorcery spell that targets you",
			effects.Or(effects.Instant(), effects.Sorcery()), effects.OldB318SpellTargetsYou),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			legal := ctx.LegalTargets()
			if len(legal) == 0 {
				return nil
			}
			spell := legal[0].ID
			if err := (effects.CounterTarget{StackID: spell}).Apply(ctx); err != nil {
				return err
			}
			if ctx.Game.StackItemForEffect(spell) != nil || !effects.SpellMastery(ctx.Game, ctx.Controller()) {
				return nil
			}
			return effects.MayChoice{
				Question: "Psychic Rebuttal — copy the spell you countered?",
				Then:     psychicRebuttalCopy,
				Carry:    []uuid.UUID{spell},
			}.Apply(ctx)
		},
	})
}

var psychicRebuttalCopy = effects.MayYes("confirm/psychic-rebuttal-copy", func(ctx *effects.Context, r game.ConfirmAnswer) error {
	if len(r.Carry) == 0 {
		return nil
	}
	return effects.CopySpell{StackID: r.Carry[0], Controller: ctx.Controller(), ChooseNewTargets: true, FromLastKnown: true}.Apply(ctx)
})
