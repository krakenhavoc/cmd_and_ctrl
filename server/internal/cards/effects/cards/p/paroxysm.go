package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Paroxysm — Enchantment — Aura {1}{R} (Exodus, S82 OLD b739, #5626):
//
//	"Enchant creature
//	 At the beginning of the upkeep of enchanted creature's controller, that
//	 player reveals the top card of their library. If that card is a land
//	 card, destroy that creature. Otherwise, it gets +3/+3 until end of turn."
//
// The trigger is the Aura controller's but watches the enchanted creature's
// controller's upkeep (Unnatural Hunger's gate). That player reveals the top
// card of their library; a land destroys the creature the Aura is attached to as
// the ability resolves, anything else (an empty library included) gives it
// +3/+3 until end of turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f00dd41d-4b11-491a-9e7b-12245092ab77",
		Name:         "Paroxysm",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginUpkeep, effects.UpkeepOfEnchantedCreaturesController,
				"Paroxysm — that player reveals the top card of their library; if it is a land card destroy that creature, otherwise it gets +3/+3 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					ctx := effects.NewContext(g, item)
					host, ok := effects.EnchantedCreatureAsItResolves(ctx)
					if !ok || !effects.OnBattlefield(g, host) {
						return nil
					}
					player := item.Trigger.Event.Actor
					isLand := false
					for _, id := range g.RevealTopOfLibraryForEffect(player, item.SourceCardID, 1, "Paroxysm — reveal the top card of your library") {
						if c, found := g.LookupCardForEffect(id); found && c.IsLand() {
							isLand = true
						}
					}
					if isLand {
						return effects.DestroyTarget{Target: host}.Apply(ctx)
					}
					return effects.BoostUntilEOT{Target: host, Power: 3, Toughness: 3, Label: "Paroxysm — +3/+3"}.Apply(ctx)
				}),
		},
	})
}
