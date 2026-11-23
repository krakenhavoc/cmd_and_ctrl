package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Play with Fire — Instant {R} (Innistrad: Midnight Hunt, #5497):
//
//	"Play with Fire deals 2 damage to any target. If a player is dealt
//	 damage this way, scry 1."
//
// The scry is the damage's continuation told how much a player was
// actually dealt (Chain Lightning's seam), so damage that is prevented
// entirely scrys nothing, and a creature or planeswalker never does.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5ffc1e77-152b-4a82-8ea8-2c79cc31d5ef",
		Name:         "Play with Fire",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		Purpose:      game.Purpose{Scry: 1, Targets: game.ForTargets(effects.DamageToTarget(0, 2)), Answers: game.AnswerRemove},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			ts := ctx.LegalTargets()
			if len(ts) == 0 {
				return nil
			}
			t := ts[0]
			if ctx.Game.PlayerByIDForEffect(t.ID) != nil {
				return ctx.Game.DealDamageToPlayerKeyedForEffect(ctx.Source(), t.ID, 2, effects.WithN(ctx.Tail(playWithFireScry), 1))
			}
			return effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 2}.Apply(ctx)
		},
	})
}

// playWithFireScry is "if a player is dealt damage this way, scry 1": the
// tail is told the damage dealt. An on-disk key.
var playWithFireScry = effects.CardTail("then/play-with-fire-scry", func(ctx *effects.Context, o game.TailOutcome) error {
	if o.Amount <= 0 {
		return nil
	}
	return effects.Scry{Player: o.Tail.Player, N: o.Tail.N}.Apply(ctx)
})
