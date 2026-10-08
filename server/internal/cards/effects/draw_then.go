package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// draw_then.go — "draw N cards, then X" (#2391, CR 608.2c).
//
// A clause printed after a draw must wait for the draw to resolve, and
// a draw can pause on a prompt (a dredge offer, a draw-instead pick, a
// CR 616 ordering prompt). Writing the clause as the statement after
// DrawCards runs it as soon as the call returns, with the draw still
// unanswered. Every "draw, then …" card therefore hands the rest of
// itself to Game.DrawNThenForEffect as a registered body plus scalars
// (game.DrawThen), which the engine runs once the last draw is done.
//
// The bodies here are stateless and capture only a label, so they sit
// in a registry keyed by what they capture, exactly as
// game.RegisterDrawInstead does.

// drawThenDiscardRef registers "then discard N" with a prompt
// question (empty for the default).
func drawThenDiscardRef(question string) game.DrawThenRef {
	return game.RegisterDrawThen("discard:"+question, func(g *game.Game, d game.DrawThen) error {
		g.QueueDiscardChoiceForEffect(game.DiscardPrompt{
			Player:   d.Player,
			Source:   d.Source,
			N:        d.N,
			Question: question,
		})
		return nil
	})
}

// drawThenDiscard is "<player> draws `draw` cards, then discards
// `discard` cards" — the discard prompt opens only after every draw has
// resolved, so the drawn cards are legal discards.
func drawThenDiscard(g *game.Game, player, source uuid.UUID, draw, discard int, question string) error {
	ref := drawThenDiscardRef(question)
	return g.DrawNThenForEffect(player, draw, game.DrawThen{Ref: ref, Player: player, Source: source, N: discard})
}

var (
	// "then, unless you attacked this turn, discard N" — Chart a Course.
	drawThenDiscardUnlessAttacked = game.RegisterDrawThen("discard-unless-attacked", func(g *game.Game, d game.DrawThen) error {
		if b18AttackedThisTurn(g, d.Player) {
			return nil
		}
		g.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: d.Player, Source: d.Source, N: d.N})
		return nil
	})

	// "then put two cards from your hand on top of your library in any
	// order" — Brainstorm.
	drawThenBrainstormPutBack = game.RegisterDrawThen("brainstorm-put-back", func(g *game.Game, d game.DrawThen) error {
		ctx := NewContext(g, &game.StackItem{Controller: d.Player, SourceCardID: d.Source})
		return PutFromHandOnTopInAnyOrder{
			Player: d.Player,
			N:      d.N,
			Label:  "Brainstorm — put two cards from your hand on top of your library",
		}.Apply(ctx)
	})

	// "then put a card from your hand on top of your library" — Enter
	// the Infinite.
	drawThenEnterTheInfinitePutBack = game.RegisterDrawThen("enter-the-infinite-put-back", func(g *game.Game, d game.DrawThen) error {
		ctx := NewContext(g, &game.StackItem{Controller: d.Player, SourceCardID: d.Source})
		return PutFromHandOnTopInAnyOrder{
			Player: d.Player,
			N:      d.N,
			Label:  "Enter the Infinite — put a card from your hand on top of your library",
		}.Apply(ctx)
	})
)
