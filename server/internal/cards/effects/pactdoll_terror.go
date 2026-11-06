package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pactdoll Terror — Artifact Creature — Toy {3}{B}, 3/4 (Aetherdrift,
// #4464):
//
//	"Whenever this creature or another artifact you control enters, each
//	 opponent loses 1 life and you gain 1 life."
//
// Dragonspark Reactor's artifact-entry watch, the source included
// (enteredUnderYourControl with another=false), with "this creature"
// matched by identity so the Terror's own entry counts even if it is not
// an artifact as it enters. The drain is a point from each opponent and
// one point of life, whatever the opponents lost. Eliminated seats are
// skipped.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e02fff6b-9ee4-474f-97fc-d70d4bc8001b",
		Name:         "Pactdoll Terror",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				return ok && (c.InstanceID == source.InstanceID || c.IsArtifact())
			}, "Pactdoll Terror — each opponent loses 1 life and you gain 1 life",
				Do(EachOpponentLosesLife{N: 1}, GainLife{Amount: 1})),
		},
	})
}
