package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hypothesizzle — Instant {3}{U}{R}:
//
//	"Draw two cards. Then you may discard a nonland card. When you do,
//	 Hypothesizzle deals 4 damage to target creature."
//
// Glacial Dragonhunt's loot over two draws, with the discard narrowed to
// a nonland card by the set rule (Dokuchi Silencer's shape): a land is
// refused, so the reflexive trigger (CR 603.12) is created only by a
// nonland discard. Its target is chosen as it goes on the stack, and the
// damage comes from Hypothesizzle as it last existed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e4e6cb61-f673-4f36-b95c-c84a68459503",
		Name:         "Hypothesizzle",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 2, Discards: 1},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return ctx.Game.DrawNThenForEffect(item.Controller, 2, game.DrawThen{
				Ref: hypothesizzleDiscard, Player: item.Controller, Source: item.SourceCardID, N: 1,
			})
		},
	})
}

var (
	hypothesizzleDiscard = game.RegisterDrawThen("hypothesizzle/may-discard-nonland", func(g *game.Game, d game.DrawThen) error {
		return g.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
			Player:   d.Player,
			Source:   d.Source,
			N:        1,
			UpTo:     true,
			Question: "Hypothesizzle — you may discard a nonland card",
		}, hypothesizzleDiscarded, nil)
	})

	hypothesizzleDiscarded = game.RegisterDiscardThenRule("choose-cards/hypothesizzle-discarded",
		func(picked []game.Card) bool { return len(picked) == 1 && !picked[0].IsLand() },
		func(g *game.Game, r game.DiscardsDone) error {
			if len(r.Discarded.Cards()) == 0 {
				return nil
			}
			t := WhenYouDo("Hypothesizzle — 4 damage to target creature", hypothesizzleDamageBody)
			t.Params = game.EffectParams{Amount: 4}
			return t.Apply(keyedContext(g, r.Player, r.Source))
		})

	hypothesizzleDamageBody = game.ReflexiveBody("hypothesizzle/damage", breechesBlastEffect,
		constTargets(func() *game.TargetSpec { return TargetCreature("target creature") }))
)
