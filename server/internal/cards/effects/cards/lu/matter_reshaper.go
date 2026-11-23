package cardslu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Matter Reshaper — Creature — Eldrazi {2}{C}, 3/2 (Commander Masters):
//
//	"When this creature dies, reveal the top card of your library. You may
//	 put that card onto the battlefield if it's a permanent card with mana
//	 value 3 or less. Otherwise, put that card into your hand."
//
// The top card is revealed (a library with none reveals nothing and does
// nothing). A permanent card of mana value 3 or less is offered to the
// battlefield; saying no, or a card that does not qualify, puts it into the
// hand. The offer is a keyed yes/no that carries the revealed card.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "38538aff-b6f9-4c30-992e-7fd78da3c44a",
		Name:         "Matter Reshaper",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Matter Reshaper — reveal the top card of your library; you may put it onto the battlefield if it's a permanent card with mana value 3 or less, otherwise put it into your hand",
				matterReshaperReveal),
		},
	})
}

func matterReshaperReveal(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	var revealed []uuid.UUID
	if err := (effects.RevealTopOfLibrary{
		Player:   ctx.Controller(),
		N:        1,
		Reason:   "Matter Reshaper — reveal the top card of your library",
		Revealed: &revealed,
	}).Apply(ctx); err != nil {
		return err
	}
	if len(revealed) == 0 {
		return nil
	}
	c, ok := g.LookupCardForEffect(revealed[0])
	if !ok {
		return nil
	}
	if !c.IsPermanent() || c.ManaValue() > 3 {
		g.TakeFromLibraryToHandForEffect(ctx.Controller(), revealed)
		return nil
	}
	return effects.MayChoice{
		Question: "Matter Reshaper — put " + c.Name + " onto the battlefield? (Otherwise it goes into your hand.)",
		Then:     matterReshaperAnswer,
		Carry:    revealed,
	}.Apply(ctx)
}

// matterReshaperAnswer is the keyed answer: yes puts the revealed card onto
// the battlefield, no (or a refused entry) puts it into the hand. An on-disk
// key: never renamed, never reused.
var matterReshaperAnswer = effects.MayAnswer("confirm/gap-b98-matter-reshaper", func(ctx *effects.Context, r game.ConfirmAnswer) error {
	if len(r.Carry) == 0 {
		return nil
	}
	id := r.Carry[0]
	if r.Accepted {
		if _, err := ctx.Game.PutFromLibraryOntoBattlefieldForEffect(id, game.LibraryEntryOptions{Controller: ctx.Controller()}); err != nil {
			return err
		}
		return nil
	}
	ctx.Game.TakeFromLibraryToHandForEffect(ctx.Controller(), r.Carry)
	return nil
})
