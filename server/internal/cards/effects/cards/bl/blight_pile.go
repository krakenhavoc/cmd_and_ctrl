package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "1c0fc9b4-177e-4529-827f-a1e17ff7a752",
		Name:            "Blight Pile",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{B}, {T}: Each opponent loses X life, where X is the number of creatures with defender you control.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}{B}"), effects.TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				n := effects.CmdrB17DefendersControlled(g, item.Controller)
				if n <= 0 {
					return nil
				}
				return effects.EachOpponentLosesLife{N: n}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
