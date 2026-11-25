package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Faith of the Devoted — Enchantment {2}{B} (Commander 2019, #5517):
//
//	"Whenever you cycle or discard a card, you may pay {1}. If you do, each
//	 opponent loses 2 life and you gain 2 life."
//
// "Cycle or discard" is ONE event, the discard (Archfiend of Ifnir): cycling's
// cost is to discard the card (CR 702.29d), so a cycled card is a discarded
// card and watching EventCycle as well would count one cycle twice. One
// trigger per card, as printed. The {1} is paid as the trigger resolves
// ("you may pay {1}. If you do"), and the drain is life loss, not damage.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "99af5247-06cf-40b2-9a7c-65ad59b1ddf6",
		Name:         "Faith of the Devoted",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDiscardCard, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.DiscardedByYou(ev, source)
			}, "Faith of the Devoted — you may pay {1}: each opponent loses 2 life and you gain 2 life", effects.OldB153FaithOffer),
		},
	})
}
