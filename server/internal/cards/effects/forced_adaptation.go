package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "1c859c97-74df-4360-baf8-0f79eae97c45",
		Name:         "Forced Adaptation",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Forced Adaptation — put a +1/+1 counter on enchanted creature", forcedAdaptationCounter),
		},
	})
}

func forcedAdaptationCounter(g *game.Game, item *game.StackItem) error {
	host := attachedHostFor(g, item.SourceCardID)
	if host == nil {
		return nil
	}
	return AddCounter{Target: host.InstanceID, Kind: game.CounterPlusOne, N: 1}.Apply(NewContext(g, item))
}
