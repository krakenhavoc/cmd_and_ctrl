package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tower Defense — Instant {1}{G} (Tarkir: Dragonstorm Commander):
//
//	"Creatures you control get +0/+5 and gain reach until end of turn."
//
// Atarka's Command's last mode, with +0/+5. The set is locked in as the
// spell resolves (CR 611.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "58853200-efbe-41cc-87f1-720841308f4d",
		Name:         "Tower Defense",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			yours := effects.And(effects.Creature(), effects.YouControl())
			if err := (effects.BoostUntilEOT{Match: yours, Power: 0, Toughness: 5, Label: "Tower Defense — +0/+5"}).Apply(ctx); err != nil {
				return err
			}
			return effects.GrantKeywordUntilEOT{Match: yours, Keywords: []string{"reach"}, Label: "Tower Defense — reach"}.Apply(ctx)
		},
	})
}
