package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cryptic Serpent — Creature — Serpent {5}{U}{U}, 6/5
// (Amonkhet, #5535):
//
//	"This spell costs {1} less to cast for each instant and sorcery card in your graveyard."
//
// Eddymurk Crab's cost reduction.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0f26c04a-9bf7-4cbd-a95f-73ee4a3b1af1",
		Name:         "Cryptic Serpent",
		Completeness: effects.CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLessEach(effects.SocB37InstantAndSorceryCardsInYourGraveyard,
				"This spell costs {1} less to cast for each instant and sorcery card in your graveyard."),
		},
	})
}
