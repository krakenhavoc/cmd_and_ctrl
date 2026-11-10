package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hoarder's Greed — Sorcery {3}{B} (Lorwyn Eclipsed Commander, #5424):
//
//	"You lose 2 life and draw two cards, then clash with an opponent. If
//	 you win, repeat this process. (Each clashing player reveals the top
//	 card of their library, then puts that card on their choice of the top
//	 or bottom. A player wins if their card had a greater mana value.)"
//
// "This process" is the whole sentence: the life, the two cards, a new
// choice of opponent (CR 701.30b: choosing is part of clashing with an
// opponent) and the clash. It is one chain of registered continuations,
// run again from the top each time the clash is won (CR 701.30d), so each
// step waits for a prompt the one before it asked: a replacement on the
// life loss or a draw, the choice of opponent, and both players'
// top-or-bottom decisions (CR 701.30c). A tie, a loss, or a library of
// yours with no card to reveal ends it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "60040023-7bf0-42e0-ab30-46a730675564",
		Name:         "Hoarder's Greed",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 2},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return hoardersGreedProcess(ctx.Game, item.Controller, item.SourceCardID)
		},
	})
}

// hoardersGreedProcess starts the process once: "you lose 2 life", with
// the rest behind it.
func hoardersGreedProcess(g *game.Game, controller, source uuid.UUID) error {
	return g.LoseLifeEachKeyedForEffect(source, []uuid.UUID{controller}, 2,
		game.KeyedTail{Ref: hoardersGreedLost, Player: controller, Source: source})
}

// hoardersGreedLost is "and draw two cards". The keys below are on-disk
// identities: never renamed, never reused.
var hoardersGreedLost = game.RegisterTail("hoarders-greed/draw-two", func(g *game.Game, o game.TailOutcome) error {
	return g.DrawNThenForEffect(o.Tail.Player, 2, game.DrawThen{
		Ref: hoardersGreedDrawn, Player: o.Tail.Player, Source: o.Tail.Source,
	})
})

// hoardersGreedDrawn is "then clash with an opponent".
var hoardersGreedDrawn = game.RegisterDrawThen("hoarders-greed/clash", func(g *game.Game, d game.DrawThen) error {
	return g.ClashWithAnOpponentForEffect(d.Source, d.Player,
		game.KeyedTail{Ref: hoardersGreedClashed, Player: d.Player, Source: d.Source})
})

// hoardersGreedClashed is "If you win, repeat this process."
var hoardersGreedClashed = game.RegisterTail("hoarders-greed/repeat-if-won", func(g *game.Game, o game.TailOutcome) error {
	if !wonTheClash(o) {
		return nil
	}
	return hoardersGreedRepeat(g, o.Tail.Player, o.Tail.Source)
})

// hoardersGreedRepeat is hoardersGreedProcess, reached through a variable
// so the package's initialisation order has no cycle (the process's last
// step starts the process).
var hoardersGreedRepeat func(g *game.Game, controller, source uuid.UUID) error

func init() { hoardersGreedRepeat = hoardersGreedProcess }
