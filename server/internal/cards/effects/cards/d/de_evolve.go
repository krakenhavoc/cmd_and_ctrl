package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// De-evolve — Enchantment — Aura {1}{B}:
//
//	"Enchant creature
//	 Enchanted creature gets -4/-0, can't block, and is a Salamander. (It
//	 loses all other creature types.)
//	 Threshold — At the beginning of each upkeep, if there are seven or
//	 more cards in your graveyard, exile enchanted creature."
//
// Three statics on the host: Aether Meltdown's -4/-0, a can't-block
// restriction, and Fog on the Barrow-Downs' layer 4 replacement of the
// creature types alone (the host keeps its card types). The threshold
// trigger watches every player's upkeep and carries an intervening "if"
// (CR 603.4): the Aura's controller's graveyard is counted as the upkeep
// begins and again as the trigger resolves. The creature exiled is the
// one the Aura is attached to as the trigger resolves, or was attached to
// when it left (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "663e120a-a74f-47fb-aa0a-9d1a75db626e",
		Name:         "De-evolve",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static: []game.StaticAbility{
			effects.PumpEnchantedCreature(-4, 0),
			effects.RestrictAttached(game.CantBlock),
			effects.SetAttachedCreatureTypes("Salamander"),
		},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginUpkeep, func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.FraThresholdMet(g, source.Controller)
			}, "De-evolve — threshold: exile enchanted creature", func(g *game.Game, item *game.StackItem) error {
				if !effects.FraThresholdMet(g, item.Controller) {
					return nil
				}
				return effects.ExileEnchantedCreature(g, item)
			}),
		},
	})
}
