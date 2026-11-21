package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cindervines — Enchantment {R}{G} (Ravnica Remastered):
//
//	"Whenever an opponent casts a noncreature spell, this enchantment deals
//	 1 damage to that player.
//	 {1}, Sacrifice this enchantment: Destroy target artifact or
//	 enchantment. This enchantment deals 2 damage to that permanent's
//	 controller."
//
// The damage is dealt by Cindervines as it last existed (CR 608.2h), to the
// permanent's controller as read before the destruction (as Boseiju's
// channel does), and rides the destruction's continuation so a commander
// window settles first. An indestructible target still costs its controller
// the 2 damage: "that permanent's controller" is not gated on the destroy.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "92d95e12-fa51-4360-88e4-9078ccc5d8d3",
		Name:         "Cindervines",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Actor == source.Controller {
					return false
				}
				spell, ok := g.LookupCardForEffect(ev.CardID)
				return ok && !game.ThroughTextOf(source, spell).IsCreature()
			}, "Cindervines — 1 damage to that player", cindervinesPing),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}, Sacrifice this enchantment: Destroy target artifact or enchantment. This enchantment deals 2 damage to that permanent's controller.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    game.AbilityCost{SacrificeSelf: true, Mana: "{1}"},
			Targets: TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
			Effect:  cindervinesDestroy,
		}},
	})
}

func cindervinesPing(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	caster := item.Trigger.Event.Actor
	if p := g.PlayerByIDForEffect(caster); p == nil || p.Eliminated {
		return nil
	}
	return DealDamage{Source: item.SourceCardID, Target: caster, Amount: 1}.Apply(NewContext(g, item))
}

func cindervinesDestroy(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var victim, owner uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		victim = t.ID
		if c, ok := g.LookupCardForEffect(victim); ok {
			owner = c.Controller
		}
		break
	}
	if victim == uuid.Nil || owner == uuid.Nil {
		return nil
	}
	return g.DestroyPermanentsKeyedForEffect([]uuid.UUID{victim}, withCarry(ctx.ItemTail(cindervinesDamage), owner))
}

// cindervinesDamage is "This enchantment deals 2 damage to that permanent's
// controller": the controller is Carry[0].
var cindervinesDamage = CardTail("then/cindervines-damage", func(ctx *Context, o game.TailOutcome) error {
	owner := tailCarry0(o)
	if owner == uuid.Nil {
		return nil
	}
	if p := ctx.Game.PlayerByIDForEffect(owner); p == nil || p.Eliminated {
		return nil
	}
	return DealDamage{Source: ctx.Source(), Target: owner, Amount: 2}.Apply(ctx)
})
