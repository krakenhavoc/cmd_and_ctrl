package cardsap

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Augury Adept — Creature — Kithkin Wizard {1}{W/U}{W/U}, 2/2 (Commander 2013,
// S82 OLD b386):
//
//	"Whenever this creature deals combat damage to a player, reveal the top
//	 card of your library and put that card into your hand. You gain life equal
//	 to its mana value."
//
// Dark Confidant's flip with a gain instead of a loss: the card is revealed,
// its mana value read, the card put into hand, then the life gained (a land or
// a card with no mana cost gains none). An empty library does nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "be5a65fd-0d06-4771-bee7-0e42cc9871da",
		Name:         "Augury Adept",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Augury Adept — reveal the top card of your library and put it into your hand; you gain life equal to its mana value",
				auguryAdeptFlip),
		},
	})
}

func auguryAdeptFlip(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	var revealed []uuid.UUID
	if err := (effects.RevealTopOfLibrary{
		Player:   item.Controller,
		N:        1,
		Reason:   "Augury Adept — reveal the top card of your library",
		Revealed: &revealed,
	}).Apply(ctx); err != nil {
		return err
	}
	if len(revealed) == 0 {
		return nil
	}
	life := 0
	if c, ok := g.LookupCardForEffect(revealed[0]); ok {
		life = c.ManaValue()
	}
	return effects.BounceToHand{Target: revealed[0], ThenTail: effects.WithN(ctx.ItemTail(auguryAdeptGain), life)}.Apply(ctx)
}

// auguryAdeptGain is "you gain life equal to its mana value", read before the
// card moved and paid only once it reached a hand (a window that cancelled or
// redirected the move gains nothing). An on-disk key.
var auguryAdeptGain = effects.CardTail("then/old-b386-augury-adept-gain", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 || o.Tail.N <= 0 {
		return nil
	}
	return effects.GainLife{Player: ctx.Controller(), Amount: o.Tail.N}.Apply(ctx)
})
