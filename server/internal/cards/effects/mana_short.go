package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mana Short — Instant {2}{U}:
//
//	"Tap all lands target player controls and that player loses all unspent
//	 mana."
//
// Nature's Will's tap-every-land walk, then the whole pool is lost through
// LoseAllUnspentManaForEffect (CR 106.4). A player who left before
// resolution is an illegal target (CR 608.2b) and nothing happens.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "48207d1c-448a-4e1b-974a-642dfea75933",
		Name:         "Mana Short",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			target, ok := firstLegalPlayerTarget(ctx)
			if !ok {
				return nil
			}
			if err := b24TapAllLandsControlledBy(ctx, target); err != nil {
				return err
			}
			ctx.Game.LoseAllUnspentManaForEffect(target)
			return nil
		},
	})
}
