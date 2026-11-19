package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Memorial to Glory — Land:
//
//	"This land enters tapped.
//	 {T}: Add {W}.
//	 {3}{W}, {T}, Sacrifice this land: Create two 1/1 white Soldier
//	 creature tokens."
//
// The white Memorial: a tapped Plains that makes two Soldiers late.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "eb8ec34c-ae07-4a09-940f-ee965146a787",
		Name:         "Memorial to Glory",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W}",
			Label:    "Add {W}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{3}{W}, {T}, Sacrifice this land: Create two 1/1 white Soldier creature tokens.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{3}{W}"), TapCost(), SacrificeThis()),
			Effect:  Do(CreateToken{Template: TokenCard("1/1 white Soldier"), N: 2}),
		}},
	})
}
