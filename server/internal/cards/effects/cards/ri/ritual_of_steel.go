package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ritual of Steel — Enchantment — Aura {2}{W} (Mirage, S82 OLD batch 811,
// #5633):
//
//	"Enchant creature
//	 When this Aura enters, draw a card at the beginning of the next turn's
//	 upkeep.
//	 Enchanted creature gets +0/+2."
//
// The enters trigger schedules Mishra's Bauble's delayed draw (CR 603.7) for
// the next upkeep of any turn, controlled by the Aura's controller; the pump
// is the Aura's static.
//
// No simplification.
func init() {
	const label = "Ritual of Steel — draw a card at the beginning of the next turn's upkeep"
	effects.Register(effects.Spec{
		OracleID:     "0af91b63-a874-4184-8f8f-3871429fd1ad",
		Name:         "Ritual of Steel",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit, Draws: 1},
		Static:       []game.StaticAbility{effects.PumpAttached(0, 2)},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters(label, func(g *game.Game, item *game.StackItem) error {
				return effects.ScheduleDelayedTrigger{
					At:    game.StepUpkeep,
					Label: label,
					Body:  effects.PortentDrawBody,
				}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}
