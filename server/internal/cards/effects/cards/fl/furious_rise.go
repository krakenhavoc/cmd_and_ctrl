package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Furious Rise — Enchantment {2}{R}:
//
//	"At the beginning of your end step, if you control a creature with
//	 power 4 or greater, exile the top card of your library. You may
//	 play that card until you exile another card with this enchantment."
//
// The intervening-if (CR 603.4) is checked as the end step begins and
// again on resolution, with the current power Colossal Majesty reads
// (youControlPowerFourOrGreater). The exile is #2539's window
// (exileTopUntilYouExileAnother): it closes when this enchantment's
// trigger next exiles a card for the same player, and outlasts the
// enchantment if it leaves (ruling 2020-06-23). A land exiled this way
// needs a land play to be played (ruling 2020-06-23).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "52b3a979-3b93-4441-9237-0c8a754e5523",
		Name:         "Furious Rise",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginEndStep, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Actor == source.Controller && effects.YouControlPowerFourOrGreater(g, source.Controller)
			}, "Furious Rise — exile the top card of your library", func(g *game.Game, item *game.StackItem) error {
				if !effects.YouControlPowerFourOrGreater(g, item.Controller) {
					return nil
				}
				return effects.ExileTopUntilYouExileAnother(g, item)
			}),
		},
	})
}
