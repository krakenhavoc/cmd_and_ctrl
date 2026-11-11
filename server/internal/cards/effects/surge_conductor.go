package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "dc11d9dc-7d18-45dc-955b-d0ba71fc0db8",
		Name:         "Surge Conductor",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, true)
				return ok && c.IsArtifact() && !c.IsToken()
			}, "Surge Conductor — proliferate", Do(Proliferate{})),
		},
	})
}
