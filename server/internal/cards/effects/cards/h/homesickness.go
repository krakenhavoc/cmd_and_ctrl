package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Homesickness — Instant {4}{U}{U} (Secrets of Strixhaven, #4456):
//
//	"Target player draws two cards. Tap up to two target creatures. Put a
//	 stun counter on each of them."
//
// Two target clauses (Clauses, read by slot): the player and up to two
// creatures. Each half is re-checked at resolution and skipped when its
// target has left or is no longer legal (CR 608.2b), so choosing no creature
// still draws and a vanished player still lets the creatures be stunned.
// The stun counter goes on a creature whether or not it was untapped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ee57fba8-e68b-4c52-8de7-9b9873d84506",
		Name:         "Homesickness",
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.TargetPlayer("target player"),
			effects.TargetCreature("up to two target creatures").WithCount(0, 2),
		),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if t, ok := ctx.ClauseTarget(0); ok && t.Kind == game.TargetPlayer {
				if err := (effects.DrawCards{Player: t.ID, N: 2}).Apply(ctx); err != nil {
					return err
				}
			}
			for _, t := range ctx.ClauseTargets(1) {
				if !ctx.IsTargetLegal(t) || !effects.OnBattlefield(ctx.Game, t.ID) {
					continue
				}
				if err := (effects.TapTarget{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
				if err := (effects.AddCounter{Target: t.ID, Kind: game.CounterStun, N: 1}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
