package cardsd

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

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
	effects.Register(effects.Spec{
		OracleID:     "28dd0ae4-fc56-4abf-af24-6c6c8d0a10cd",
		Name:         "Damnable Pact",
		XMatters:     true,
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: effects.ResolveSteps(
			// The life loss waits for a prompt the draws ask (CR 608.2c, #4155).
			effects.Instruction(func(ctx *effects.Context) error {
				victim, x, ok := damnablePactVictim(ctx)
				if !ok {
					return nil
				}
				return effects.DrawCards{Player: victim, N: x}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				victim, x, ok := damnablePactVictim(ctx)
				if !ok {
					return nil
				}
				return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), victim, -x)
			}),
		),
	})
}

// damnablePactVictim is the targeted player and X, or false when either is gone.
func damnablePactVictim(ctx *effects.Context) (uuid.UUID, int, bool) {
	item := ctx.Item
	if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer {
		return uuid.Nil, 0, false
	}
	x := ctx.X()
	if x <= 0 {
		return uuid.Nil, 0, false
	}
	return item.Targets[0].ID, x, true
}
