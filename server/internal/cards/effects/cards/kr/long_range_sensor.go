package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Long-Range Sensor — Artifact {2}{R}:
//
//	"Whenever you attack a player, put a charge counter on this
//	 artifact.
//	 {1}, Remove two charge counters from this artifact: Discover 4.
//	 Activate only as a sorcery."
//
// "Whenever you attack a player" is once per player attacked in a
// declaration (OncePerBatchPerPlayer, #784), Horizon Explorer's shape.
// The counter removal is a cost (#625). Discover is ADR 0099's
// (game/discover.go).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "358297cd-4a9f-48e2-aaa5-131b0849f8dc",
		Name:         "Long-Range Sensor",
		Completeness: effects.CompletenessFull,
		Discovers:    true,
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatchPerPlayer(effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B16YouAttackedAPlayer(ev, source, g)
			}, "Long-Range Sensor — put a charge counter on it", effects.B33PutChargeCounterOnSelf)),
		},
		Activated: []effects.ActivatedAbility{{
			Label:        "{1}, Remove two charge counters from this artifact: Discover 4",
			Purpose:      game.Purpose{Discover: 4},
			Cost:         effects.Plus(effects.ManaCost("{1}"), effects.RemoveCountersFromThis("charge", 2)),
			SorcerySpeed: true,
			Effect:       effects.DiscoverN(4),
		}},
	})
}
