package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Origin of Thor — {2}{R} Enchantment — Saga (Marvel Super Heroes
// Commander, #4455):
//
//	"I — You may discard a card. If you do, draw two cards.
//	 II — Whenever you cast a spell this turn, put a +1/+1 counter on target
//	      creature you control.
//	 III — Target creature you control deals damage equal to its power to
//	       each opponent."
//
// Chapter I is Abandon Attachments' optional rummage (the same keyed
// continuation). Chapter II is a repeating delayed trigger (CR 603.7b) with
// its own target clause, announced as each instance goes on the stack
// (Roads Go Ever, Ever On). Chapter III is damage from the creature, as
// its power is when the chapter resolves; a creature that has left deals
// nothing (CR 608.2b).
//
// No simplification.
func init() {
	const name = "Origin of Thor"
	Register(Spec{
		OracleID:     "d4e4614d-8247-4fc5-8883-e37204939aa4",
		Name:         name,
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, SagaChapterLabel(name, 1, "you may discard a card; if you do, draw two cards"),
				func(g *game.Game, item *game.StackItem) error {
					return g.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
						Player:   item.Controller,
						Source:   item.SourceCardID,
						N:        1,
						UpTo:     true,
						Question: "Origin of Thor — you may discard a card; if you do, draw two cards",
					}, abandonAttachmentsDraw, nil)
				}),
			ChapterTrigger(2, SagaChapterLabel(name, 2, "whenever you cast a spell this turn, put a +1/+1 counter on target creature you control"),
				func(g *game.Game, item *game.StackItem) error {
					return DelayedOnEvent{
						Label:     "Origin of Thor — put a +1/+1 counter on target creature you control",
						On:        []game.EventKind{game.EventCast},
						Condition: youNextCastCondition,
						Body:      originOfThorCounterBody,
						Repeats:   true,
					}.Apply(NewContext(g, item))
				}),
			ChapterTriggerTargeting(3, SagaChapterLabel(name, 3, "target creature you control deals damage equal to its power to each opponent"),
				TargetCreature("target creature you control", YouControl()), originOfThorDamage),
		},
	})
}

// originOfThorCounterBody is chapter II's delayed trigger; its target clause
// is announced as each instance goes on the stack. An on-disk key.
var originOfThorCounterBody = game.TargetedDelayedBody("msc-m5/origin-of-thor-counter", simpleBody(b36CounterOnChosenAnimal),
	constTargets(func() *game.TargetSpec { return TargetCreature("target creature you control", YouControl()) }))

func originOfThorDamage(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	c, ok := g.LookupCardForEffect(id)
	if !ok || c.CurrentPower() <= 0 {
		return nil
	}
	return originOfThorDealToEachOpponent(ctx, id, c.CurrentPower())
}

// originOfThorDealToEachOpponent is one damage instance from the creature
// to every opponent.
func originOfThorDealToEachOpponent(ctx *Context, source uuid.UUID, n int) error {
	return ctx.Game.DamageInstanceForEffect(func() error {
		for _, opp := range ctx.Opponents() {
			if err := ctx.Game.DealDamageToPlayerForEffect(source, opp, n); err != nil {
				return err
			}
		}
		return nil
	})
}
