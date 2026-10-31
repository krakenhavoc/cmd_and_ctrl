package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Machinesmith Automaton — Artifact Creature — Kree Robot Villain {2}{R}, 2/2
// (Marvel Super Heroes, #4455):
//
//	"Trample
//	 Whenever another artifact you control enters, put a +1/+1 counter on
//	 this creature."
//
// Dragonspark Reactor's artifact-entry watch, minus the source itself
// (enteredUnderYourControl with another=true). Trample is the printed
// keyword.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "945e8443-9647-4eab-b091-ee8e7bc347f5",
		Name:            "Machinesmith Automaton",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, true)
				return ok && c.IsArtifact()
			}, "Machinesmith Automaton — a +1/+1 counter on it", Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
