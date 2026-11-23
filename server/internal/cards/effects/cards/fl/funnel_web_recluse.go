package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Funnel-Web Recluse — Creature — Spider {4}{G}, 3/5 (Modern Horizons 2,
// #5501):
//
//	"Reach
//	 Morbid — When this creature enters, if a creature died this turn,
//	 investigate. (Create a Clue token. It's an artifact with "{2},
//	 Sacrifice this token: Draw a card.")"
//
// Wakedancer's morbid enter trigger (an intervening if, CR 603.4, over
// condACreatureDiedThisTurn: any creature, any controller) making a Clue.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6148fcf5-466a-43aa-a0f3-d3d9fae4d883",
		Name:            "Funnel-Web Recluse",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			effects.IfVentureCond(effects.CondACreatureDiedThisTurn,
				effects.WhenThisEnters("Funnel-Web Recluse — morbid: investigate",
					effects.Do(effects.Investigate{N: 1}))),
		},
	})
}
