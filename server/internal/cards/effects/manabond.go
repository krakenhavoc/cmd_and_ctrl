package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Manabond — Enchantment {G}:
//
//	"At the beginning of your end step, you may reveal your hand and put
//	 all land cards from it onto the battlefield. If you do, discard your
//	 hand."
//
// One optional end-step trigger. "You may" is the trigger's own prompt
// (Optional); a "yes" reveals the hand (every seat becomes a knower), puts
// every land card in it onto the battlefield as ONE simultaneous entry
// (PutCardsFromHandOntoBattlefieldKeyedForEffect, so landfall and ETBs see
// them together and a replacement window is judged on the board as it stood),
// and then the keyed tail discards what is left in hand, "if you do" being
// the may. A hand with no land card still reveals and still discards (the
// instruction is the may, not the lands). The discard is an effect discard,
// so madness and "whenever you discard" see each card.
func init() {
	Register(Spec{
		OracleID:     "482006a0-bb2a-4533-891b-2195da5e0f01",
		Name:         "Manabond",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(AtYourEndStep("Manabond — reveal your hand, put all land cards from it onto the battlefield, then discard your hand", manabondResolve),
				"Reveal your hand and put all land cards from it onto the battlefield? If you do, discard your hand."),
		},
	})
}

// manabondDiscardHand discards whatever the controller still holds, after the
// lands have entered. An on-disk key.
var manabondDiscardHand = CardTail("then/manabond-discard-hand", func(ctx *Context, _ game.TailOutcome) error {
	return manabondDiscardAll(ctx)
})

func manabondDiscardAll(ctx *Context) error {
	p := ctx.Game.PlayerByIDForEffect(ctx.Controller())
	if p == nil {
		return nil
	}
	ids := make([]uuid.UUID, 0, len(p.Hand.Cards))
	for _, c := range p.Hand.Cards {
		ids = append(ids, c.InstanceID)
	}
	if len(ids) == 0 {
		return nil
	}
	return ctx.Game.DiscardCardsKeyedForEffect(p.ID, ids, ctx.Source(), game.KeyedTail{})
}

func manabondResolve(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	p := g.PlayerByIDForEffect(ctx.Controller())
	if p == nil {
		return nil
	}
	g.RevealHandForEffect(p.ID)
	var lands []uuid.UUID
	for _, c := range p.Hand.Cards {
		if c.IsLand() {
			lands = append(lands, c.InstanceID)
		}
	}
	if len(lands) == 0 {
		return manabondDiscardAll(ctx)
	}
	return g.PutCardsFromHandOntoBattlefieldKeyedForEffect(lands, game.HandEntryOptions{}, ctx.ItemTail(manabondDiscardHand))
}
