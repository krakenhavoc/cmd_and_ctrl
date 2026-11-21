package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "048925e1-3bd6-4caa-90b9-e75e890032a5",
		Name:            "Specter of the Fens",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{5}{B}: Target opponent loses 2 life and you gain 2 life.",
			Cost:    effects.ManaCost("{5}{B}"),
			Targets: effects.TargetPlayer("target opponent", effects.Opponent()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.TargetOpponentLosesAndYouGain(g, item, 2)
			},
		}},
	})
}
