package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Forced Adaptation — Enchantment — Aura {G} (Ravnica Remastered):
//
//	"Enchant creature
//	 At the beginning of your upkeep, put a +1/+1 counter on enchanted
//	 creature."
//
// "Your" is the Aura's controller, not the creature's. With no host the
// ability does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1c859c97-74df-4360-baf8-0f79eae97c45",
		Name:         "Forced Adaptation",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Forced Adaptation — put a +1/+1 counter on enchanted creature", forcedAdaptationCounter),
		},
	})
}

func forcedAdaptationCounter(g *game.Game, item *game.StackItem) error {
	host := effects.AttachedHostFor(g, item.SourceCardID)
	if host == nil {
		return nil
	}
	return effects.AddCounter{Target: host.InstanceID, Kind: game.CounterPlusOne, N: 1}.Apply(effects.NewContext(g, item))
}
