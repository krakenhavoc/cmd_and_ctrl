package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rapier Wit — Instant {1}{W} (Secrets of Strixhaven, #4456):
//
//	"Tap target creature. If it's your turn, put a stun counter on it. (If a
//	 permanent with a stun counter would become untapped, remove one from it
//	 instead.)
//	 Draw a card."
//
// "If it's your turn" is read as the spell resolves (the caster is the active
// player), and the stun counter goes on whether or not the creature was
// already tapped. With its one target gone the spell does nothing, not even
// the draw (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cf43518b-bfec-443d-a6c0-f2d848dc5fa8",
		Name:         "Rapier Wit",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			ts := ctx.LegalTargets()
			if len(ts) == 0 {
				return nil
			}
			target := ts[0].ID
			if err := (TapTarget{Target: target}).Apply(ctx); err != nil {
				return err
			}
			if IsYourTurn(ctx.Game, ctx.Controller()) {
				if err := (AddCounter{Target: target, Kind: game.CounterStun, N: 1}).Apply(ctx); err != nil {
					return err
				}
			}
			return DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
		},
	})
}
