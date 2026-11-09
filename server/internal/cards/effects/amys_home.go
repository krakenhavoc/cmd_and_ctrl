package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Amy's Home — Plane — Earth (Doctor Who, #4870):
//
//	"When you planeswalk to Amy's Home and at the beginning of your
//	 upkeep, you may exile a nonland card from your hand with a number of
//	 time counters on it equal to its mana value. If it doesn't have
//	 suspend, it gains suspend.
//	 Whenever chaos ensues, time travel. (For each suspended card you own
//	 and each permanent you control with a time counter on it, you may
//	 add or remove a time counter.)"
//
// A plane (ADR 0216 §6), controlled by the planar controller (CR 901.6).
// The first ability is one ability with two trigger conditions: the
// planar controller may choose a nonland card from their hand (none is
// the "you may"), it is exiled, and the exile's keyed tail puts the time
// counters on it and gives suspend to a card without it
// (game.GainSuspendExileTail, CR 702.62a), so a card that never reached
// exile gets nothing. A card with mana value 0 gets no counters, so it is
// not suspended (CR 702.62b) and there is nothing to count down. The chaos
// ability (CR 311.7) time travels (CR 701.56a, TimeTravel).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5b200e41-126d-4017-8e01-f50582b31af0",
		Name:         "Amy's Home",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenYouPlaneswalkToThisAndAtYourUpkeep("Amy's Home — you may exile a nonland card from your hand with time counters equal to its mana value; it gains suspend",
				func(g *game.Game, item *game.StackItem) error {
					p := g.PlayerByIDForEffect(item.Controller)
					if p == nil || p.Hand == nil {
						return nil
					}
					var cards []uuid.UUID
					for _, c := range p.Hand.Cards {
						if !c.IsLand() {
							cards = append(cards, c.InstanceID)
						}
					}
					if len(cards) == 0 {
						return nil
					}
					g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
						Chooser:  item.Controller,
						Source:   item.SourceCardID,
						Question: "Amy's Home — you may exile a nonland card from your hand with time counters equal to its mana value",
						Cards:    cards,
						Min:      0,
						Max:      1,
						LiveZone: game.ZoneHand,
						ThenKey:  amysHomeExile,
						Asked:    game.NewAskedItem(item),
					})
					return nil
				}),
			WheneverChaosEnsues("Amy's Home — time travel", Do(TimeTravel{})),
		},
	})
}

// amysHomeExile exiles the chosen card; the exile's tail counts it down.
// An on-disk key: never renamed, never reused.
var amysHomeExile = askedPick("choose-cards/amys-home-exile", func(ctx *Context, r game.CardsChosen) error {
	for _, id := range r.Picked {
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok {
			return nil
		}
		return ctx.Game.ExileCardKeyedForEffect(id, game.KeyedTail{
			Ref:    game.GainSuspendExileTail,
			Player: ctx.Controller(),
			N:      c.ManaValue(),
		})
	}
	return nil
})
