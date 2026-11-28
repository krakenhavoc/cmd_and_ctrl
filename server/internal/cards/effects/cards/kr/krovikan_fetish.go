package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Krovikan Fetish — Enchantment — Aura {2}{B} (Ice Age, S82 OLD b828, #5636):
//
//	"Enchant creature
//	 When this Aura enters, draw a card at the beginning of the next turn's
//	 upkeep.
//	 Enchanted creature gets +1/+1."
//
// The enters trigger schedules the delayed draw (DrawACardAtTheNextUpkeep,
// Fevered Strength's); PumpAttached is the static.
//
// No simplification.
func init() {
	const label = "Krovikan Fetish — draw a card at the beginning of the next turn's upkeep"
	effects.Register(effects.Spec{
		OracleID:     "20c7630d-a97f-4f09-9ec5-b7a1739107de",
		Name:         "Krovikan Fetish",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit, Draws: 1},
		Static:       []game.StaticAbility{effects.PumpAttached(1, 1)},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters(label, func(g *game.Game, item *game.StackItem) error {
				return effects.DrawACardAtTheNextUpkeep(label).Apply(effects.NewContext(g, item))
			}),
		},
	})
}
