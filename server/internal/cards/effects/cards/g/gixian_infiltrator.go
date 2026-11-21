package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gixian Infiltrator — Creature — Phyrexian Human {1}{B}, 2/1 (The
// Brothers' War, S82):
//
//	"Whenever you sacrifice another permanent, put a +1/+1 counter on this
//	 creature."
//
// One EventSacrifice per permanent, actor the Infiltrator's controller
// (b23YouSacrificedAPermanent); the Infiltrator itself is excluded
// ("another"). It fires while the sacrificed permanent is still on the
// battlefield, so a Treasure, a Clue or a token counts.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2497926c-b253-4400-b540-776ee2cd807c",
		Name:         "Gixian Infiltrator",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventSacrifice, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return effects.B23YouSacrificedAPermanent(ev, source, lki, g) && ev.CardID != source.InstanceID
			}, "Gixian Infiltrator — a +1/+1 counter on it", effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
