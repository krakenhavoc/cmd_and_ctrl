package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tranquil Domain — Instant {1}{G} (Mirage):
//
//	"Destroy all non-Aura enchantments."
//
// Tranquility with the Auras spared: every enchantment that is not an Aura,
// whoever controls it, is destroyed at once (one event, CR 700.4). An
// enchantment creature or an enchantment artifact that is not an Aura goes
// too. No regeneration clause, so a shield saves one. Declared as a partial
// enchantment sweep.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "87547bf0-02cf-4d83-9a1f-c82d2f1a22f1",
		Name:         "Tranquil Domain",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepEnchantments, How: game.SweepDestroy, Partial: true}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DestroyAllMatching{Match: effects.And(effects.Enchantment(), effects.Not(effects.Subtype("Aura")))}.Apply(ctx)
		},
	})
}
