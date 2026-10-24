package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tragic Trajectory — Sorcery {B} (Edge of Eternities):
//
//	"Target creature gets -2/-2 until end of turn.
//	 Void — That creature gets -10/-10 until end of turn instead if a
//	 nonland permanent left the battlefield this turn or a spell was
//	 warped this turn."
//
// Void (void.go) is read as the spell resolves and the size is then
// locked in, so a permanent leaving after the Trajectory resolves does not
// grow the penalty. The spell itself is not a permanent, and a warp cast
// before it counts. Tragic Slip's shape.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e5e19326-3e90-439c-a3b2-749d43a1af0c",
		Name:         "Tragic Trajectory",
		Purpose:      ForTargets(HarmUntilEOT(0, 0, 0)),
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			n, label := 2, "Tragic Trajectory — -2/-2"
			if Void(ctx.Game) {
				n, label = 10, "Tragic Trajectory — void, -10/-10"
			}
			return BoostUntilEOT{
				Target:    item.Targets[0].ID,
				Power:     -n,
				Toughness: -n,
				Label:     label,
			}.Apply(ctx)
		},
	})
}
