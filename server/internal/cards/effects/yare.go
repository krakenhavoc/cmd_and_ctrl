package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "f0f24c6d-80fb-4e99-a74d-6ed9d349ed3d",
		Name:         "Yare",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature defending player controls", ControlledByDefendingPlayer()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			ts := ctx.LegalTargets()
			if len(ts) == 0 {
				return nil
			}
			id := ts[0].ID
			if err := (BoostUntilEOT{Target: id, Power: 3, Label: "Yare — +3/+0"}).Apply(ctx); err != nil {
				return err
			}
			return BlockCapacityUntilEOT{
				Target:     id,
				Additional: 2,
				Label:      "Yare — can block up to two additional creatures",
			}.Apply(ctx)
		},
	})
}
