package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ThisDealtCombatDamageToAnOpponent is "whenever this creature deals combat
// damage to an opponent" (Mindscour Dragon): positive combat damage from the
// source to a player who is not the source's controller. Damage to a
// planeswalker or a battle is not damage to an opponent.
func ThisDealtCombatDamageToAnOpponent(ev game.Event, source *game.Card, c game.Characteristic, g *game.Game) bool {
	return ThisDealtCombatDamageToAPlayer(ev, source, c, g) && ev.Target != source.Controller
}

// arcbondCondition and arcbondBody are Arcbond's delayed trigger, "Whenever that creature is
// dealt damage this turn, it deals that much damage to each other creature and
// each player" (CR 603.7b, 120.3): the damaged creature is the damage source
// and the amount is the damage it was just dealt.
var arcbondCondition = game.DelayedCondition("damage/the-chosen-creature-is-dealt-damage",
	func(ev game.Event, _ *game.DelayedTrigger, g *game.Game, p game.EffectParams) bool {
		if ev.Kind != game.EventDealDamage || ev.Amount <= 0 || len(p.Objects) == 0 || ev.Target != p.Objects[0].ID {
			return false
		}
		_, ok := g.PermanentForEffect(p.Objects[0])
		return ok
	})

var arcbondBody = game.SimpleDelayedBody("arcbond/reflect-the-damage", func(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ev := ctx.Trigger().Event
	creature := ev.Target
	if _, ok := g.PermanentRefForEffect(creature); !ok || ev.Amount <= 0 {
		return nil
	}
	var victims []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsCreature() && c.InstanceID != creature {
			victims = append(victims, c.InstanceID)
		}
	}
	for _, p := range g.Seats {
		if p != nil && !p.Eliminated {
			victims = append(victims, p.ID)
		}
	}
	return g.DamageInstanceForEffect(func() error {
		for _, v := range victims {
			if err := (DealDamage{Source: creature, Target: v, Amount: ev.Amount}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	})
})

// ArcbondChooseTarget schedules Arcbond's repeating delayed trigger about the
// target creature as the spell resolves (CR 603.7c: it follows that object).
func ArcbondChooseTarget(ctx *Context, id uuid.UUID) error {
	ref, ok := ctx.Game.PermanentRefForEffect(id)
	if !ok {
		return nil
	}
	ctx.Game.ScheduleDelayedTriggerForEffect(game.DelayedTrigger{
		Controller:   ctx.Controller(),
		SourceCardID: ctx.Source(),
		Label:        "Arcbond — the chosen creature deals that much damage to each other creature and each player",
		On:           []game.EventKind{game.EventDealDamage},
		Condition:    arcbondCondition,
		CondParams:   game.EffectParams{Objects: []game.ObjectRef{ref}},
		Body:         arcbondBody,
		Repeats:      true,
	})
	return nil
}
