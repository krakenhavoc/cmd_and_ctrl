package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fruit of the First Tree — {3}{G} Enchantment — Aura (Fate Reforged):
//
//	"Enchant creature
//	 When enchanted creature dies, you gain X life and draw X cards, where X is
//	 its toughness."
//
// X is the creature's toughness as it last existed on the battlefield (CR
// 608.2h), read through the layers at the moment it died, so a pumped or shrunk
// creature counts as it was. "You" is the Aura's controller (CR 109.5).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8d579f04-b494-41a1-ba7c-b9779f75b818",
		Name:         "Fruit of the First Tree",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Triggered: []game.TriggeredAbility{
			effects.WhenEnchantedCreatureDies("Fruit of the First Tree — you gain X life and draw X cards, where X is its toughness",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					x := effects.DepartedCreatureToughness(g, item.Trigger.Event.CardID)
					if x <= 0 {
						return nil
					}
					ctx := effects.NewContext(g, item)
					if err := (effects.GainLife{Amount: x}).Apply(ctx); err != nil {
						return err
					}
					return effects.DrawCards{N: x}.Apply(ctx)
				}),
		},
	})
}
