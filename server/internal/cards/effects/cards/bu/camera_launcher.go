package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Camera Launcher — Artifact Creature — Construct {3}, 2/2 (Aetherdrift,
// #4464):
//
//	"Exhaust — {3}: Put a +1/+1 counter on this creature. Create a 1/1
//	 colorless Thopter artifact creature token with flying. (Activate each
//	 exhaust ability only once.)"
//
// Pacesetter Paragon's exhaust (CR 702.177a: "Activate only once"), keyed
// to the object, so a Launcher back from a flicker may use it again
// (CR 400.7). The counter lands only while the Launcher is still the
// permanent that activated; the Thopter is made either way, since the
// second sentence does not depend on the first. The token is the
// catalog's predefined Thopter.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b46bac37-5b45-4be5-93d0-76e48dbd9f74",
		Name:         "Camera Launcher",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Exhaust — {3}: Put a +1/+1 counter on this creature. Create a 1/1 colorless Thopter artifact creature token with flying.",
			Purpose: game.Purpose{Answers: game.AnswerPump, SourceCounters: 1, Tokens: 1, CreatureTokens: 1},
			Exhaust: true,
			Cost:    effects.ManaCost("{3}"),
			// The token waits for any prompt the counter asks (CR 608.2c).
			Effect: effects.Do(
				effects.Instruction(effects.PlusOneCounterOnThisIfStillHere),
				effects.CreateToken{Template: effects.TokenCard("1/1 colorless Thopter artifact with flying"), N: 1},
			),
		}},
	})
}
