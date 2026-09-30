package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Geological Appraiser — Creature — Human Artificer {2}{R}{R}, 3/2:
//
//	"When this creature enters, if you cast it, discover 3."
//
// "If you cast it" is an intervening if (CR 603.4), read off the
// permanent's cast provenance (CR 400.7d, game/cast_provenance.go): a
// reanimated, flickered or discovered-into-play Appraiser was not cast
// as it entered and discovers nothing. Discover is ADR 0099's
// (game/discover.go).
func init() {
	Register(Spec{
		OracleID:     "a0a1cd2b-ed7c-4cce-8497-4873e1e5e3df",
		Name:         "Geological Appraiser",
		Completeness: CompletenessFull,
		Discovers:    true,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, AllOf(Self, geologicalAppraiserWasCast), "Geological Appraiser — discover 3",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if ctx.CastProvenance().FromZone == "" {
						return nil
					}
					return Discover{N: 3}.Apply(ctx)
				}),
		},
	})
}

// geologicalAppraiserWasCast is "if you cast it": the entering permanent
// came from a spell.
func geologicalAppraiserWasCast(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return source != nil && source.Provenance.FromZone != ""
}
