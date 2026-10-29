package effects

import (
	"sort"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Enchanted River's Grasp — Enchantment — Aura {2}{U} (The Hobbit):
//
//	"Enchant creature
//	 When this Aura enters, tap enchanted creature and remove all counters
//	 from it.
//	 Enchanted creature loses all abilities and doesn't untap during its
//	 controller's untap step."
//
// Infinite Coursework's shape (the same two statics, an entry trigger that
// taps) with the counters off instead of the unprepare. The trigger reads the
// Aura's attachment as it resolves, so an Aura that has left in response
// still taps the creature it last enchanted (CR 608.2h). Every counter of
// every kind comes off, a kind at a time in name order, each through the
// effect-side adder at a negative delta (as Sin, Unending Cataclysm does) so
// the layer pass sees the counters go: a +1/+1 counter stops pumping the
// creature, and the counters stay off when the Aura leaves, because only the
// statics end with it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:              "fcc9f3bc-0bc3-4189-a4d0-9698594b99a8",
		Name:                  "Enchanted River's Grasp",
		Completeness:          CompletenessFull,
		Targets:               EnchantCreature(),
		Purpose:               game.Purpose{Aura: game.AuraHarm},
		UntapStepRestrictions: []game.UntapStepRestriction{enchantedDoesntUntap()},
		Static:                []game.StaticAbility{LoseAllAbilities()},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Enchanted River's Grasp — tap enchanted creature and remove all counters from it",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					aura, ok := ctx.TriggeringPermanent()
					if !ok || aura.AttachedTo.Kind != game.TargetCard {
						return nil
					}
					if err := (TapTarget{Target: aura.AttachedTo.ID}).Apply(ctx); err != nil {
						return err
					}
					return removeEveryCounterFrom(ctx, aura.AttachedTo.ID)
				}),
		},
	})
}

// removeEveryCounterFrom is "remove all counters from [a permanent]": every
// kind it carries, in name order, a kind at a time. A permanent that has gone
// has none to lose.
func removeEveryCounterFrom(ctx *Context, id uuid.UUID) error {
	c, ok := ctx.Game.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	kinds := make([]string, 0, len(c.Counters))
	for kind, n := range c.Counters {
		if n > 0 {
			kinds = append(kinds, kind)
		}
	}
	sort.Strings(kinds)
	for _, kind := range kinds {
		if err := (AddCounter{Target: id, Kind: kind, N: -c.Counters[kind]}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
