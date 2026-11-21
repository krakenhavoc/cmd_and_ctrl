package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Commander's Insight — Instant {X}{U}{U}{U}:
//
//	"Target player draws X cards plus an additional card for each time
//	 they've cast a commander from the command zone this game."
//
// Blue Sun's Zenith's draw with a rider. The rider is the TARGET player's
// commander-cast tally (Player.CommanderCasts, which only a cast from the
// command zone bumps), summed over every commander they have cast, read
// as the spell resolves: a player who has cast their commander twice
// draws X+2. A player who has never cast one draws X.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "54d7d7f8-22cd-4859-b203-924d248b422b",
		Name:         "Commander's Insight",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			who, ok := effects.FirstLegalPlayerTarget(ctx)
			if !ok {
				return nil
			}
			n := ctx.X() + effects.SocB3CommanderCastsFromCommandZone(ctx.Game, who)
			if n <= 0 {
				return nil
			}
			return effects.DrawCards{Player: who, N: n}.Apply(ctx)
		},
	})
}
