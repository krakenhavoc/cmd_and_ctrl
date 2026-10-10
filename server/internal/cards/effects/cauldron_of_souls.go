package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cauldron of Souls — Artifact {5}:
//
//	"{T}: Choose any number of target creatures. Each of those creatures
//	 gains persist until end of turn."
//
// Cauldron Haze on an artifact (#2075).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5811d905-cae5-4f92-a78e-abb43f8864a7",
		Name:         "Cauldron of Souls",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: Choose any number of target creatures. Each of those creatures gains persist until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerProtect},
			Cost:    TapCost(),
			Targets: TargetCreature("any number of target creatures").WithCount(0, 0),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return grantEachLegalTargetUntilEOT(NewContext(g, item), game.KeywordPersist,
					"Cauldron of Souls — persist until end of turn")
			},
		}},
	})
}
