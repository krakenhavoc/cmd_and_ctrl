package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kastral, the Windcrested — Legendary Creature — Bird Scout {3}{W}{U}, 4/5 (Bloomburrow):
//
//	"Flying
//	 Whenever one or more Birds you control deal combat damage to a player,
//	 choose one —
//	 • You may put a Bird creature card from your hand or graveyard onto the
//	   battlefield with a finality counter on it.
//	 • Put a +1/+1 counter on each Bird you control.
//	 • Draw a card."
//
// One trigger per player the Birds connect with in a combat damage step (CR
// 603.2c, the batch collapses), a modal triggered ability whose mode is chosen
// as it goes on the stack. The first bullet is an optional pick over the Bird
// creature cards in your hand and graveyard (none is legal), entering with a
// finality counter put on as it enters (ZoneEntryOptions.EntersWithCounters), so
// it is exiled instead if it would die (CR 122.1h). The second counts the Birds
// you control as it resolves (any creature with the type counts, a changeling
// included); the third is a plain draw.
//
// No simplification.
func init() {
	t := WheneverOneOrMoreCreaturesYouControlDealCombatDamageToAPlayer(OfCreatureType("Bird"),
		"Kastral, the Windcrested — choose one: put a Bird onto the battlefield with a finality counter, a +1/+1 counter on each Bird, or draw a card",
		func(*game.Game, *game.StackItem) error { return nil })
	t.Modes = ChooseOne(
		ModeDoing("You may put a Bird creature card from your hand or graveyard onto the battlefield with a finality counter on it.", nil,
			func(_ *game.StackItem, ctx *Context, _ int) error {
				p := ctx.PlayerByID(ctx.Controller())
				if p == nil {
					return nil
				}
				bird := ctx.TextSubtype("Bird")
				var cards []uuid.UUID
				for _, pile := range []*game.Zone{p.Hand, p.Graveyard} {
					if pile == nil {
						continue
					}
					for _, c := range pile.Cards {
						if c.IsCreature() && !c.IsToken() && c.HasSubtype(bird) {
							cards = append(cards, c.InstanceID)
						}
					}
				}
				if len(cards) == 0 {
					return nil
				}
				ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
					Chooser:  ctx.Controller(),
					Source:   ctx.Source(),
					Question: "Kastral, the Windcrested — you may put a Bird creature card from your hand or graveyard onto the battlefield with a finality counter on it",
					Cards:    cards,
					Min:      0,
					Max:      1,
					ThenKey:  kastralPut,
					Asked:    game.NewAskedItem(ctx.Item),
				})
				return nil
			}),
		ModeDoing("Put a +1/+1 counter on each Bird you control.", nil,
			func(_ *game.StackItem, ctx *Context, _ int) error {
				bird := ctx.TextSubtype("Bird")
				ctx = ctx.asGroupMember()
				for _, c := range ctx.Game.BattlefieldCardsForEffect() {
					if c.Controller == ctx.Controller() && c.IsCreature() && c.HasSubtype(bird) {
						if err := (AddCounter{Target: c.InstanceID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
							return err
						}
					}
				}
				return nil
			}),
		ModeDoing("Draw a card.", nil,
			func(_ *game.StackItem, ctx *Context, _ int) error {
				return DrawCards{N: 1}.Apply(ctx)
			}),
	)
	Register(Spec{
		OracleID:        "ced5fe9a-4237-4c89-9dd6-b133b966ebcc",
		Name:            "Kastral, the Windcrested",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{t},
	})
}

// kastralPut is the Bird chosen: it enters from the zone it is in with a
// finality counter. The card must still be in the hand or graveyard. An
// on-disk key: never renamed, never reused.
var kastralPut = askedPick("choose-cards/kastral-put-bird", func(ctx *Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	z := ctx.Game.FindCardZoneForEffect(r.Picked[0])
	if z == nil || (z.Kind != game.ZoneHand && z.Kind != game.ZoneGraveyard) {
		return nil
	}
	return ctx.Game.PutOntoBattlefieldTogetherKeyedForEffect(
		[]game.BatchEntry{{CardID: r.Picked[0], From: z.Kind}},
		game.ZoneEntryOptions{Controller: ctx.Controller(), EntersWithCounters: map[string]int{game.CounterFinality: 1}},
		game.KeyedTail{})
})
