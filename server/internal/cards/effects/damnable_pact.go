package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Damnable Pact — Sorcery {X}{B}{B} (EDHREC rank 3165):
//
//	"Target player draws X cards and loses X life."
//
// Sign in Blood at any size, pointed at anyone: the caster's own
// refill or the last X points of an opponent's life. The draw comes
// first, as printed, and the loss is a life change rather than
// damage, so it is not prevented and does not trigger damage
// payoffs. X is the announced value; X of zero does nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "28dd0ae4-fc56-4abf-af24-6c6c8d0a10cd",
		Name:         "Damnable Pact",
		XMatters:     true,
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer {
				return nil
			}
			x := ctx.X()
			if x <= 0 {
				return nil
			}
			victim := item.Targets[0].ID
			if err := (DrawCards{Player: victim, N: x}).Apply(ctx); err != nil {
				return err
			}
			return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), victim, -x)
		},
	})
}
