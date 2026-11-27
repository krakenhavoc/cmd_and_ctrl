package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_old_b534_helpers.go — S82 OLD batch 534 (#5591, Shadowmoor auras).

// MinusOneCounterOnEnchantedCreature is "put a -1/-1 counter on enchanted
// creature" as an Aura's own ability (Torture, Biting Tether): the creature the
// Aura is attached to as the ability resolves, or was attached to when it left
// (CR 608.2h). Nothing when the Aura was attached to no card.
func MinusOneCounterOnEnchantedCreature(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	host, ok := EnchantedCreatureAsItResolves(ctx)
	if !ok {
		return nil
	}
	return AddCounter{Target: host, Kind: game.CounterMinusOne, N: 1}.Apply(ctx)
}
