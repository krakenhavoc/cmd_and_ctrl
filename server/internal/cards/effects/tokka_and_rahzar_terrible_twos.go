package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tokka & Rahzar, Terrible Twos — Legendary Creature — Turtle Wolf Mutant
// {B/R}{B/R}, 3/2 (Teenage Mutant Ninja Turtles, #4457):
//
//	"This spell can't be countered.
//	 Menace
//	 Whenever a player casts a spell, if the amount of mana spent to cast
//	 it was less than its mana value, Tokka & Rahzar deal 3 damage to
//	 that player."
//
// The punisher is Roiling Vortex's cast trigger with a different
// comparison: the spell's own payment record (StackItemPaidForEffect,
// the count Kurbis reads) against its mana value on the stack, X
// included (CR 202.3e). A free cast, a cost reduction (affinity,
// convoke's tapped creatures), or any alternative cost cheaper than the
// mana value all trigger it; a copy is not cast and never does. The
// intervening if (CR 603.4) is read as the spell is cast and again as
// the trigger resolves while the spell is still on the stack. Neither
// number can change once the spell is cast (CR 601.2h), so a spell
// countered in response keeps the answer it had when it was cast.
//
// Like Roiling Vortex it declares ReadsEveryCast: a payment the engine
// waived asks its caster what they paid outside the pool, so the amount
// is exact.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "86ef86de-104f-4eae-a2db-893b50e95ad7",
		Name:            "Tokka & Rahzar, Terrible Twos",
		ReadsManaSpent:  game.ReadsEveryCast,
		Completeness:    CompletenessFull,
		CantBeCountered: true,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Actor != uuid.Nil && tokkaAndRahzarSpentLessThanManaValue(g, ev.CardID)
			}, "Tokka & Rahzar, Terrible Twos — 3 damage to a player who spent less mana than the spell's mana value",
				func(g *game.Game, item *game.StackItem) error {
					p := triggeringActor(item)
					if p == uuid.Nil {
						return nil
					}
					spell := item.Trigger.Event.CardID
					if g.Stack != nil && g.Stack.Contains(spell) && !tokkaAndRahzarSpentLessThanManaValue(g, spell) {
						return nil
					}
					return DealDamage{Source: item.SourceCardID, Target: p, Amount: 3}.Apply(NewContext(g, item))
				}),
		},
	})
}

// tokkaAndRahzarSpentLessThanManaValue is the intervening if for a spell
// on the stack: the mana spent to cast it is less than its mana value. A
// spell whose cost cannot be read has no mana value to fall short of.
func tokkaAndRahzarSpentLessThanManaValue(g *game.Game, spellID uuid.UUID) bool {
	mv, ok := g.LastKnownSpellManaValueForEffect(spellID)
	if !ok {
		return false
	}
	return g.StackItemPaidForEffect(spellID).ManaSpentCount() < mv
}
