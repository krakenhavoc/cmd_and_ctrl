package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lightless Evangel — Creature — Vampire Cleric {1}{B}, 2/2 (Edge of
// Eternities, #4461):
//
//	"Whenever you sacrifice another creature or artifact, put a +1/+1
//	 counter on this creature."
//
// Ravenous Squirrel's trigger (b28YouSacrificedArtifactOrCreature: you
// are the one who sacrificed it, and it was a creature or an artifact)
// with Giant Fly's "another": sacrificing the Evangel itself is not one.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b162ccf8-e0c2-46ae-84e7-dff2b6fceec1",
		Name:         "Lightless Evangel",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventSacrifice, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.CardID != source.InstanceID && b28YouSacrificedArtifactOrCreature(ev, source, g)
			}, "Lightless Evangel — put a +1/+1 counter on this creature", Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
