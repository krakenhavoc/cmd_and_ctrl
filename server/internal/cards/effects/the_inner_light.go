package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Inner Light — Enchantment — Saga {3}{G}{W} (Star Trek, #4973):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I — Return target green or white creature card from your graveyard
//	     to the battlefield.
//	 II — Choose target creature. Whenever it deals combat damage to a
//	      player this turn, you draw that many cards.
//	 III — You gain life equal to the greatest power among creatures you
//	       control. Create a Treasure token."
//
// The lore counters and the sacrifice are the engine's Saga rules
// (CR 714). Chapter I targets a green or white creature card in its
// controller's own graveyard, its colours read in the graveyard. Chapter
// II targets any creature and binds CR 603.7b's repeating delayed trigger
// to that object (Captain Howler's watch): every time it deals combat
// damage to a player before the turn ends, the Saga's controller draws
// that many cards (a double striker draws twice). Chapter III reads the
// greatest power among its controller's creatures as it resolves, after
// which the Treasure is made.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0de3aad2-5f4c-4437-b0b3-4b867309528d",
		Name:         "The Inner Light",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTriggerTargeting(1, "The Inner Light — I: return target green or white creature card from your graveyard to the battlefield",
				TargetCardInGraveyard("target green or white creature card from your graveyard", YouOwn(), Creature(), Or(OfColor("G"), OfColor("W"))),
				func(g *game.Game, item *game.StackItem) error {
					return returnLegalGraveyardTargetsToBattlefield(NewContext(g, item))
				}),
			ChapterTriggerTargeting(2, "The Inner Light — II: whenever target creature deals combat damage to a player this turn, you draw that many cards",
				TargetCreature("target creature"), theInnerLightWatch),
			ChapterTrigger(3, "The Inner Light — III: gain life equal to the greatest power among creatures you control; create a Treasure",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if err := (GainLife{Amount: b42GreatestPowerControlledBy(g, item.Controller)}).Apply(ctx); err != nil {
						return err
					}
					return CreateToken{Template: TreasureToken(), N: 1}.Apply(ctx)
				}),
		},
	})
}

// theInnerLightDrawThatMany is "you draw that many cards": the damage
// the watched creature dealt in the event that fired the trigger.
var theInnerLightDrawThatMany = game.SimpleDelayedBody("draw/the-inner-light-that-many", func(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil || item.Trigger.Event.Amount <= 0 {
		return nil
	}
	return DrawCards{Player: item.Controller, N: item.Trigger.Event.Amount}.Apply(NewContext(g, item))
})

// theInnerLightWatch binds the delayed draw to chapter II's target.
func theInnerLightWatch(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ids := legalTargetCardIDs(ctx)
	if len(ids) == 0 {
		return nil
	}
	c, ok := g.LookupCardForEffect(ids[0])
	if !ok {
		return nil
	}
	return DelayedOnEvent{
		Label:      "The Inner Light — draw that many cards",
		On:         []game.EventKind{game.EventDealDamage},
		Condition:  thatCreatureHitAPlayerCondition,
		CondParams: game.EffectParams{Object: game.ObjectRef{ID: ids[0], Epoch: c.ObjectEpoch}},
		Body:       theInnerLightDrawThatMany,
		Repeats:    true,
	}.Apply(ctx)
}
