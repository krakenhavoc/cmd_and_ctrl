package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// destroyNonlandsWithManaValueOfChargeCounters is "Destroy each
// nonland permanent with mana value equal to the number of charge
// counters on this <permanent>" as an ability's effect — Engineered
// Explosives and Blast Zone. The permanent is sacrificed as a cost, so
// the number is its last-known information (CR 608.2h), read through
// ctx.SourcePermanent rather than the graveyard card, whose counters
// are gone. A token's mana value is 0.
func destroyNonlandsWithManaValueOfChargeCounters(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	info, ok := ctx.SourcePermanent()
	if !ok {
		return nil
	}
	return DestroyAllMatching{Match: And(Nonland(), b03ManaValueIs(info.Counters[game.CounterCharge]))}.Apply(ctx)
}
