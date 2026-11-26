package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Glassdust Hulk — Artifact Creature — Golem {3}{W}{U}, 3/4 (Alara Reborn,
// #5586):
//
//	"Whenever another artifact you control enters, this creature gets +1/+1
//	 until end of turn and can't be blocked this turn.
//	 Cycling {W/U}"
//
// Weldfast Wingsmith's "artifact entered under your control" watcher, which
// excludes the Hulk itself (it is an artifact too), with Nimrodel Watcher's
// pump-and-unblockable body.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a1379c25-6daf-4efa-a517-6e6e25c31875",
		Name:         "Glassdust Hulk",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.Cycling("{W/U}")},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.IsArtifactEnteredUnderYourControl(ev, source, g)
			}, "Glassdust Hulk — this creature gets +1/+1 until end of turn and can't be blocked this turn",
				effects.S82OldB501GlassdustHulk("Glassdust Hulk")),
		},
	})
}
