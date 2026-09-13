package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Peer into the Abyss — Sorcery {4}{B}{B}{B} (EDHREC rank 1272):
//
//	"Target player draws cards equal to half the number of cards in
//	 their library and loses half their life. Round up each time."
//
// Half a library and half a life total, both rounded up, both read
// at resolution. The draw comes first, as printed; the life loss is
// loss, not damage, so no prevention applies. A life total already at
// or below zero loses nothing (CR 107.1b — a negative loss is zero).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "21fa2442-6eac-4dce-a9cc-76f0053fdb8f",
		Name:         "Peer into the Abyss",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer {
				return nil
			}
			p := ctx.PlayerByID(item.Targets[0].ID)
			if p == nil || p.Eliminated {
				return nil
			}
			draw := 0
			if p.Library != nil {
				draw = (p.Library.Size() + 1) / 2
			}
			if err := (DrawCards{Player: p.ID, N: draw}).Apply(ctx); err != nil {
				return err
			}
			lose := (p.Life + 1) / 2
			if lose <= 0 {
				return nil
			}
			return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), p.ID, -lose)
		},
	})
}
