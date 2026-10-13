package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Captain Howler, Sea Scourge — Legendary Creature — Shark Pirate
// {2}{U}{R}, 5/4 (issue #1112):
//
//	"Ward—{2}, Pay 2 life.
//	 Whenever you discard one or more cards, target creature gets
//	 +2/+0 until end of turn for each card discarded this way.
//	 Whenever that creature deals combat damage to a player this
//	 turn, you draw a card."
//
// The second and third sentences are one trigger, complete since #3172
// slice cav-e:
//
//   - "ONE OR MORE … FOR EACH" is the batch total (ADR 0108 amendment
//     2026-10-13, game/batch_total.go): the engine emits one
//     EventDiscardCard per card, and `AtBatchEnd: ThatManyEvents()` holds
//     the trigger until the discard has finished, so it goes on the stack
//     ONCE — one target pick — with `item.Trigger.Event.Amount` the number
//     of cards discarded. The pump is +2/+0 times that.
//   - "WHENEVER THAT CREATURE DEALS COMBAT DAMAGE TO A PLAYER THIS TURN"
//     is CR 603.7b's repeating delayed trigger (#2169, DelayedOnEvent
//     with Repeats), scheduled as the pump resolves and bound to the
//     chosen creature's object. It fires for every hit until cleanup (a
//     double striker draws twice) and a second discard this turn
//     schedules a second, independent watch, as the card reads.
//
// Ward uses the single mana-and-life payment shared with Gisa (#2945,
// ADR 0161). All printed behavior is implemented.
func init() {
	Register(Spec{
		OracleID:     "46c99d67-8307-4af0-afd0-377714806614",
		Name:         "Captain Howler, Sea Scourge",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Ward(WardManaAndLife("{2}", 2), "Captain Howler, Sea Scourge — ward {2}, pay 2 life"),
			withAtBatchEnd(Targeting(On(game.EventDiscardCard, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return discardedByYou(ev, source)
			}, "Captain Howler, Sea Scourge — target creature gets +2/+0 until end of turn for each card discarded this way",
				captainHowlerPump), TargetCreature("target creature")), ThatManyEvents()),
		},
	})
}

var (
	// "Whenever <that creature> deals combat damage to a player": an
	// EventDealDamage, combat, from CondParams.Object, to a player.
	thatCreatureHitAPlayerCondition = game.DelayedCondition("combat/that-creature-hit-a-player", thatCreatureHitAPlayer)

	// "You draw a card."
	drawACardForTheControllerBody = game.SimpleDelayedBody("draw/controller-one", drawACardForTheController)
)

func thatCreatureHitAPlayer(ev game.Event, _ *game.DelayedTrigger, g *game.Game, p game.EffectParams) bool {
	if ev.Kind != game.EventDealDamage || !ev.Combat || ev.Amount <= 0 || p.Object.ID == uuid.Nil || ev.Source != p.Object.ID {
		return false
	}
	return g.PlayerByIDForEffect(ev.Target) != nil
}

func drawACardForTheController(g *game.Game, item *game.StackItem) error {
	return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
}

// captainHowlerPump is the resolution body: +2/+0 until end of turn for
// each card discarded in the batch, on the target chosen when this
// instance of the trigger went on the stack, and the delayed draw
// trigger bound to that creature.
//
// Caller holds g.mu.
func captainHowlerPump(g *game.Game, item *game.StackItem) error {
	if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
		return nil
	}
	ctx := NewContext(g, item)
	target := item.Targets[0].ID
	if err := (BoostUntilEOT{
		Target: target,
		Power:  2 * settledEventAmount(item),
		Label:  "Captain Howler, Sea Scourge — +2/+0 for each card discarded",
	}).Apply(ctx); err != nil {
		return err
	}
	c, ok := g.LookupCardForEffect(target)
	if !ok {
		return nil
	}
	return DelayedOnEvent{
		Label:      "Captain Howler, Sea Scourge — draw a card",
		On:         []game.EventKind{game.EventDealDamage},
		Condition:  thatCreatureHitAPlayerCondition,
		CondParams: game.EffectParams{Object: game.ObjectRef{ID: target, Epoch: c.ObjectEpoch}},
		Body:       drawACardForTheControllerBody,
		Repeats:    true,
	}.Apply(ctx)
}
