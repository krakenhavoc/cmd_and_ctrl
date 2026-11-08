package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fungal Colossus — Creature — Fungus Beast {6}{G}, 5/5 (Edge of
// Eternities, #4461):
//
//	"This spell costs {X} less to cast, where X is the number of
//	 differently named lands you control."
//
// Cloak and Phaser's self cost modifier on a creature spell: each land
// the caster controls is counted once per name as the cost is determined
// (CR 601.2f), and it reduces generic mana only, so {G} is always paid.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ce25dd89-dcc7-4576-b148-ef4c1cd5290c",
		Name:         "Fungal Colossus",
		Completeness: CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			CostsLessEach(trekB5DifferentlyNamedLandsYouControl(),
				"This spell costs {X} less to cast, where X is the number of differently named lands you control."),
		},
	})
}
