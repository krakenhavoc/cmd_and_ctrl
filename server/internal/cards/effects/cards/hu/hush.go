package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hush — Sorcery {3}{G} (Urza's Saga, S82 OLD b722):
//
//	Destroy all enchantments.
//	Cycling {2} ({2}, Discard this card: Draw a card.)
//
// Tranquility's sweep with cycling.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "eaa0c6c2-eee5-443a-867f-ab77b995b6ef",
		Name:         "Hush",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepEnchantments, How: game.SweepDestroy}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DestroyAllMatching{Match: effects.Enchantment()}.Apply(ctx)
		},
		Activated: []effects.ActivatedAbility{effects.Cycling("{2}")},
	})
}
