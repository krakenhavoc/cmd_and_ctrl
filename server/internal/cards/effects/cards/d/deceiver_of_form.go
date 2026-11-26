package cardsd

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Deceiver of Form — Creature — Eldrazi {6}{C}, 8/8 (Oath of the Gatewatch,
// #5545):
//
//	"At the beginning of combat on your turn, reveal the top card of your
//	 library. If a creature card is revealed this way, you may have creatures
//	 you control other than this creature become copies of that card until end
//	 of turn. You may put that card on the bottom of your library."
//
// The reveal is public. For a creature card the controller is asked whether to
// make the copies: every OTHER creature they control becomes a copy of the
// revealed card (its copiable values, CR 707.2, read as the effect begins) until
// end of turn, a set locked as the answer is given (CR 611.2c). Then, for any
// revealed card, they are asked whether to put it on the bottom of their
// library; if not it stays on top. The Deceiver is never copied.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "18c58328-518f-4228-bb1b-2433f3fb21f2",
		Name:         "Deceiver of Form",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtBeginningOfYourCombat("Deceiver of Form — reveal the top card; creatures you control other than this one may become copies of it until end of turn; you may put it on the bottom", deceiverOfFormReveal),
		},
	})
}

func deceiverOfFormReveal(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	top := g.RevealTopOfLibraryForEffect(item.Controller, item.SourceCardID, 1, "Deceiver of Form — the top card of the library")
	if len(top) == 0 {
		return nil
	}
	if c, ok := g.LookupCardForEffect(top[0]); ok && c.IsCreature() {
		return effects.MayChoice{
			Question: "Deceiver of Form — have creatures you control other than Deceiver of Form become copies of " + c.Name + " until end of turn?",
			Then:     deceiverOfFormCopies,
			Carry:    top[:1],
		}.Apply(ctx)
	}
	return deceiverOfFormAskBottom(ctx, top[0])
}

// deceiverOfFormCopies is the "you may" over the copies; either way the
// bottom question follows. An on-disk key: never renamed, never reused.
var deceiverOfFormCopies = effects.MayAnswer("may/deceiver-of-form-copies", func(ctx *effects.Context, r game.ConfirmAnswer) error {
	if len(r.Carry) == 0 {
		return nil
	}
	card := r.Carry[0]
	if r.Accepted {
		var ids []uuid.UUID
		for _, c := range effects.MatchingBattlefield(ctx, effects.And(effects.Creature(), effects.YouControl(), effects.OtherThan(ctx.Source()))) {
			ids = append(ids, c.InstanceID)
		}
		if len(ids) > 0 {
			if err := (effects.BecomeCopy{
				Targets: ids,
				Of:      card,
				Label:   "Deceiver of Form — becomes a copy of the revealed card until end of turn",
			}).Apply(ctx); err != nil {
				return err
			}
		}
	}
	return deceiverOfFormAskBottom(ctx, card)
})

func deceiverOfFormAskBottom(ctx *effects.Context, card uuid.UUID) error {
	if z := ctx.Game.FindCardZoneForEffect(card); z == nil || z.Kind != game.ZoneLibrary {
		return nil
	}
	return effects.MayChoice{
		Question: "Deceiver of Form — put the revealed card on the bottom of your library?",
		Then:     deceiverOfFormBottom,
		Carry:    []uuid.UUID{card},
	}.Apply(ctx)
}

// deceiverOfFormBottom is "you may put that card on the bottom of your
// library". An on-disk key: never renamed, never reused.
var deceiverOfFormBottom = effects.MayYes("may/deceiver-of-form-bottom", func(ctx *effects.Context, r game.ConfirmAnswer) error {
	if len(r.Carry) == 0 {
		return nil
	}
	return ctx.Game.PutOnBottomInRandomOrderForEffect(r.Chooser, game.ZoneLibrary, r.Carry[:1])
})
