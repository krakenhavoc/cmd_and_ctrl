package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tamiyo Meets the Story Circle — Enchantment — Saga {1}{U} (Modern Horizons
// 3):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I — Until your next turn, whenever a creature attacks you or a planeswalker
//	 you control, it gets -2/-0 until end of turn.
//	 II — Discard any number of cards, then investigate twice for each card
//	 discarded this way.
//	 III — Shuffle up to three target cards from your graveyard into your
//	 library."
//
// I is Jace, Reality Sculptor's -3 at -2/-0: a CR 603.7b delayed trigger with a
// stated duration, one per attacker, its condition "attacks you or a planeswalker
// you control" shared with Jace's (a registered condition). II is a discard of any
// number through the run's own prompt, the count of cards that really left the hand
// doubled into Clues (Investigate). III is Ghostly Castigator's shuffle of up to three
// targets from your graveyard, chosen as the chapter goes on the stack.
//
// No simplification.
func init() {
	shrink := effects.ChapterTrigger(1, "Tamiyo Meets the Story Circle — until your next turn, creatures that attack you or your planeswalkers get -2/-0 until end of turn",
		func(g *game.Game, item *game.StackItem) error {
			ctx := effects.NewContext(g, item)
			controller := ctx.Controller()
			until := effects.DurationUntilYourNextTurn(ctx, controller)
			g.ScheduleDelayedTriggerForEffect(game.DelayedTrigger{
				Controller:   controller,
				SourceCardID: ctx.Source(),
				Label:        "Tamiyo Meets the Story Circle — the attacking creature gets -2/-0 until end of turn",
				On:           []game.EventKind{game.EventAttack},
				Condition:    effects.JaceRealitySculptorAttacksYouCondition,
				Body:         tamiyoStoryCircleShrinkBody,
				Repeats:      true,
				Duration:     &until,
			})
			return nil
		})
	discard := effects.ChapterTrigger(2, "Tamiyo Meets the Story Circle — discard any number of cards, then investigate twice for each card discarded this way",
		func(g *game.Game, item *game.StackItem) error {
			hand := 0
			if p := g.PlayerByIDForEffect(item.Controller); p != nil && p.Hand != nil {
				hand = p.Hand.Size()
			}
			return g.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
				Player:   item.Controller,
				Source:   item.SourceCardID,
				N:        hand,
				UpTo:     true,
				Question: "Tamiyo Meets the Story Circle — discard any number of cards (investigate twice for each)",
				Asked:    game.NewAskedItem(item),
			}, effects.GapB56TamiyoInvestigate, nil)
		})
	shuffle := effects.ChapterTriggerTargeting(3, "Tamiyo Meets the Story Circle — shuffle up to three target cards from your graveyard into your library",
		effects.TargetCardInGraveyard("up to three target cards from your graveyard", effects.YouOwn()).WithCount(0, 3),
		effects.ShuffleTargetsIntoYourLibrary)
	effects.Register(effects.Spec{
		OracleID:     "d53d5d03-b180-4bdb-8801-260f8a75644d",
		Name:         "Tamiyo Meets the Story Circle",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{shrink, discard, shuffle},
	})
}

// tamiyoStoryCircleShrinkBody gives the attacking creature (the trigger's
// payload) -2/-0 until end of turn. An on-disk key.
var tamiyoStoryCircleShrinkBody = game.SimpleDelayedBody("tamiyo-meets-the-story-circle/attacker-gets-minus-two", func(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, id := range ctx.PayloadCards() {
		if _, ok := g.PermanentRefForEffect(id); !ok {
			continue
		}
		if err := (effects.BoostUntilEOT{Target: id, Power: -2, Label: "Tamiyo Meets the Story Circle — -2/-0"}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})
