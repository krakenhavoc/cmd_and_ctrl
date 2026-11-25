package cardsri

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ruin Processor — Creature — Eldrazi Processor {7}, 7/8 (Battle for
// Zendikar, #5547):
//
//	"When you cast this spell, you may put a card an opponent owns from
//	 exile into that player's graveyard. If you do, you gain 5 life."
//
// A cast trigger (CR 603.2), on the stack above the spell. The card is
// chosen as the trigger resolves from the cards in exile an opponent owns
// (never a token, CR 111.8), and goes to its owner's graveyard; the life is
// gained only if a card moved.
//
// No simplification.
func init() {
	const label = "Ruin Processor — you may put a card an opponent owns from exile into that player's graveyard; if you do, you gain 5 life"
	effects.Register(effects.Spec{
		OracleID:     "1cf82e44-99dd-48d3-a272-8df9835ebc0d",
		Name:         "Ruin Processor",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			FromStack: true,
			Watches:   []game.EventKind{game.EventCast},
			AppliesTo: effects.Self,
			Key:       label,
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, label)
				item.Controller, item.Owner = ev.Actor, ev.Actor
				return item
			},
			Effect: ruinProcessorChoose,
		}},
	})
}

func ruinProcessorChoose(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	me := ctx.Controller()
	var cands []uuid.UUID
	for _, c := range g.Exile.Cards {
		if c.Owner != me && !c.IsToken() {
			cands = append(cands, c.InstanceID)
		}
	}
	if len(cands) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  me,
		Source:   ctx.Source(),
		Question: "Ruin Processor — you may put a card an opponent owns from exile into that player's graveyard",
		Cards:    cands,
		Min:      0,
		Max:      1,
		LiveZone: game.ZoneExile,
		ThenKey:  ruinProcessorPick,
	})
	return nil
}

var ruinProcessorPick = game.RegisterChooseCardsThen("choose-cards/s82-old-b313-ruin-processor",
	func(g *game.Game, r game.CardsChosen) error {
		if !r.Answered || len(r.Picked) != 1 {
			return nil
		}
		if err := g.PutIntoGraveyardForEffect(r.Picked[0]); err != nil {
			return err
		}
		ctx, _ := effects.AskedContext(g, nil, r.Chooser, r.Source, nil)
		return effects.GainLife{Player: r.Chooser, Amount: 5}.Apply(ctx)
	})
