package cardsn

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nael, Avizoa Aeronaut — Legendary Creature — Elf Scout {2}{G}{U}, 2/4
// (Dominaria United):
//
//	"Flying
//	 Domain — Whenever Nael deals combat damage to a player, look at the
//	 top X cards of your library, where X is the number of basic land
//	 types among lands you control. Put up to one of them on top of your
//	 library and the rest on the bottom of your library in a random
//	 order. Then if there are five basic land types among lands you
//	 control, draw a card."
//
// X is the shared domain count (DomainCount, #5666), read as the ability
// resolves. The look and the arrangement are Thassa's Oracle's: a private
// pick of up to one card to keep on top, every other card to the bottom in
// a random order. The "then if" is asked after the arrangement, against
// the lands as they are then (CR 608.2c), and the draw comes after the
// cards are placed, so it takes the kept card. With nothing to look at
// (no lands, or an empty library) the count is still read for the draw.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6044696a-0f64-474a-840e-8ed4dbd30221",
		Name:            "Nael, Avizoa Aeronaut",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Nael, Avizoa Aeronaut — look at the top X cards of your library (X is the number of basic land types among lands you control), put up to one of them on top and the rest on the bottom in a random order, then if there are five basic land types, draw a card",
				naelLookAndArrange),
		},
	})
}

func naelLookAndArrange(g *game.Game, item *game.StackItem) error {
	me := item.Controller
	ctx := effects.NewContext(g, item)
	looked := g.LookAtTopOfPlayersLibraryForEffect(me, me, effects.DomainCount(g, me))
	if len(looked) == 0 {
		return naelThenDraw(ctx)
	}
	return effects.RevealPick{
		Player:   me,
		Owner:    me,
		Question: "Nael, Avizoa Aeronaut — put up to one of these on top of your library; the rest go to the bottom in a random order",
		Cards:    looked,
		Min:      0,
		Max:      1,
		Keyed:    naelArranged,
	}.Apply(ctx)
}

// naelThenDraw is "Then if there are five basic land types among lands
// you control, draw a card".
func naelThenDraw(ctx *effects.Context) error {
	if effects.DomainCount(ctx.Game, ctx.Controller()) < len(game.BasicLandTypes) {
		return nil
	}
	return ctx.Game.DrawNForEffect(ctx.Controller(), 1)
}

// naelArranged puts the cards not kept on the bottom in a random order,
// then draws if the controller has all five basic land types. An on-disk
// key.
var naelArranged = effects.RegisterRevealPick("choose-cards/nael-avizoa-aeronaut-arranged", func(ctx *effects.Context, _, left []uuid.UUID, _ []int) error {
	if err := ctx.Game.PutOnBottomInRandomOrderForEffect(ctx.Controller(), game.ZoneLibrary, left); err != nil {
		return err
	}
	return naelThenDraw(ctx)
})
