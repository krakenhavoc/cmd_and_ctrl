package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Headless Rider — Creature — Zombie {2}{B}, 3/1 (EDHREC rank
// 2866):
//
//	"Whenever this creature or another nontoken Zombie you control
//	 dies, create a 2/2 black Zombie creature token."
//
// The Zombie deck's death-replacement engine. One trigger watching
// EventLTB: the Rider's own death (cardDied — graveyard only) or a
// NONTOKEN Zombie its controller controlled dying — a Zombie as it
// last existed on the battlefield (CR 603.10a, #1679), so a
// changeling counts and so does a creature that was a Zombie only
// through a grant. The token is the shared 2/2 black Zombie; a
// token dying does not fire it, which is exactly why the printed
// card says "nontoken" — the Rider's own Zombies would chain
// forever otherwise.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d4fdacd7-3101-44e2-a880-dde7326137a4",
		Name:         "Headless Rider",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B27SelfOrNontokenZombieYouControlDied(ev, source, g)
			}, "Headless Rider — create a 2/2 black Zombie", effects.Do(effects.CreateToken{Template: effects.BlackZombieToken(), N: 1})),
		},
	})
}
