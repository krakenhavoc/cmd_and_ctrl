package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sigardian Evangel — Creature — Human Cleric {1}{W}, 3/1 (Mystery
// Booster 2, #5948):
//
//	"When Sigardian Evangel enters the battlefield, conjure a card named
//	 Sigardian Evangel into your hand. Discard that card at the beginning
//	 of the next end step.
//	 When Sigardian Evangel enters the battlefield, tap target permanent
//	 you don't control."
//
// The conjured Evangel is a real card (ADR 0099 amendment 2026-11-24),
// so casting it before the end step conjures another. "That card" is the
// object conjured: the delayed trigger carries its ID and object epoch,
// and does nothing if it has left the hand (a card cast and returned is
// a new object, CR 400.7). Two printed triggers, two rows.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a5550795-8cd3-4f29-a9df-02c48b124dbc",
		Name:         "Sigardian Evangel",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Sigardian Evangel — conjure a Sigardian Evangel into your hand; discard it at the beginning of the next end step",
				effects.Do(effects.CheckConjure("Sigardian Evangel", effects.Conjure{Names: []string{"Sigardian Evangel"}, Zone: game.ZoneHand, ThenTail: sigardianEvangelDiscardLater}))),
			effects.Targeting(effects.WhenThisEnters("Sigardian Evangel — tap target permanent you don't control",
				effects.TapEachLegalTarget),
				effects.TargetPermanent("target permanent you don't control", effects.Not(effects.YouControl()))),
		},
	})
}

// sigardianEvangelDiscardLater schedules the discard of the card the
// conjure made. An on-disk key.
var sigardianEvangelDiscardLater = effects.CardTail("conjure/sigardian-evangel-discard-later", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	id := o.IDs[0]
	return effects.ScheduleDelayedTrigger{
		Label:  "Sigardian Evangel — discard the conjured Sigardian Evangel",
		Cards:  []uuid.UUID{id},
		Params: game.EffectParams{Amount: ctx.Game.ObjectEpochForEffect(id)},
		Body:   sigardianEvangelDiscard,
	}.Apply(ctx)
})

// sigardianEvangelDiscard is the delayed trigger's resolution: discard
// that card if it is still the same object in its owner's hand. An
// on-disk key.
var sigardianEvangelDiscard = game.SimpleDelayedBody("conjure/sigardian-evangel-discard", func(g *game.Game, item *game.StackItem) error {
	for _, t := range item.Targets {
		z := g.FindCardZoneForEffect(t.ID)
		if z == nil || z.Kind != game.ZoneHand || z.Owner != item.Controller || g.ObjectEpochForEffect(t.ID) != item.Params.Amount {
			continue
		}
		return g.DiscardCardsKeyedForEffect(item.Controller, []uuid.UUID{t.ID}, item.SourceCardID, game.KeyedTail{})
	}
	return nil
})
