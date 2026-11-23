package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Burgeoning — Enchantment {G} (EDHREC top 1,000, #3654):
//
//	"Whenever an opponent plays a land, you may put a land card from
//	 your hand onto the battlefield."
//
// Horn of Greed's land-play read (b20LandPlayed: a PLAY, not a fetch or
// a flicker), narrowed to opponents, then Atarka's Command's "may put a
// land card from your hand". Your own land plays do not trigger it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "88cf4c1a-5863-45fc-b76d-7ad33b175b2b",
		Name:         "Burgeoning",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventZoneMove, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return source != nil && effects.B20LandPlayed(ev, g) && ev.Actor != source.Controller
			}, "Burgeoning — you may put a land card from your hand onto the battlefield",
				effects.Do(effects.MayPutALandFromHand("Burgeoning"))),
		},
	})
}
