package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "194af2be-50c7-484b-bdd3-402bad70335b",
		Name:            "Donatello, Way with Machines",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.ArtifactEnteredUnderYourControl,
				"Donatello, Way with Machines — a +1/+1 counter on Donatello",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
