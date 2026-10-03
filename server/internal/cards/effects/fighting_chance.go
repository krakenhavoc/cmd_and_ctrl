package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fighting Chance — Instant {R}:
//
//	"For each blocking creature, flip a coin. If you win the flip, prevent all combat damage that would be dealt by that creature this turn."
//
// ADR 0108 §7 (#1904, Delivery PR 7b) and ADR 0054's coin flips: the
// blocking creatures are read as the spell resolves, one coin is flipped
// for each, in battlefield order, and each won flip gives that blocker a
// combat-damage shield with itself as the source, pinned now (CR 400.7).
// The caster calls once for the batch; every coin is still its own fair
// flip, won or lost on its own. No blocker, no flip.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "fece632b-d8d5-4d3d-a4a7-643293b67538",
		Name:         "Fighting Chance",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			blockers := blockingCreatures(ctx.Game)
			if len(blockers) == 0 {
				return nil
			}
			ctx.Game.FlipCoinForEffect(game.CoinFlipSpec{
				Flipper:  ctx.Controller(),
				Source:   ctx.Source(),
				Question: "Fighting Chance — call the coin flips (one per blocking creature)",
				Coins:    len(blockers),
				Then: func(g *game.Game, result game.CoinFlipResult) error {
					return fightingChanceShields(NewContext(g, item), blockers, result)
				},
			})
			return nil
		},
	})
}

// fightingChanceShields shields each blocker whose flip was won.
func fightingChanceShields(ctx *Context, blockers []uuid.UUID, result game.CoinFlipResult) error {
	for i, won := range result.Won {
		if !won || i >= len(blockers) {
			continue
		}
		if err := (PreventDamageFromSource{From: blockers[i], CombatOnly: true, Protect: ShieldAnything}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
