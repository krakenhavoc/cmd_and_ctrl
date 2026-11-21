package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Towashi Songshaper — Artifact Creature — Human Artificer {1}{R}, 2/2
// (Neon Dynasty):
//
//	"Whenever another artifact you control enters, this creature gets
//	 +1/+0 until end of turn."
//
// Dowsing Device's entry check: the entering permanent is read through the
// layers, so an artifact creature counts, and it must enter under your
// control and not be the Songshaper itself ("another").
//
// No simplification.
func init() {
	const label = "Towashi Songshaper — +1/+0 until end of turn"
	effects.Register(effects.Spec{
		OracleID:     "d2c7d10e-86d9-4e2d-a0da-2e39c172aae2",
		Name:         "Towashi Songshaper",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := effects.EnteredUnderYourControl(ev, source, g, true)
				return ok && c.IsArtifact()
			}, label, effects.ThisGetsUntilEndOfTurn(1, 0, label)),
		},
	})
}
