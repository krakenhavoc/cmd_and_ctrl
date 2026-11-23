package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hydra's Growth — Enchantment — Aura {2}{G} (Jumpstart 2022):
//
//	"Enchant creature
//	 When this Aura enters, put a +1/+1 counter on enchanted creature.
//	 At the beginning of your upkeep, double the number of +1/+1 counters on
//	 enchanted creature."
//
// Both read "enchanted creature" as the creature the Aura is attached to as
// the ability resolves (EnchantedCreatureAsItResolves). The doubling adds as
// many counters as it already has, through AddCounter, so Doubling Season and
// Hardened Scales apply to it as printed; a creature with none gains none.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "559903cd-5cba-47a5-8b5b-7067e393e76a",
		Name:         "Hydra's Growth",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Hydra's Growth — put a +1/+1 counter on enchanted creature", effects.PlusOneCounterOnEnchantedCreature),
			effects.AtYourUpkeep("Hydra's Growth — double the +1/+1 counters on enchanted creature", func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				host, ok := effects.EnchantedCreatureAsItResolves(ctx)
				if !ok {
					return nil
				}
				n := effects.CounterCountOn(g, host, game.CounterPlusOne)
				if n <= 0 {
					return nil
				}
				return effects.AddCounter{Target: host, Kind: game.CounterPlusOne, N: n}.Apply(ctx)
			}),
		},
	})
}
