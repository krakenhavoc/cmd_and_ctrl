package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Headless Rider — Creature — Zombie {2}{B}, 3/1 (EDHREC rank
// 2866):
//
//	"Whenever this creature or another nontoken Zombie you control
//	 dies, create a 2/2 black Zombie creature token."
//
// The Zombie deck's death-replacement engine. One trigger watching
// EventLTB: the Rider's own death (cardDied — graveyard only) or a
// NONTOKEN Zombie its controller controlled dying, read post-move so
// a changeling counts. The token is the shared 2/2 black Zombie; a
// token dying does not fire it, which is exactly why the printed
// card says "nontoken" — the Rider's own Zombies would chain
// forever otherwise.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d4fdacd7-3101-44e2-a880-dde7326137a4",
		Name:         "Headless Rider",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b27SelfOrNontokenZombieYouControlDied(ev, source, g)
			}, "Headless Rider — create a 2/2 black Zombie", Do(CreateToken{Template: BlackZombieToken(), N: 1})),
		},
	})
}
