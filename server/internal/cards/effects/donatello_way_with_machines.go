package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Donatello, Way with Machines — {2}{U} Legendary Creature — Mutant Ninja
// Turtle 1/2 (Teenage Mutant Ninja Turtles, #4457):
//
//	"Flying
//	 Whenever an artifact you control enters, put a +1/+1 counter on
//	 Donatello."
//
// Machinesmith Automaton's artifact-entry trigger without "another":
// ArtifactEnteredUnderYourControl counts the source too, so a Donatello
// that has become an artifact sees its own entry. The counter goes on the
// Donatello whose trigger it is, and only while it is still on the
// battlefield (CounterOnThis, CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "194af2be-50c7-484b-bdd3-402bad70335b",
		Name:            "Donatello, Way with Machines",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, ArtifactEnteredUnderYourControl,
				"Donatello, Way with Machines — a +1/+1 counter on Donatello",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
