package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blight Pile — Creature — Phyrexian {1}{B}, 3/3 (Tarkir: Dragonstorm Commander):
//
//	"Defender
//	 {2}{B}, {T}: Each opponent loses X life, where X is the number of
//	 creatures with defender you control."
//
// X is counted as the ability resolves, the Pile included while it is on
// the battlefield.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1c0fc9b4-177e-4529-827f-a1e17ff7a752",
		Name:            "Blight Pile",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []ActivatedAbility{{
			Label:   "{2}{B}, {T}: Each opponent loses X life, where X is the number of creatures with defender you control.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{2}{B}"), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				n := cmdrB17DefendersControlled(g, item.Controller)
				if n <= 0 {
					return nil
				}
				return EachOpponentLosesLife{N: n}.Apply(NewContext(g, item))
			},
		}},
	})
}
