package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nantuko Cultivator — Creature — Insect Druid {3}{G}, 2/2 (Outlaws of Thunder
// Junction Commander, #5368):
//
//	"When this creature enters, you may discard any number of land cards. Put
//	 that many +1/+1 counters on this creature and draw that many cards."
//
// One pick over the land cards in the hand (so a nonland card is never offered),
// from none to all of them. The discard is an effect's (CR 701.9a): madness and
// "whenever you discard" see each card. The counters and the draw are the
// discard's keyed tail and count the cards that really reached the graveyard (a
// replacement may send one elsewhere), so the count is what was discarded. A
// Cultivator that has left the battlefield gets no counters but the draw still
// happens. With no land card in the hand nothing is asked.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b8dab16c-681c-4abe-b164-fcfb5dc66678",
		Name:         "Nantuko Cultivator",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Nantuko Cultivator — you may discard any number of land cards, then put that many +1/+1 counters on this creature and draw that many cards",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					p := g.PlayerByIDForEffect(ctx.Controller())
					if p == nil || p.Hand == nil {
						return nil
					}
					var lands []uuid.UUID
					for _, c := range p.Hand.Cards {
						if c.IsLand() {
							lands = append(lands, c.InstanceID)
						}
					}
					if len(lands) == 0 {
						return nil
					}
					g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
						Chooser:  ctx.Controller(),
						Source:   ctx.Source(),
						Question: "Nantuko Cultivator — discard any number of land cards",
						Cards:    lands,
						Min:      0,
						Max:      len(lands),
						Zone:     game.ZoneHand,
						ThenKey:  nantukoCultivatorPicked,
					})
					return nil
				}),
		},
	})
}

// nantukoCultivatorPicked discards the chosen lands. An on-disk key.
var nantukoCultivatorPicked = game.RegisterChooseCardsThen("choose-cards/nantuko-cultivator-discard", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	return g.DiscardCardsKeyedForEffect(r.Chooser, r.Picked, r.Source,
		game.KeyedTail{Ref: nantukoCultivatorDiscarded, Player: r.Chooser, Source: r.Source})
})

// nantukoCultivatorDiscarded puts that many counters on the Cultivator and
// draws that many cards. An on-disk key.
var nantukoCultivatorDiscarded = CardTail("nantuko-cultivator/discarded", func(ctx *Context, o game.TailOutcome) error {
	n := len(o.IDs)
	if n == 0 {
		return nil
	}
	if err := (CounterOnThis{Kind: game.CounterPlusOne, N: n}).Apply(ctx); err != nil {
		return err
	}
	return DrawCards{N: n}.Apply(ctx)
})
