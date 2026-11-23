package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// PlusOneCounterOnEnchantedCreature is "put a +1/+1 counter on enchanted
// creature" as an Aura's own ability (Stark's Ingenuity, Hydra's Growth): the
// creature the Aura is attached to as the ability resolves, or was attached to
// when it left (CR 608.2h). Nothing when the Aura was attached to no card.
func PlusOneCounterOnEnchantedCreature(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	host, ok := EnchantedCreatureAsItResolves(ctx)
	if !ok {
		return nil
	}
	return AddCounter{Target: host, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
}

// OrdealAttackCounterThenSacrifice is the attack trigger of the Ordeal Auras
// (Ordeal of Nylea, Ordeal of Purphoros): "Whenever enchanted creature
// attacks, put a +1/+1 counter on it. Then if it has three or more +1/+1
// counters on it, sacrifice this Aura." The counter and the count are one
// resolution, so a third counter put on any other way waits for the next
// attack.
func OrdealAttackCounterThenSacrifice(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	host := item.Trigger.Event.CardID
	ctx := NewContext(g, item)
	if err := (AddCounter{Target: host, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
		return err
	}
	c, ok := g.LookupCardForEffect(host)
	if !ok || c.Counters[game.CounterPlusOne] < 3 {
		return nil
	}
	return SacrificePermanent{Target: item.SourceCardID}.Apply(ctx)
}
