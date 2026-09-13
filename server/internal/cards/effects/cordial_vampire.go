package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cordial Vampire — Creature — Vampire {B}{B}, 1/1 (EDHREC rank
// 1863):
//
//	"Whenever this creature or another creature dies, put a +1/+1
//	 counter on each Vampire you control."
//
// The Vampire deck's aristocrat: every death anywhere at the table
// grows the whole team. Its own death counts too, and then the
// counters go on every OTHER Vampire its controller controls,
// because the Vampire itself is in the graveyard by the time the
// trigger resolves — as printed. Vampires are read from effective
// subtypes at resolution, so a changeling grows.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d61fb9e8-d05a-481a-a90f-5def300c9abb",
		Name:         "Cordial Vampire",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b17SelfOrAnotherCreatureDied(ev, source, g)
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Cordial Vampire — a +1/+1 counter on each Vampire you control",
					b17PutCounterOnEachVampireYouControl)
			},
		}},
	})
}
