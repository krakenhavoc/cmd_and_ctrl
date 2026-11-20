package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Specter of the Fens — Creature — Specter {3}{B}, 2/3 (Strixhaven: School
// of Mages, #5502):
//
//	"Flying
//	 {5}{B}: Target opponent loses 2 life and you gain 2 life."
//
// Flying is the printed keyword; the row is Cabal Archon's drain with an
// opponent as its target. The gain happens only when the opponent is still
// a legal target as the ability resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "048925e1-3bd6-4caa-90b9-e75e890032a5",
		Name:            "Specter of the Fens",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:   "{5}{B}: Target opponent loses 2 life and you gain 2 life.",
			Cost:    ManaCost("{5}{B}"),
			Targets: TargetPlayer("target opponent", Opponent()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return targetOpponentLosesAndYouGain(g, item, 2)
			},
		}},
	})
}
