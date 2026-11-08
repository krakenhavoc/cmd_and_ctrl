package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Starscape Cleric — Creature — Bat Cleric {1}{B}, 2/1 (Bloomburrow,
// #4467):
//
//	"Offspring {2}{B} (You may pay an additional {2}{B} as you cast this
//	 spell. If you do, when this creature enters, create a 1/1 token copy
//	 of it.)
//	 Flying
//	 This creature can't block.
//	 Whenever you gain life, each opponent loses 1 life."
//
// Offspring is the shared keyword (offspring.go), and the token copy
// flies, can't block and drains too. "Can't block" is Mudbutton
// Cursetosser's self restriction, checked as blockers are declared
// (CR 509.1b). The drain is Cliffhaven Vampire's: one trigger per
// life-gain event, whatever the amount.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8d72ddb3-ef6e-4847-b6d4-5d85b1dd5802",
		Name:            "Starscape Cleric",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		OptionalCosts:   []game.AdditionalCost{Offspring("{2}{B}")},
		Static: []game.StaticAbility{
			RestrictSelf(game.CantBlock),
		},
		Triggered: []game.TriggeredAbility{
			OffspringToken("Starscape Cleric"),
			WheneverYouGainLife("Starscape Cleric — each opponent loses 1 life", Do(EachOpponentLosesLife{N: 1})),
		},
	})
}
