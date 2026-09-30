package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Monstrous Vortex — Enchantment {3}{G}:
//
//	"Whenever you cast a creature spell with power 5 or greater,
//	 discover X, where X is that spell's mana value."
//
// X is the spell's mana value as the trigger resolves (CR 608.2h). The
// trigger goes on the stack above the spell, so the spell is still
// there, and an X it was cast for counts (CR 202.3e). A spell countered
// in response is read where it went. Discover is ADR 0099's
// (game/discover.go).
func init() {
	Register(Spec{
		OracleID:     "6e27956e-ba2f-41a9-8e39-ec54424591b6",
		Name:         "Monstrous Vortex",
		Completeness: CompletenessFull,
		Discovers:    true,
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(And(Creature(), PowerGE(5)), "Monstrous Vortex — discover X",
				func(g *game.Game, item *game.StackItem) error {
					return Discover{N: triggeringSpellManaValue(g, item)}.Apply(NewContext(g, item))
				}),
		},
	})
}

// triggeringSpellManaValue is "that spell's mana value" for a trigger
// on a cast event, read at resolution: the spell wherever it is now,
// with its announced X while it is still on the stack.
func triggeringSpellManaValue(g *game.Game, item *game.StackItem) int {
	if item.Trigger == nil {
		return 0
	}
	spell, ok := g.LookupCardForEffect(item.Trigger.Event.CardID)
	if !ok {
		return 0
	}
	mv, _ := g.ManaValueForEffect(spell)
	return mv
}
