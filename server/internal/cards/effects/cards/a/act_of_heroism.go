package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Act of Heroism — Instant {1}{W}:
//
//	"Untap target creature. It gets +2/+2 until end of turn and can
//	 block an additional creature this turn."
//
// The untap, then ONE record carrying the +2/+2 (layer 7c) and the
// extra block (layer 6, #1715): the second sentence is one effect
// (CR 613.7). The extra block only matters if the spell resolves
// before blockers are declared — the usual play is to untap a creature
// that attacked last turn or was tapped for mana, and block two.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "25edb501-11b5-4617-8a65-2eb4869cccd5",
		Name:         "Act of Heroism",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			ts := ctx.LegalTargets()
			if len(ts) == 0 {
				return nil
			}
			id := ts[0].ID
			if err := (effects.UntapTarget{Target: id}).Apply(ctx); err != nil {
				return err
			}
			mods := append([]game.Mod{game.ModifyPTMod(2, 2)}, effects.BlockCapacityMods(1, false)...)
			return effects.UntilEndOfTurn(ctx, id, nil, "Act of Heroism — +2/+2, can block an additional creature", mods...)
		},
	})
}
