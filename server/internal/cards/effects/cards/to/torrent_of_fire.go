package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Torrent of Fire — Sorcery {3}{R}{R} (Scourge):
//
//	"Torrent of Fire deals damage to any target equal to the greatest mana
//	 value among permanents you control."
//
// The amount is read as the spell resolves (scgB1GreatestManaValue), and
// the spell deals it to its one target. A target that left in response
// takes nothing (CR 608.2b); a greatest mana value of 0 deals no damage.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "00181784-9213-492a-ba8c-2028969b049e",
		Name:         "Torrent of Fire",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			n := effects.ScgB1GreatestManaValue(ctx)
			if n <= 0 {
				return nil
			}
			targets := ctx.LegalTargets()
			if len(targets) == 0 {
				return nil
			}
			return effects.DealDamage{Source: item.SourceCardID, Target: targets[0].ID, Amount: n}.Apply(ctx)
		},
	})
}
