package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Siren Reaver — Creature — Siren Pirate {3}{U}, 3/2 (Rivals of Ixalan, S82
// OLD batch 218, #5531):
//
//	"Raid — This spell costs {1} less to cast if you attacked this turn.
//	 Flying"
//
// A self cost reduction (CR 601.2f) read as the cost is determined: it applies
// only to a player who has declared an attack this turn, and takes the generic
// {1} off the {3}{U}.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "bc1c9ca8-2f00-4761-bd85-1329287ae7ac",
		Name:            "Siren Reaver",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		SelfCostModifiers: []game.CostModifier{
			effects.RaidCostsLess(1, "Raid — This spell costs {1} less to cast if you attacked this turn."),
		},
	})
}
