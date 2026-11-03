package effects

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Conspiracy Theorist — Creature — Human Shaman {1}{R}, 2/2:
//
//	"Whenever this creature attacks, you may pay {1} and discard a card.
//	 If you do, draw a card.
//	 Whenever you discard one or more nonland cards, you may exile one of
//	 them from your graveyard. If you do, you may cast it this turn."
//
// THE LOOT is Banon's: a MayPay of {1} whose payment leads into a discard
// prompt with a floor of one, and the draw is the discard's continuation,
// run only for a player who really discarded ("if you do"). With an empty
// hand the cost can't be paid in full, so nothing is offered.
//
// THE DISCARD TRIGGER is one per batch of discards (Doctor Doom's
// AtBatchEnd shape), counting only nonland cards, so discarding three
// nonland cards together triggers once. "One of them" is read as the
// trigger resolves from the batch that fired it: the nonland cards you
// discarded in that batch that are still in your graveyard as the same
// object (one that left the graveyard since is gone, CR 400.7). The
// choice is optional (up to one). The chosen card is exiled, and only if
// it lands in exile does it get the permission: cast it this turn,
// paying its costs, at its normal timing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fa81e7ff-0f2c-40df-9169-c0a22f00cd76",
		Name:         "Conspiracy Theorist",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attackDeclared(ev, source)
			}, "Conspiracy Theorist — you may pay {1} and discard a card; if you do, draw a card",
				func(g *game.Game, item *game.StackItem) error {
					if b14HandSize(g, item.Controller) == 0 {
						return nil
					}
					return MayPay{
						Chooser:  item.Controller,
						Cost:     "{1}",
						Question: "Conspiracy Theorist — pay {1} and discard a card to draw a card?",
						Then:     conspiracyTheoristPaid,
					}.Apply(NewContext(g, item))
				}),
			withAtBatchEnd(On(game.EventDiscardCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return discardedByYou(ev, source) && conspiracyTheoristNonland(g, ev.CardID)
			}, "Conspiracy Theorist — you may exile one of the discarded nonland cards; you may cast it this turn",
				conspiracyTheoristExileOne), ThatManyEvents()),
		},
	})
}

// conspiracyTheoristPaid is the paid half of the attack trigger: the
// discard, keyed, with the draw as its continuation.
var conspiracyTheoristPaid = PayUnlessBranches("pay-unless/conspiracy-theorist-discard",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		return ctx.Game.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
			Player:   ctx.Controller(),
			Source:   ctx.Source(),
			N:        1,
			Question: "Conspiracy Theorist — discard a card, then draw a card",
		}, conspiracyTheoristDraw, nil)
	}, nil)

// conspiracyTheoristDraw is "if you do, draw a card", once the discard has
// landed.
var conspiracyTheoristDraw = game.RegisterDiscardThen("choose-cards/conspiracy-theorist-draw", func(g *game.Game, r game.DiscardsDone) error {
	if !r.Discarded.Discarded(r.Player) {
		return nil
	}
	return DrawCards{Player: r.Player, N: 1}.Apply(keyedContext(g, r.Player, r.Source))
})

// conspiracyTheoristNonland says whether the discarded card is a nonland
// card.
func conspiracyTheoristNonland(g *game.Game, id uuid.UUID) bool {
	c, ok := g.LookupCardForEffect(id)
	return ok && !c.IsLand()
}

// conspiracyTheoristExileOne offers the batch's nonland discards still in
// the controller's graveyard.
func conspiracyTheoristExileOne(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	you := item.Controller
	cards := conspiracyTheoristDiscardedInBatch(g, you, item.Trigger.Event)
	if len(cards) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    you,
		FromPlayer: you,
		Source:     item.SourceCardID,
		Question:   "Conspiracy Theorist — you may exile one of them from your graveyard; you may cast it this turn",
		Cards:      cards,
		Min:        0,
		Max:        1,
		LiveZone:   game.ZoneGraveyard,
		ThenKey:    conspiracyTheoristExiled,
	})
	return nil
}

// conspiracyTheoristDiscardedInBatch lists the nonland cards `you`
// discarded in the triggering event's batch that are in your graveyard
// now and have not left it since they arrived.
func conspiracyTheoristDiscardedInBatch(g *game.Game, you uuid.UUID, trigger game.Event) []uuid.UUID {
	events := g.EventsThisTurn()
	var out []uuid.UUID
	for i, ev := range events {
		if ev.Kind != game.EventDiscardCard || ev.Batch != trigger.Batch || ev.Actor != you || ev.CardID == uuid.Nil {
			continue
		}
		if trigger.Batch == 0 && ev.Seq != trigger.Seq {
			continue
		}
		if !conspiracyTheoristNonland(g, ev.CardID) || !conspiracyTheoristInGraveyardOf(g, you, ev.CardID) {
			continue
		}
		left := false
		for _, later := range events[i+1:] {
			if later.CardID == ev.CardID && later.Kind == game.EventZoneMove && later.OldZone == game.ZoneGraveyard {
				left = true
				break
			}
		}
		if !left && !slices.Contains(out, ev.CardID) {
			out = append(out, ev.CardID)
		}
	}
	return out
}

// conspiracyTheoristExiled is "exile it. If you do, you may cast it this
// turn" for the chosen card.
var conspiracyTheoristExiled = game.RegisterChooseCardsThen("choose-cards/conspiracy-theorist-exile", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	return g.ExileCardWithPermissionForEffect(r.Picked[0], game.CastPermission{
		Player:   r.Chooser,
		CastOnly: true,
		Label:    "Conspiracy Theorist — cast it this turn",
	})
})

// conspiracyTheoristInGraveyardOf says whether the card is in `you`'s
// graveyard now.
func conspiracyTheoristInGraveyardOf(g *game.Game, you, id uuid.UUID) bool {
	z := g.FindCardZoneForEffect(id)
	return z != nil && z.Kind == game.ZoneGraveyard && z.Owner == you
}
