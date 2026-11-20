package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bloodbriar — Creature — Plant Elemental {2}{G}, 2/3 (Eldritch Moon,
// #5541):
//
//	"Whenever you sacrifice another permanent, put a +1/+1 counter on this
//	 creature."
//
// Gixian Infiltrator's trigger, word for word: one EventSacrifice per
// permanent its controller sacrifices (b23YouSacrificedAPermanent), the
// Bloodbriar itself excluded ("another").
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7aef2a40-ae18-46d6-9873-0e5056345b81",
		Name:         "Bloodbriar",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventSacrifice, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return b23YouSacrificedAPermanent(ev, source, lki, g) && ev.CardID != source.InstanceID
			}, "Bloodbriar — a +1/+1 counter on it", Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
