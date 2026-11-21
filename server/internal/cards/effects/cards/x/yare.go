package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Yare — Instant {2}{W}:
//
//	"Target creature defending player controls gets +3/+0 until end of
//	 turn. That creature can block up to two additional creatures this
//	 turn."
//
// "Defending player" is read as Commander reads it: every opponent of
// the active player, for the whole combat phase (CR 802.2, the
// attack-multiple-players option) — ControlledByDefendingPlayer. So
// Yare has no legal target outside combat. The +3/+0 and "up to two
// additional" are two sentences and two records (#1715); "up to" is a
// capacity, not an obligation, so the creature may still block one,
// two or three.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f0f24c6d-80fb-4e99-a74d-6ed9d349ed3d",
		Name:         "Yare",
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 3, 0)),
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature defending player controls", effects.ControlledByDefendingPlayer()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			ts := ctx.LegalTargets()
			if len(ts) == 0 {
				return nil
			}
			id := ts[0].ID
			if err := (effects.BoostUntilEOT{Target: id, Power: 3, Label: "Yare — +3/+0"}).Apply(ctx); err != nil {
				return err
			}
			return effects.BlockCapacityUntilEOT{
				Target:     id,
				Additional: 2,
				Label:      "Yare — can block up to two additional creatures",
			}.Apply(ctx)
		},
	})
}
