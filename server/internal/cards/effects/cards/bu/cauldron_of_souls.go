package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cauldron of Souls — Artifact {5}:
//
//	"{T}: Choose any number of target creatures. Each of those creatures
//	 gains persist until end of turn."
//
// Cauldron Haze on an artifact (#2075).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5811d905-cae5-4f92-a78e-abb43f8864a7",
		Name:         "Cauldron of Souls",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Choose any number of target creatures. Each of those creatures gains persist until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("any number of target creatures").WithCount(0, 0),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.GrantEachLegalTargetUntilEOT(effects.NewContext(g, item), game.KeywordPersist,
					"Cauldron of Souls — persist until end of turn")
			},
		}},
	})
}
