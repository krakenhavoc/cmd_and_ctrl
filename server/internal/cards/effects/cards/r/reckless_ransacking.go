package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Reckless Ransacking — Instant {1}{R} (Lorwyn Eclipsed, #4458):
//
//	"Target creature gets +3/+2 until end of turn. Create a Treasure
//	 token."
//
// The pump is a turn-long boost on the target as it is now, then the
// Treasure is the controller's. With the target illegal as the spell
// resolves nothing happens, and no Treasure is made (the ruling,
// CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ebcd198e-7fbd-46e0-a6d9-ff499e541978",
		Name:         "Reckless Ransacking",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 3, 2)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (effects.BoostUntilEOT{Target: id, Power: 3, Toughness: 2,
				Label: "Reckless Ransacking — +3/+2 until end of turn"}).Apply(ctx); err != nil {
				return err
			}
			return effects.CreateToken{Template: effects.TreasureToken(), N: 1}.Apply(ctx)
		},
	})
}
