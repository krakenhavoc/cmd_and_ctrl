package cardsu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Unnatural Hunger — Enchantment — Aura {3}{B}{B} (Mercadian Masques, S82 OLD batch 686, #5619):
//
//	"Enchant creature
//	At the beginning of the upkeep of enchanted creature's controller, this Aura deals damage equal to that creature's power to that player unless they sacrifice another creature of their choice."
//
// Soul Bleed's upkeep watch on the enchanted creature's controller, and a punisher
// asked of that player (UpkeepPayUnless over a sacrifice payment that excepts the
// enchanted creature itself, "another"): a "no", or no other creature to give,
// runs the keyed else branch, which deals the Aura's damage equal to the enchanted
// creature's power as it is then (its last-known power if it has left, CR 608.2h).
// The damage is dealt by the Aura to the player asked.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d37a361e-83e8-4f0a-bd09-3135fb918471",
		Name:         "Unnatural Hunger",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginUpkeep, effects.UpkeepOfEnchantedCreaturesController,
				"Unnatural Hunger — damage equal to that creature's power to that player unless they sacrifice another creature",
				unnaturalHungerAsk),
		},
	})
}

// unnaturalHungerDamage is the "unless" branch: the Aura deals damage equal to the
// enchanted creature's power (carried by ID) to the player who was asked.
var unnaturalHungerDamage = effects.PayUnlessElse("pay-unless/old-b686-unnatural-hunger-damage",
	func(ctx *effects.Context, r game.PayUnlessAnswer) error {
		if len(r.Carry) == 0 {
			return nil
		}
		power := 0
		if c, ok := ctx.Game.LookupCardForEffect(r.Carry[0]); ok && effects.OnBattlefield(ctx.Game, r.Carry[0]) {
			power = c.PowerForComparison()
		} else if info, ok := ctx.Game.LastKnownPermanentForEffect(r.Carry[0]); ok {
			power = info.Power
		}
		dmg := effects.DealDamage{Source: r.Source, Target: r.Chooser, Amount: power}
		return dmg.Apply(ctx)
	})

// unnaturalHungerAsk asks the enchanted creature's controller whether to sacrifice
// another creature; the host is carried so the damage can read its power.
func unnaturalHungerAsk(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	ctx := effects.NewContext(g, item)
	host, ok := effects.EnchantedCreatureAsItResolves(ctx)
	if !ok {
		return nil
	}
	player := item.Trigger.Event.Actor
	_, action := effects.SacrificePayment(1, "creature", "creatures", effects.QueryType("creature")).Prompt()
	a := *action
	a.Except = &host
	return effects.UpkeepPayUnless{
		Chooser:  player,
		Cost:     "Sacrifice another creature",
		Action:   &a,
		Question: "Unnatural Hunger — sacrifice another creature, or take damage equal to the enchanted creature's power?",
		Else:     unnaturalHungerDamage,
		Carry:    []uuid.UUID{host},
	}.Apply(ctx)
}
