package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mistfire Weaver — Khans of Tarkir (S82 OLD b345, #5554):
//
//	Creature — Djinn Wizard {3}{U}, 3/1: Flying. Morph {2}{U}. When this creature is
//	turned face up, target creature you control gains hexproof until end of turn.

func init() {
	effects.Register(effects.Spec{
		OracleID:         "9af9c353-6b71-46e7-8815-a920e012dbb8",
		Name:             "Mistfire Weaver",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"flying"},
		AlternativeCosts: []game.AlternativeCost{effects.Morph("{2}{U}")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisIsTurnedFaceUp("Mistfire Weaver — target creature you control gains hexproof until end of turn",
					effects.EbTargetCreatureGainsUntilEOT("hexproof", "Mistfire Weaver — hexproof until end of turn")),
				effects.TargetCreature("target creature you control", effects.YouControl())),
		},
	})
}
