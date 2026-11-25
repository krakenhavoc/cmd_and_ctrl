package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Sea Devils — Enchantment — Saga {2}{G} (Doctor Who, #6100):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I, II — Create a 2/2 green Alien Salamander creature token with
//	 islandwalk.
//	 III — Until end of turn, whenever a Salamander deals combat damage to
//	 a player, it deals that much damage to target creature that player
//	 controls."
//
// Chapter III sets up a CR 603.7 delayed trigger that fires on every
// matching event until the turn ends (CR 603.7b), and outlives the Saga,
// which is sacrificed once the chapter has resolved (CR 714.4). It
// watches every Salamander, whoever controls it, judged by its creature
// types after the layers (so a changeling counts). Its target clause is
// chosen as each instance goes on the stack (CR 603.3d) by you, the
// delayed trigger's controller (CR 603.7d), among the creatures the player
// the Salamander damaged controls (ControlledByThatPlayer, read off the
// triggering event); an instance with nothing to target is removed. "It"
// deals the damage: the Salamander, by its last-known information if it
// has left (CR 608.2h), and "that much" is the combat damage it dealt.
//
// No simplification.
const seaDevilsName = "The Sea Devils"

var (
	// seaDevilsSalamanderHitAPlayer is the delayed trigger's event. An
	// on-disk key: never renamed, never reused.
	seaDevilsSalamanderHitAPlayer = game.DelayedCondition("combat-damage/a-salamander-to-a-player", seaDevilsSalamanderHit)

	// seaDevilsDamageBody is "it deals that much damage to target creature
	// that player controls". An on-disk key: never renamed, never reused.
	seaDevilsDamageBody = game.TargetedDelayedBody("the-sea-devils/salamander-deals-that-much-to-target-creature",
		effects.SimpleBody(seaDevilsDamage),
		effects.ConstTargets(func() *game.TargetSpec {
			return effects.ControlledByThatPlayer(effects.TargetCreature("target creature that player controls"))
		}))
)

func init() {
	salamander := func(n int) game.TriggeredAbility {
		return effects.ChapterTrigger(n, effects.SagaChapterLabel(seaDevilsName, n, "create a 2/2 green Alien Salamander creature token with islandwalk"),
			effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 green Alien Salamander with islandwalk"), N: 1}))
	}
	effects.Register(effects.Spec{
		OracleID:     "09d7a2d5-9e6e-4d84-8f97-c36b11ccc6fb",
		Name:         seaDevilsName,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			salamander(1),
			salamander(2),
			effects.ChapterTrigger(3, effects.SagaChapterLabel(seaDevilsName, 3, "until end of turn, whenever a Salamander deals combat damage to a player, it deals that much damage to target creature that player controls"),
				func(g *game.Game, item *game.StackItem) error {
					return effects.DelayedOnEvent{
						Label:     seaDevilsName + " — the Salamander deals that much damage to target creature that player controls",
						On:        []game.EventKind{game.EventDealDamage},
						Condition: seaDevilsSalamanderHitAPlayer,
						Body:      seaDevilsDamageBody,
						Repeats:   true,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// seaDevilsSalamanderHit is "a Salamander deals combat damage to a player".
func seaDevilsSalamanderHit(ev game.Event, dt *game.DelayedTrigger, g *game.Game, _ game.EffectParams) bool {
	if ev.Kind != game.EventDealDamage || !ev.Combat || ev.Amount <= 0 || g.PlayerByIDForEffect(ev.Target) == nil {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.Source)
	return ok && effects.OfCreatureType("Salamander")(g, dt.Controller, c)
}

// seaDevilsDamage is the fired instance: the Salamander deals the damage it
// dealt to the player to the chosen creature.
func seaDevilsDamage(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	ev := item.Trigger.Event
	ctx := effects.NewContext(g, item)
	victim := uuid.Nil
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			victim = t.ID
			break
		}
	}
	if victim == uuid.Nil || ev.Source == uuid.Nil || ev.Amount <= 0 {
		return nil
	}
	return effects.DealDamage{Source: ev.Source, Target: victim, Amount: ev.Amount}.Apply(ctx)
}
