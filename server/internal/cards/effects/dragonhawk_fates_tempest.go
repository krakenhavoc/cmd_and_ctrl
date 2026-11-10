package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dragonhawk, Fate's Tempest — Legendary Creature — Bird Dragon {3}{R}{R}, 5/5:
//
//	"Flying
//	 Whenever Dragonhawk enters or attacks, exile the top X cards of your
//	 library, where X is the number of creatures you control with power
//	 4 or greater. You may play those cards until your next end step. At
//	 the beginning of your next end step, Dragonhawk deals 2 damage to
//	 each opponent for each of those cards that are still exiled."
//
// X is counted once, as the ability resolves. The cards are played at the
// usual timing and cost (a land only in a main phase with the stack empty,
// as your land drop) until your next end step begins. The delayed trigger
// names exactly the cards this ability exiled, each with its CR 400.7
// epoch, so an earlier Dragonhawk ability's cards, or a card that was
// played and exiled again, are not counted; it deals its damage as
// Dragonhawk last existed if he has left.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f7f435bc-1716-4412-9505-a1f2f2d83d08",
		Name:            "Dragonhawk, Fate's Tempest",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrAttacks("Dragonhawk, Fate's Tempest — exile the top X cards of your library, where X is the number of creatures you control with power 4 or greater; you may play them until your next end step",
				func(g *game.Game, item *game.StackItem) error {
					x := cmdrB5CountYouControl(g, item.Controller, func(c game.Card) bool {
						return c.IsCreature() && c.CurrentPower() >= 4
					})
					if x <= 0 {
						return nil
					}
					perm := game.CastPermission{Duration: g.UntilYourNextEndStepDuration(item.Controller)}
					return g.ExileTopWithPermissionKeyedForEffect(item.Controller, item.Controller, x, perm,
						NewContext(g, item).ItemTail(dragonhawkScheduleDamage))
				}),
		},
	})
}

// dragonhawkScheduleDamage runs once the exile has settled, told the cards
// that landed: it schedules the end-step damage over exactly those objects.
// An on-disk key: never renamed, never reused.
var dragonhawkScheduleDamage = CardTail("dragonhawk-fates-tempest/schedule-damage", func(ctx *Context, o game.TailOutcome) error {
	var objs []game.ObjectRef
	for _, id := range o.IDs {
		if z := ctx.Game.FindCardZoneForEffect(id); z != nil && z.Kind == game.ZoneExile {
			objs = append(objs, game.ObjectRef{ID: id, Epoch: ctx.Game.ObjectEpochForEffect(id)})
		}
	}
	if len(objs) == 0 {
		return nil
	}
	return ScheduleDelayedTrigger{
		At:                 game.StepEnd,
		ControllerTurnOnly: true,
		Label:              "Dragonhawk, Fate's Tempest — 2 damage to each opponent for each of those cards still exiled",
		Body:               dragonhawkDamageBody,
		Params:             game.EffectParams{Objects: objs},
	}.Apply(ctx)
})

// dragonhawkDamageBody is the delayed half: 2 damage to each opponent for
// each listed object that is still in exile.
var dragonhawkDamageBody = game.DelayedBody("dragonhawk-fates-tempest/damage-per-exiled",
	func(g *game.Game, item *game.StackItem, p game.EffectParams) error {
		n := 0
		for _, ref := range p.Objects {
			if z := g.FindCardZoneForEffect(ref.ID); z != nil && z.Kind == game.ZoneExile && g.ObjectEpochForEffect(ref.ID) == ref.Epoch {
				n++
			}
		}
		if n == 0 {
			return nil
		}
		return damageToEachOpponent(g, item, 2*n)
	})
