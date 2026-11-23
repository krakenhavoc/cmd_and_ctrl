package cardsst

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sword of Hours — Artifact — Equipment {2} (Adventures in the Forgotten
// Realms Commander, S82):
//
//	"Whenever equipped creature attacks, put a +1/+1 counter on it.
//	 Whenever equipped creature deals combat damage, roll a d12. If the
//	 result is greater than the damage dealt or the result is 12, double
//	 the number of +1/+1 counters on that creature.
//	 Equip {2}"
//
// Both triggers act on the creature that attacked or dealt the damage, read
// off the trigger's event, so the Sword coming off it in response does not
// change who is counted. The combat-damage ability is one trigger per damage
// step (WheneverEquippedCreatureDealsCombatDamage's batch rule, CR 510.2,
// 603.2c); "the damage dealt" is everything that creature dealt in that step,
// summed over the recipients, read as the ability triggers and carried to the
// tail with the creature. The counters are doubled on that creature only if it
// is still on the battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "abaddcce-9039-4b4e-baf8-fdab304b5c52",
		Name:         "Sword of Hours",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.AttachedCreatureAttacked(ev, source)
			}, "Sword of Hours — put a +1/+1 counter on equipped creature", swordOfHoursCounter),
			effects.OncePerBatch(effects.On(game.EventDealDamage, effects.B43AttachedCreatureDealtCombatDamage,
				"Sword of Hours — roll a d12, if the result is greater than the damage dealt or is 12, double the +1/+1 counters on that creature", swordOfHoursRoll)),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}

// swordOfHoursCounter puts the +1/+1 counter on the creature that attacked.
func swordOfHoursCounter(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	id := item.Trigger.Event.CardID
	if !effects.OnBattlefield(g, id) {
		return nil
	}
	return effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}.Apply(effects.NewContext(g, item))
}

// swordOfHoursRoll rolls the d12, handing the tail the creature and the
// combat damage it dealt in this step.
func swordOfHoursRoll(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	first := item.Trigger.Event
	dealt := 0
	for _, ev := range g.Events {
		if ev.Kind == game.EventDealDamage && ev.Combat && ev.Batch == first.Batch && ev.Source == first.Source && ev.Amount > 0 {
			dealt += ev.Amount
		}
	}
	return effects.RollDice{Sides: 12, N: 1, Then: swordOfHoursDouble, TailN: dealt, TailCarry: []uuid.UUID{first.Source}}.Apply(effects.NewContext(g, item))
}

// swordOfHoursDouble is the rest of the ability, after the d12. An on-disk key.
var swordOfHoursDouble = effects.CardTail("dice/sword-of-hours-double", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.Rolls) == 0 || len(o.Tail.Carry) == 0 {
		return nil
	}
	result := effects.DiceResult(o)
	if result != 12 && result <= o.Tail.N {
		return nil
	}
	id := o.Tail.Carry[0]
	return effects.DoublePlusOneCountersOnEach(ctx.Game, ctx.Item, func(c game.Card) bool { return c.InstanceID == id })
})
