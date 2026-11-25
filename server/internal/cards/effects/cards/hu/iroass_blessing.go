package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Iroas's Blessing — Enchantment — Aura {3}{R}:
//
//	"Enchant creature you control
//	 When this Aura enters, it deals 4 damage to target creature or
//	 planeswalker an opponent controls.
//	 Enchanted creature gets +1/+1."
//
// The Aura is the source of the damage; the trigger's target is chosen
// when the Aura has entered, apart from the Aura's own enchant target.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "76905904-b3e4-44ee-bfcb-e518f0815af5",
		Name:         "Iroas's Blessing",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(effects.YouControl()),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.PumpAttached(1, 1)},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Iroas's Blessing — 4 damage to target creature or planeswalker an opponent controls",
					func(g *game.Game, item *game.StackItem) error {
						if len(item.Targets) == 0 {
							return nil
						}
						ref, ok := effects.NewContext(g, item).SourceRef()
						if !ok {
							return nil
						}
						return effects.DealDamage{SourceObject: &ref, Target: item.Targets[0].ID, Amount: 4}.Apply(effects.NewContext(g, item))
					}),
				effects.TargetPermanent("target creature or planeswalker an opponent controls",
					effects.Or(effects.Creature(), effects.Planeswalker()), effects.OpponentControls())),
		},
	})
}
