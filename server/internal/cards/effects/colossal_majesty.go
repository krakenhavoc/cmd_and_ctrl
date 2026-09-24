package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Colossal Majesty — Enchantment {2}{G} (EDHREC rank 1342):
//
//	"At the beginning of your upkeep, if you control a creature with
//	 power 4 or greater, draw a card."
//
// Garruk's Uprising's upkeep half on its own. The intervening-if (CR
// 603.4) is checked when the upkeep begins AND again on resolution —
// so a big creature killed in response to the trigger leaves no card,
// which is the printed outcome and the weaker direction. Power is
// the current power (counters and anthems count), the same read
// youControlPowerFourOrGreater gives the Uprising.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cac95494-0db0-4bec-8665-998431a6f76b",
		Name:         "Colossal Majesty",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginUpkeep, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Actor == source.Controller && youControlPowerFourOrGreater(g, source.Controller)
			}, "Colossal Majesty — draw a card", drawIfYouControlPowerFourOrGreater),
		},
	})
}
