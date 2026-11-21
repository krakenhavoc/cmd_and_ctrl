package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "48207d1c-448a-4e1b-974a-642dfea75933",
		Name:         "Mana Short",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			target, ok := effects.FirstLegalPlayerTarget(ctx)
			if !ok {
				return nil
			}
			if err := effects.B24TapAllLandsControlledBy(ctx, target); err != nil {
				return err
			}
			ctx.Game.LoseAllUnspentManaForEffect(target)
			return nil
		},
	})
}
