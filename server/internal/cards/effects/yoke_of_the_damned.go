package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Yoke of the Damned — Enchantment — Aura {1}{B} (Conflux, #5587):
//
//	"Enchant creature
//	 When a creature dies, destroy enchanted creature."
//
// Black Market's "a creature dies" condition (any creature, anyone's)
// on an Aura. The enchanted creature is the one the Aura is attached to
// as the trigger resolves: if the Aura has fallen off by then, or the
// enchanted creature was the one that died, nothing is destroyed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "af08dae2-ba32-4a34-b8c1-b49d45ce0b6b",
		Name:         "Yoke of the Damned",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, ok := diedCreature(ev, g)
				return ok
			}, "Yoke of the Damned — destroy enchanted creature", func(g *game.Game, item *game.StackItem) error {
				host := attachedHostFor(g, item.SourceCardID)
				if host == nil {
					return nil
				}
				return DestroyTarget{Target: host.InstanceID}.Apply(NewContext(g, item))
			}),
		},
	})
}
