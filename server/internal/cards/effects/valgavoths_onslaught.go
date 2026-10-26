package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Valgavoth's Onslaught — Sorcery {X}{X}{G}:
//
//	"Manifest dread X times, then put X +1/+1 counters on each of those
//	 creatures."
//
// X manifest dreads in a row, each its own look at the top two cards of
// the library as the earlier ones left it. "Each of those creatures" is
// the permanents that entered, so a manifest the library was too small
// to make, or one a replacement sent elsewhere, is not counted; a
// creature that left the battlefield before the counters go on is
// skipped.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "663d9559-6619-46fe-8ff4-655f1d5e0e4c",
		Name:         "Valgavoth's Onslaught",
		Completeness: CompletenessFull,
		XMatters:     true,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			x := ctx.X()
			if x <= 0 {
				return nil
			}
			t := ctx.Tail(valgavothsOnslaughtAgain)
			t.N, t.Word = x-1, strconv.Itoa(x)
			return ManifestDread{ThenTail: t}.Apply(ctx)
		},
	})
}

// valgavothsOnslaughtAgain is one manifest dread of the X (#4253 slice 9):
// N is how many are left, Carry the creatures manifested so far and Word
// X. After the last, X +1/+1 counters go on each of them still on the
// battlefield.
var valgavothsOnslaughtAgain game.TailRef

func init() {
	valgavothsOnslaughtAgain = ManifestDreadTail("valgavoths-onslaught/manifest-again", func(ctx *Context, res game.ManifestDreadResult, o game.TailOutcome) error {
		done := append([]uuid.UUID(nil), o.Tail.Carry...)
		if res.Manifested != uuid.Nil {
			done = append(done, res.Manifested)
		}
		if o.Tail.N > 0 {
			next := o.Tail
			next.N, next.Carry = o.Tail.N-1, done
			return ctx.Game.ManifestDreadKeyedForEffect(next.Player, next.Source, next)
		}
		x, _ := strconv.Atoi(o.Tail.Word)
		if x <= 0 {
			return nil
		}
		for _, id := range done {
			if !b15OnBattlefield(ctx.Game, id) {
				continue
			}
			if err := ctx.Game.AddCounterForEffect(id, game.CounterPlusOne, x); err != nil {
				return err
			}
		}
		return nil
	})
}
