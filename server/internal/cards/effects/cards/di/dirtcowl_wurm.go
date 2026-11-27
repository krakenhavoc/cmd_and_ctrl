package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dirtcowl Wurm — Creature — Wurm {4}{G}, 3/4 (Tempest, S82 OLD batch 759,
// #5628):
//
//	"Whenever an opponent plays a land, put a +1/+1 counter on this
//	 creature."
//
// Burgeoning's land-played test (a land played, not put onto the
// battlefield by an effect) on an opponent of the Wurm's controller, one
// counter each time.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "608293f1-34aa-4e55-8389-73f975858de6",
		Name:         "Dirtcowl Wurm",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventZoneMove, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return source != nil && effects.B20LandPlayed(ev, g) && ev.Actor != source.Controller
			}, "Dirtcowl Wurm — put a +1/+1 counter on this creature",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
