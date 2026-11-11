package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// King Narfi's Betrayal — Enchantment — Saga {1}{U}{B} (Lost Caverns of
// Ixalan Commander):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I — Each player mills four cards. Then you may exile a creature or
//	 planeswalker card from each graveyard.
//	 II, III — Until end of turn, you may cast spells from among cards
//	 exiled with this Saga, and you may spend mana as though it were mana of
//	 any color to cast those spells."
//
// Chapter I mills every player in turn order, and only once the last mill
// has settled (CR 608.2c: the cards are chosen from what is there after the
// mills) asks the controller, for each graveyard in turn order, for up to
// one creature or planeswalker card. Each chosen card is exiled LINKED to
// this Saga (CR 607.2a), as The Cloning of Shredder does. Chapters II and
// III grant, over the cards still exiled with this Saga, an until-end-of-turn
// cast permission spending mana as any color; "cast spells" is cast only, so
// a land card is never playable, and there is no limit on how many are cast.
// A card that has left exile is no longer one of them.
func init() {
	const name = "King Narfi's Betrayal"
	castExiled := func(n int) game.TriggeredAbility {
		return ChapterTrigger(n, SagaChapterLabel(name, n, "until end of turn, you may cast spells from among cards exiled with this Saga, spending mana as though it were mana of any color"),
			kingNarfisBetrayalGrant)
	}
	Register(Spec{
		OracleID:     "35c416c6-f522-4b4c-93db-75c4e01843d1",
		Name:         name,
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, SagaChapterLabel(name, 1, "each player mills four cards; then you may exile a creature or planeswalker card from each graveyard"),
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return millEachPlayerInTurn(ctx, apnapPlayers(g), constantMill(4), ctx.ItemTail(kingNarfisBetrayalAfterMill))
				}),
			castExiled(2),
			castExiled(3),
		},
	})
}

// kingNarfisBetrayalAfterMill is chapter I's "then you may exile a creature
// or planeswalker card from each graveyard", run once every mill is done:
// one optional pick per graveyard that has a candidate, in turn order. An
// on-disk key: never renamed, never reused.
var kingNarfisBetrayalAfterMill = CardTail("tail/king-narfis-betrayal-after-mill", func(ctx *Context, _ game.TailOutcome) error {
	for _, p := range apnapPlayers(ctx.Game) {
		cands := graveyardCardIDs(ctx, p, func(c game.Card) bool { return c.IsCreature() || c.IsPlaneswalker() })
		if len(cands) == 0 {
			continue
		}
		ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:    ctx.Controller(),
			FromPlayer: p,
			Source:     ctx.Source(),
			Question:   "King Narfi's Betrayal — you may exile a creature or planeswalker card from this graveyard",
			Cards:      cands,
			Min:        0,
			Max:        1,
			LiveZone:   game.ZoneGraveyard,
			ThenKey:    kingNarfisBetrayalExile,
			Asked:      game.NewAskedItem(ctx.Item),
		})
	}
	return nil
})

// kingNarfisBetrayalExile is one graveyard's answer: the card chosen, if
// any, is exiled linked to the Saga. An on-disk key.
var kingNarfisBetrayalExile = itemChoice("choose-cards/king-narfis-betrayal-exile", func(ctx *Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	return ExileTarget{Target: r.Picked[0], Linked: true}.Apply(ctx)
})

// kingNarfisBetrayalGrant is chapters II and III: an until-end-of-turn cast
// permission over every card still exiled with this Saga.
func kingNarfisBetrayalGrant(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	src, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	var cards []game.Card
	for _, c := range g.ExiledWithForEffect(game.PermissionCardRef(src)) {
		cards = append(cards, *c)
	}
	if len(cards) == 0 {
		return nil
	}
	g.GrantCastPermissionToCardsForEffect(game.CastPermission{
		Player:     item.Controller,
		Zone:       game.ZoneExile,
		CastOnly:   true,
		AnyColor:   true,
		Duration:   g.UntilEndOfTurnDuration(),
		Source:     item.SourceCardID,
		SourceName: "King Narfi's Betrayal",
		Label:      "King Narfi's Betrayal — cast it this turn, spending mana as though it were mana of any color",
	}, cards)
	return nil
}
