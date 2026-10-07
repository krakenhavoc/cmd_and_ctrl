package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Secret Arcade // Dusty Parlor — Enchantment — Room (ADR 0103):
//
//	Secret Arcade {4}{W}: "Nonland permanents you control and permanent
//	 spells you control are enchantments in addition to their other
//	 types."
//	Dusty Parlor {2}{W}: "Whenever you cast an enchantment spell, put a
//	 number of +1/+1 counters equal to that spell's mana value on up to
//	 one target creature."
//
// Secret Arcade is Mycosynth Lattice's additive layer 4 shape, aimed at
// your nonland permanents. The "permanent spells" half is NOT
// implemented: the layer pass does not reach a spell on the stack, so a
// creature spell you cast is not an enchantment until it enters. Dusty
// Parlor reads the spell's mana value off the cast event when the
// trigger resolves; the spell is still on the stack then, because the
// trigger resolves first.
func init() {
	Register(Room(RoomSpec{
		OracleID:     "c8abde48-a07a-42a0-a43b-c357e6d9cad4",
		Name:         "Secret Arcade // Dusty Parlor",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Secret Arcade makes your permanents enchantments, but a permanent spell on the stack isn't an enchantment until it enters."},
		Left: Door{Static: []game.StaticAbility{{
			Layer: game.Layer4Type,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.Controller == source.Controller && !target.IsLand()
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				for _, t := range c.Types {
					if t == "Enchantment" {
						return
					}
				}
				c.Types = append(c.Types, "Enchantment")
			},
		}}},
		Right: Door{Triggered: []game.TriggeredAbility{
			Targeting(
				WheneverYouCast(Enchantment(), "Dusty Parlor — put +1/+1 counters equal to the spell's mana value on up to one target creature",
					dustyParlorCounters),
				TargetCreature("up to one target creature").WithCount(0, 1)),
		}},
	}))
}

func dustyParlorCounters(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	spell, ok := g.LookupCardForEffect(item.Trigger.Event.CardID)
	if !ok {
		return nil
	}
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return AddCounter{Target: id, Kind: game.CounterPlusOne, N: spell.ManaValue()}.Apply(ctx)
}
