package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lord Windgrace — Legendary Planeswalker — Windgrace {2}{B}{R}{G}, loyalty 5
// (Special Guests, S82 BNS):
//
//	"+2: Discard a card, then draw a card. If a land card is discarded this
//	 way, draw an additional card.
//	 −3: Return up to two target land cards from your graveyard to the
//	 battlefield.
//	 −11: Destroy up to six target nonland permanents, then create six 2/2
//	 green Cat Warrior creature tokens with forestwalk.
//	 Lord Windgrace can be your commander."
//
// The +2 discards by choice (a prompt over the hand), then draws; the card
// discarded is read back off the discard (a land discarded draws one more),
// and an empty hand discards nothing but still draws. The −3 and −11 each
// judge their targets as they resolve (CR 608.2b): a land or permanent that
// left in response is skipped, and the tokens still come. "Can be your
// commander" is deck-building, read from the oracle text by the deck validator.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "38a55562-1e2d-4240-9454-219a4a25d38d",
		Name:            "Lord Windgrace",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 5,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+2: Discard a card, then draw a card. If a land card is discarded this way, draw an additional card.",
				Cost:    effects.LoyaltyCost(2),
				Purpose: game.Purpose{Draws: 1, Discards: 1},
				Effect: func(g *game.Game, item *game.StackItem) error {
					return g.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
						Player:   item.Controller,
						Source:   item.SourceCardID,
						N:        1,
						Question: "Lord Windgrace — discard a card, then draw a card",
						Asked:    game.NewAskedItem(item),
					}, lordWindgraceDraw, nil)
				},
			},
			{
				Label:   "−3: Return up to two target land cards from your graveyard to the battlefield.",
				Cost:    effects.LoyaltyCost(-3),
				Targets: effects.TargetCardInGraveyard("up to two target land cards from your graveyard", effects.YouOwn(), effects.Land()).WithCount(0, 2),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						if err := (effects.ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneBattlefield}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Label:   "−11: Destroy up to six target nonland permanents, then create six 2/2 green Cat Warrior creature tokens with forestwalk.",
				Cost:    effects.LoyaltyCost(-11),
				Purpose: game.Purpose{Tokens: 6},
				Targets: effects.TargetPermanent("up to six target nonland permanents", effects.Nonland()).WithCount(0, 6),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						if err := (effects.DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
							return err
						}
					}
					return effects.CreateToken{Template: effects.TokenCard("2/2 green Cat Warrior with forestwalk"), N: 6}.Apply(ctx)
				},
			},
		},
	})
}

// lordWindgraceDraw is the clause after the discard: draw a card, and one
// more if a land card was discarded this way. An on-disk key.
var lordWindgraceDraw = game.RegisterDiscardThen("choose-cards/lord-windgrace-draw", func(g *game.Game, r game.DiscardsDone) error {
	if r.Item == nil {
		return nil
	}
	n := 1
	for _, id := range r.Discarded.Cards() {
		if c, ok := g.LookupCardForEffect(id); ok && c.IsLand() {
			n++
			break
		}
	}
	return effects.DrawCards{Player: r.Item.Controller, N: n}.Apply(effects.NewContext(g, r.Item))
})
