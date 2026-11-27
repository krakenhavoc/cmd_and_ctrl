package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Segmented Wurm — Creature — Wurm {3}{R}{G}, 5/5 (Tempest, S82 OLD batch
// 760, #5628):
//
//	"Whenever this creature becomes the target of a spell or ability, put a
//	 -1/-1 counter on it."
//
// Mossdog's becomes-the-target trigger without its opponent-only test: any
// spell or ability, from any player, counts, once per target announcement.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "eb1baf72-4ef6-4967-8a22-2022c633f3d3",
		Name:         "Segmented Wurm",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesTarget, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Kind == game.EventBecomesTarget && ev.CardID == source.InstanceID
			}, "Segmented Wurm — put a -1/-1 counter on this creature",
				effects.Do(effects.CounterOnThis{Kind: game.CounterMinusOne, N: 1})),
		},
	})
}
