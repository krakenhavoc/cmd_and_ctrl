package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Daredevil, Fearless Fighter — Legendary Creature — Human Hero {2}{R},
// 3/3 (Marvel Super Heroes Commander):
//
//	"Whenever a source you control deals damage to you, Daredevil deals
//	 that much damage to target opponent.
//	 Whenever Daredevil attacks, exile the top card of your library.
//	 Daredevil deals damage to you equal to that card's mana value. You
//	 may play it this turn."
//
// The first ability reads the damage event (Quest for Pure Flame's
// "a source you control" test) and the amount off it, one trigger per
// damage event, so two sources are two triggers. Its own self-damage from
// the second ability feeds it, as printed. The second ability is
// Daredevil, Man Without Fear's impulse exile with a rider that reads the
// card that really reached exile.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9436a736-9139-4d8a-9eb0-9dbf706da51c",
		Name:         "Daredevil, Fearless Fighter",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					if ev.Amount <= 0 || ev.Target != source.Controller {
						return false
					}
					controller, ok := b34DamageSourceController(ev, g)
					return ok && controller == source.Controller
				}, "Daredevil, Fearless Fighter — it deals that much damage to target opponent", msc6DaredevilReflect),
				TargetPlayer("target opponent", Opponent())),
			On(game.EventAttack, msc6ThisAttacks,
				"Daredevil, Fearless Fighter — exile the top card of your library; it deals damage to you equal to its mana value; you may play it this turn",
				func(g *game.Game, item *game.StackItem) error {
					return b12ImpulseExileForTurnThen(g, item, 1, msc6DaredevilExiled)
				}),
		},
	})
}

func msc6DaredevilReflect(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetPlayer {
			return DealDamage{Source: ctx.Source(), Target: t.ID, Amount: item.Trigger.Event.Amount}.Apply(ctx)
		}
	}
	return nil
}

var msc6DaredevilExiled = CardTail("msc6/daredevil-fearless-fighter-exiled", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	card, ok := ctx.Game.LookupCardForEffect(o.IDs[0])
	if !ok {
		return nil
	}
	mv, ok := ctx.Game.ManaValueForEffect(card)
	if !ok || mv <= 0 {
		return nil
	}
	return DamageController{N: mv}.Apply(ctx)
})
