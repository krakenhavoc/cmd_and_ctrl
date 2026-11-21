package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spectral Lynx — Creature — Cat Spirit {1}{W}, 2/1 (Dominaria Remastered):
//
//	"Protection from green
//	 {B}: Regenerate this creature."
//
// Protection from green is the engine's (CR 702.16). The regeneration is
// Corrupted Harvester's: a shield for the Lynx, used up the next time it
// would be destroyed this turn (CR 701.19).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b5187725-9a39-433c-ae09-eccd6111574c",
		Name:            "Spectral Lynx",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.ProtectionFromColor("G")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}: Regenerate this creature.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.ManaCost("{B}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.Regenerate{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
