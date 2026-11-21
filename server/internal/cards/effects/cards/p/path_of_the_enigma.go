package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Path of the Enigma — Sorcery {4}{U} (March of the Machine Commander,
// #5648):
//
//	"Target player draws four cards.
//	 Will of the Planeswalkers — Starting with you, each player votes for
//	 planeswalk or chaos. If planeswalk gets more votes, planeswalk. If
//	 chaos gets more votes or the vote is tied, chaos ensues."
//
// The spell has one target, so a target that has become illegal by the
// time it resolves means the whole spell does nothing, the vote included
// (CR 608.2b). Otherwise that player draws four, then the vote
// (WillOfThePlaneswalkers) waits for anything the draws ask (CR 608.2c).
// The vote happens at every table; planeswalking and chaos do something
// only at a table playing Planechase.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "602f3203-91ae-423c-8a19-c6c1a5034f27",
		Name:         "Path of the Enigma",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		Purpose:      effects.ForTargets(game.TargetPurpose{Slot: 0, Draws: 4}),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				item := ctx.Item
				if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer {
					return nil
				}
				return effects.DrawCards{Player: item.Targets[0].ID, N: 4}.Apply(ctx)
			}),
			effects.WillOfThePlaneswalkers{Card: "Path of the Enigma"},
		),
	})
}
