package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "ce25dd89-dcc7-4576-b148-ef4c1cd5290c",
		Name:         "Fungal Colossus",
		Completeness: effects.CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLessEach(effects.TrekB5DifferentlyNamedLandsYouControl(),
				"This spell costs {X} less to cast, where X is the number of differently named lands you control."),
		},
	})
}
