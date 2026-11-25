package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Narcolepsy — Enchantment — Aura {1}{U} (Jumpstart, #5510):
//
//	"Enchant creature
//	 At the beginning of each upkeep, if enchanted creature is untapped, tap
//	 it."
//
// An intervening "if" (CR 603.4): checked as each player's upkeep begins and
// again as the ability resolves, so a creature that is already tapped asks
// nothing and one tapped in response is not tapped again.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3da1412e-a72c-4da1-8848-37145ca6614a",
		Name:         "Narcolepsy",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondBoard{
				Name: "enchanted creature is untapped",
				Test: func(ctx *effects.Context) bool {
					aura, ok := ctx.Game.LookupCardForEffect(ctx.Source())
					return ok && effects.S82OldB100EnchantedUntapped(ctx.Game, &aura)
				},
			}, effects.On(game.EventStepBegan, effects.AllOf(effects.StepBegan(game.StepUpkeep, false),
				func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return effects.S82OldB100EnchantedUntapped(g, source)
				}), "Narcolepsy — tap enchanted creature", effects.S82OldB100TapEnchanted)),
		},
	})
}
