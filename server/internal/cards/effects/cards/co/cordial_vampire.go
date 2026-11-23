package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "d61fb9e8-d05a-481a-a90f-5def300c9abb",
		Name:         "Cordial Vampire",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B17SelfOrAnotherCreatureDied(ev, source, g)
			}, "Cordial Vampire — a +1/+1 counter on each Vampire you control", effects.B17PutCounterOnEachVampireYouControl),
		},
	})
}
