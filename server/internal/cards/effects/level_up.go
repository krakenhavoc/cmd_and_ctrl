package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Level Up — {1}{G} Enchantment — Aura (Teenage Mutant Ninja Turtles
// Commander, #5011):
//
//	"Enchant creature
//	 When this Aura enters, put a +1/+1 counter on enchanted creature.
//	 Enchanted creature has "Whenever this creature attacks, double the
//	 number of +1/+1 counters on it. Then if it has power 10 or greater,
//	 draw a card.""
//
// The enters trigger is Accelerated Evolution's shape: the creature the
// Aura enchants as the trigger resolves (enchantedCreatureAsItResolves)
// gets the counter, and a host that has left gets nothing (CR 608.2b).
//
// The quoted ability is a granted bundle (ADR 0093, Clever Distraction's
// shape), so it is the creature's own trigger and its controller draws. It
// doubles by putting as many +1/+1 counters on the creature as it already
// has (the printed ruling), through the keyed placement so a Hardened
// Scales or High Score applies, and the power check reads the creature
// after the placement has landed, as the tail of that placement. With no
// counter to double the check still happens. The power is the creature's
// through the layers, or as it last existed if it has left (CR 608.2h).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "88cd43ce-b168-45a4-8d29-6cac5ab42857",
		Name:         "Level Up",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Level Up — put a +1/+1 counter on enchanted creature", levelUpAuraEnters),
		},
		Grants: []AbilityGrant{{
			Key: levelUpAuraAttackKey,
			Triggered: []game.TriggeredAbility{
				WheneverThisAttacks("Level Up — double the +1/+1 counters on this creature, then draw if its power is 10 or greater", levelUpAuraDouble),
			},
			Text: "Whenever this creature attacks, double the number of +1/+1 counters on it. Then if it has power 10 or greater, draw a card.",
		}},
		Static: []game.StaticAbility{GrantAbilitiesToAttached(levelUpAuraAttackKey)},
	})
}

const levelUpAuraAttackKey = "level-up/attack-double-then-draw"

// levelUpAuraEnters puts the +1/+1 counter on the creature the Aura
// enchants, if it is still on the battlefield.
func levelUpAuraEnters(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	host, ok := enchantedCreatureAsItResolves(ctx)
	if !ok || !onBattlefield(g, host) {
		return nil
	}
	return AddCounter{Target: host, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
}

// levelUpAuraDouble is the granted attack trigger: as many +1/+1 counters
// as the creature has, then the power check once they have landed.
func levelUpAuraDouble(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id := item.SourceCardID
	then := ctx.ItemTail(levelUpAuraThenDraw)
	then.Carry = []uuid.UUID{id}
	if !onBattlefield(g, id) {
		return then.Run(g, nil)
	}
	c, ok := g.LookupCardForEffect(id)
	if !ok || c.Counters[game.CounterPlusOne] <= 0 {
		return then.Run(g, nil)
	}
	return g.AddCounterByKeyedForEffect(uuid.Nil, id, game.CounterPlusOne, c.Counters[game.CounterPlusOne], then)
}

// levelUpAuraThenDraw is "Then if it has power 10 or greater, draw a card";
// Carry[0] is the creature. An on-disk key: never renamed, never reused.
var levelUpAuraThenDraw = CardTail("level-up/then-draw-if-power-10", func(ctx *Context, o game.TailOutcome) error {
	id := carryFirst(o.Tail)
	if id == uuid.Nil {
		return nil
	}
	ctx.Game.RecomputeLayersIfStaleLocked()
	if b43PowerNowOrLastKnown(ctx.Game, id) < 10 {
		return nil
	}
	return DrawCards{N: 1}.Apply(ctx)
})
