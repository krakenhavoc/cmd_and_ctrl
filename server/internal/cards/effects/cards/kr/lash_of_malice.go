package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lash of Malice — Instant {B} (Strixhaven: School of Mages, #5502):
//
//	"Target creature gets +2/-2 until end of turn."
//
// One until-end-of-turn boost with a negative toughness side, so it is a
// pump for an attacker or removal for a creature with 2 toughness or less
// (it dies to CR 704.5f).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fc3d2a9c-08a5-4aea-9fb8-42466b9152a2",
		Name:         "Lash of Malice",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return effects.BoostUntilEOT{Target: id, Power: 2, Toughness: -2, Label: "Lash of Malice — +2/-2"}.Apply(ctx)
		},
	})
}
