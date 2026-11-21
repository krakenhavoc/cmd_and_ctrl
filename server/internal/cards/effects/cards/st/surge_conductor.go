package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Surge Conductor — Artifact Creature — Robot {3}, 3/2 (Edge of
// Eternities Commander):
//
//	"Whenever another nontoken artifact you control enters,
//	 proliferate."
//
// Donatello's artifact-entered watcher (enteredUnderYourControl with
// "another") narrowed to nontoken artifacts, over Flux Channeler's
// proliferate. One trigger per artifact, as the ruling says for artifacts
// that enter together.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dc11d9dc-7d18-45dc-955b-d0ba71fc0db8",
		Name:         "Surge Conductor",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := effects.EnteredUnderYourControl(ev, source, g, true)
				return ok && c.IsArtifact() && !c.IsToken()
			}, "Surge Conductor — proliferate", effects.Do(effects.Proliferate{})),
		},
	})
}
