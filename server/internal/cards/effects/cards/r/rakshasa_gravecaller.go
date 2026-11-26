package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rakshasa Gravecaller — Creature — Demon {4}{B}, 3/6 (Dragons of Tarkir, #5551):
//
//	"Exploit (When this creature enters, you may sacrifice a creature.)
//	 When this creature exploits a creature, create two 2/2 black Zombie creature tokens."
//
// Exploit is the shared constructor (exploit.go, CR 702.110). "When this
// creature exploits a creature" is its own triggered ability (CR 702.110b),
// written as a watcher of the exploited event the exploit ability emits once a
// creature really was sacrificed, and only for this creature as the exploiter,
// including when it sacrificed itself. As a one-sentence token trigger it is
// an effects.Do row, so it declares its footprint.
// The tokens are the catalogued 2/2 black Zombie.
//
// No simplification.

func init() {
	effects.Register(effects.Spec{
		OracleID:     "1b57e9f8-68af-4f83-9062-6d715857dcef",
		Name:         "Rakshasa Gravecaller",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Exploit("Rakshasa Gravecaller", game.BodyRef{}),
			effects.On(game.EventExploited, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Kind == game.EventExploited && ev.CardID == source.InstanceID && ev.Actor == source.Controller
			}, "Rakshasa Gravecaller — create two 2/2 black Zombie creature tokens",
				effects.Do(effects.CreateToken{Template: effects.BlackZombieToken(), N: 2})),
		},
	})
}
