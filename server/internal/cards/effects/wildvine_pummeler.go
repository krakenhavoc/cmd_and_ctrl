package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wildvine Pummeler — Creature — Giant Berserker {6}{G} (Lorwyn Eclipsed, #4458):
//
//	"Vivid — This spell costs {1} less to cast for each color among
//	 permanents you control.
//	 Reach, trample"
func init() {
	Register(Spec{
		OracleID:        "6c34674f-856b-4062-87d5-6ed73d9a9855",
		Name:            "Wildvine Pummeler",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach", "trample"},
		SelfCostModifiers: []game.CostModifier{
			CostsLessEach(func(q game.CostQuery) int {
				return len(colorsAmongPermanentsYouControl(q.Game, q.Controller))
			}, "Vivid — This spell costs {1} less to cast for each color among permanents you control."),
		},
	})
}
