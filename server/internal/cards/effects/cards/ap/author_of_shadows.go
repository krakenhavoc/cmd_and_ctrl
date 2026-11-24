package cardsap

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Author of Shadows — Creature — Shade Warlock {4}{B}, 3/3 (Commander 2021,
// S82 OLD b058, #5504):
//
//	"When this creature enters, exile all opponents' graveyards. Choose a
//	 nonland card exiled this way. You may cast that card for as long as it
//	 remains exiled, and you may spend mana as though it were mana of any
//	 color to cast that spell."
//
// Crypt Incursion's keyed exile over every opponent's graveyard, with a tail
// told the cards that really reached exile (a commander sent to the command
// zone is not "exiled this way"). The nonland ones are the candidates: with
// one it is taken with no prompt, with several the controller picks one
// (choose_cards, not a target). The pick gets the controller's cast-only
// permission for as long as it stays exiled (CR 611.2a), spending mana as
// though it were of any color; a land is never offered (it is not cast).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8bad4e43-4375-4e7f-897e-d28be72308d7",
		Name:         "Author of Shadows",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Author of Shadows — exile all opponents' graveyards, then choose a nonland card exiled this way; you may cast it for as long as it remains exiled, spending mana as though it were mana of any color",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					var ids []uuid.UUID
					for _, opp := range ctx.Opponents() {
						ids = append(ids, effects.B42CardsInGraveyardOf(g, opp, nil)...)
					}
					if len(ids) == 0 {
						return nil
					}
					return g.ExileCardsKeyedForEffect(ids, ctx.Tail(authorOfShadowsChoose))
				}),
		},
	})
}

// authorOfShadowsChoose is "choose a nonland card exiled this way" over the
// cards that landed. An on-disk key.
var authorOfShadowsChoose = effects.CardTail("author-of-shadows/choose-nonland", func(ctx *effects.Context, o game.TailOutcome) error {
	var nonland []uuid.UUID
	for _, id := range o.IDs {
		if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneExile {
			continue
		}
		if c, ok := ctx.Game.LookupCardForEffect(id); ok && !c.IsLand() {
			nonland = append(nonland, id)
		}
	}
	switch len(nonland) {
	case 0:
		return nil
	case 1:
		return authorOfShadowsGrant(ctx.Game, ctx.Controller(), ctx.Source(), nonland[0])
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  ctx.Controller(),
		Source:   ctx.Source(),
		Question: "Author of Shadows — choose a nonland card exiled this way",
		Cards:    nonland,
		Min:      1,
		Max:      1,
		LiveZone: game.ZoneExile,
		ThenKey:  authorOfShadowsPicked,
		Asked:    game.NewAskedItem(ctx.Item),
	})
	return nil
})

// authorOfShadowsPicked grants the cast permission over the chosen card. An
// on-disk key.
var authorOfShadowsPicked = game.RegisterChooseCardsThen("choose-cards/author-of-shadows-picked", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) != 1 {
		return nil
	}
	return authorOfShadowsGrant(g, r.Chooser, r.Source, r.Picked[0])
})

// authorOfShadowsGrant is "you may cast that card for as long as it remains
// exiled, and you may spend mana as though it were mana of any color".
func authorOfShadowsGrant(g *game.Game, player, source, card uuid.UUID) error {
	if z := g.FindCardZoneForEffect(card); z == nil || z.Kind != game.ZoneExile {
		return nil
	}
	return g.ExileCardWithPermissionForEffect(card, game.CastPermission{
		Player:     player,
		CastOnly:   true,
		AnyColor:   true,
		Duration:   game.WhileInZoneDuration(),
		Source:     source,
		SourceName: "Author of Shadows",
		Label:      "You may cast that card for as long as it remains exiled, and you may spend mana as though it were mana of any color to cast that spell",
	})
}
