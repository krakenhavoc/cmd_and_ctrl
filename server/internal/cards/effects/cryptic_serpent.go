package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cryptic Serpent — Creature — Serpent {5}{U}{U}, 6/5
// (Amonkhet, #5535):
//
//	"This spell costs {1} less to cast for each instant and sorcery card in your graveyard."
//
// Eddymurk Crab's cost reduction.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0f26c04a-9bf7-4cbd-a95f-73ee4a3b1af1",
		Name:         "Cryptic Serpent",
		Completeness: CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			CostsLessEach(socB37InstantAndSorceryCardsInYourGraveyard,
				"This spell costs {1} less to cast for each instant and sorcery card in your graveyard."),
		},
	})
}
