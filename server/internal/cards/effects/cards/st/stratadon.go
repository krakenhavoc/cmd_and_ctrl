package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stratadon — Artifact Creature — Beast {10}, 5/5 (Planeshift, S82 OLD b642, #5615):
//
//	"Domain — This spell costs {1} less to cast for each basic land type among
//	 lands you control.
//	 Trample"
//
// Yavimaya Sojourner's DomainCostReduction: generic mana only (CR 601.2f), so five
// types leave {5}. Trample is the printed keyword.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:          "3b78b672-7773-44c6-9a21-f89c058966ff",
		Name:              "Stratadon",
		Completeness:      effects.CompletenessFull,
		PrintedKeywords:   []string{"trample"},
		SelfCostModifiers: []game.CostModifier{effects.DomainCostReduction("Domain — This spell costs {1} less to cast for each basic land type among lands you control")},
	})
}
