package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "358297cd-4a9f-48e2-aaa5-131b0849f8dc",
		Name:         "Long-Range Sensor",
		Completeness: CompletenessFull,
		Discovers:    true,
		Triggered: []game.TriggeredAbility{
			OncePerBatchPerPlayer(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b16YouAttackedAPlayer(ev, source, g)
			}, "Long-Range Sensor — put a charge counter on it", b33PutChargeCounterOnSelf)),
		},
		Activated: []ActivatedAbility{{
			Label:        "{1}, Remove two charge counters from this artifact: Discover 4",
			Cost:         Plus(ManaCost("{1}"), RemoveCountersFromThis("charge", 2)),
			SorcerySpeed: true,
			Effect:       DiscoverN(4),
		}},
	})
}
