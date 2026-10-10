package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Architect of the Untamed — Creature — Elf Artificer Druid {2}{G}, 2/3:
//
//	"Landfall — Whenever a land you control enters, you get {E} (an
//	 energy counter).
//	 Pay eight {E}: Create a 6/6 colorless Beast artifact creature
//	 token."
//
// ADR 0129 PR 1 (#1995).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "be6cab52-b3ae-46a2-a581-2468a49bff25",
		Name:         "Architect of the Untamed",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Landfall("Architect of the Untamed — you get {E}", ebYouGetEnergy(1)),
		},
		Activated: []ActivatedAbility{{
			Label:   "Pay eight {E}: Create a 6/6 colorless Beast artifact creature token.",
			Cost:    PayEnergy(8),
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker, Tokens: 1},
			Effect: Do(CreateToken{N: 1, Template: game.Card{
				Name:      "Beast",
				TypeLine:  "Token Artifact Creature — Beast",
				Power:     6,
				Toughness: 6,
			}}),
		}},
	})
}
