package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rebellion of the Flamekin — Kindred Enchantment — Elemental {3}{R} (Lorwyn,
// S82 OLD, #5593):
//
//	"Whenever you clash, you may pay {1}. If you do, create a 3/1 red Elemental
//	 Shaman creature token. If you won, that token gains haste until end of turn.
//	 (This ability triggers after the clash ends.)"
//
// One trigger on the clash event (CR 701.30a) of its controller, whoever began
// the clash: the pay-or-not is asked as it resolves, and the token is made only
// when {1} was paid. Whether the controller won is read off the clash's own win
// event, which shares the trigger event's batch.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0e2a283b-c06a-4b54-a627-5b4959318b1a",
		Name:         "Rebellion of the Flamekin",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventClashed, effects.S82OldB554YouClashed,
				"Rebellion of the Flamekin — you may pay {1} to create a 3/1 red Elemental Shaman", effects.S82OldB554MayPayForAFlamekinShaman),
		},
	})
}
