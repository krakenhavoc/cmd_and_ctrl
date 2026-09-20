package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Read the Runes — Instant {X}{U}:
//
//	"Draw X cards. For each card drawn this way, discard a card unless
//	 you sacrifice a permanent."
//
// An instant-speed Braingeyser you pay for twice: once in mana and
// once, a card or a permanent at a time, out of the board you already
// had. In a deck full of Treasures, Clues and tokens the second price
// is nearly free, which is the whole reason the card is played.
//
// # It is not an additional cost
//
// Every clause here happens on RESOLUTION. Nothing is paid to cast it,
// so countering it costs you nothing, and a board that emptied in
// response changes what you can pay with. Written as an
// `AdditionalCost` the card would be uncastable with an empty board
// and would pay before the draw — both wrong, and both observable.
//
// # X separate two-option choices, asked one at a time
//
// "For each card drawn this way" is a repetition, and each repetition
// is its own choice: pitch a card or sacrifice a permanent. They may
// be answered differently, and they are strictly sequential — every
// question is queued by the ANSWER to the one before it, so a player
// who sacrificed their last token to repetition two is offered only
// the discard in repetition three. That is Torment of Hailfire's
// chain (#568) pointed at its own controller, and it is why this uses
// PickOption + SacrificeChoice rather than the seams the two earlier
// triages filed it under ("variable-count sacrifice cost",
// "resolution-time choose N of your own permanents"). Neither is
// needed: the count is not a cost and the choice is one permanent at
// a time.
//
// "Discard a card" is FIRST and is always offered, which is both the
// printed default ("discard a card UNLESS…") and the option-pick
// kind's own contract — the enumerator marks the first branch
// always-legal, so it has to be one that never fails. It never does:
// a player with an empty hand discards as many as they can, which is
// none (CR 701.8a). "Sacrifice a permanent" is offered only when
// there is one, per CR 608.2's "as much as possible", and a permanent
// means ANY permanent you control, lands included.
//
// # The repetition count
//
// "For each card drawn THIS WAY" counts cards actually drawn, so a
// library shorter than X is asked about only the cards it had. The
// count is read off the library before the draw; a draw REPLACEMENT
// that turned a draw into something else would not be seen, which is
// the one place this is approximate and is a strictly smaller number
// than X in every case it is wrong about.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "74316be1-041f-4a2e-a296-0b496f90ca25",
		Name:         "Read the Runes",
		XMatters:     true,
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return readTheRunesResolve(ctx)
		},
	})
}

// The two branch keys. Plain strings rather than closures so the
// option list and the branch list cannot drift: the index the chooser
// answers with indexes both.
const (
	readTheRunesDiscard   = "discard"
	readTheRunesSacrifice = "sacrifice"
)

// readTheRunesResolve draws, then starts the chain of questions.
//
// Caller holds g.mu.
func readTheRunesResolve(ctx *Context) error {
	x := ctx.X()
	if x <= 0 {
		return nil
	}
	drawn := x
	if p := ctx.PlayerByID(ctx.Controller()); p != nil && p.Library != nil && p.Library.Size() < drawn {
		drawn = p.Library.Size()
	}
	if err := (DrawCards{N: x}).Apply(ctx); err != nil {
		return err
	}
	return readTheRunesStep(ctx, drawn)
}

// readTheRunesStep asks the question owed for one drawn card and
// hands the remaining count to its answer.
//
// A package-level function over an int: no *Game and no player
// pointer is captured, so an undo that replays an answer resolves it
// against the restored game.
//
// Caller holds g.mu.
func readTheRunesStep(ctx *Context, remaining int) error {
	if remaining <= 0 {
		return nil
	}
	options := []game.ChoiceOption{{Label: "Discard a card"}}
	branches := []string{readTheRunesDiscard}
	if len(PermanentsControlledBy(ctx.Game, ctx.Controller())) > 0 {
		options = append(options, game.ChoiceOption{Label: "Sacrifice a permanent"})
		branches = append(branches, readTheRunesSacrifice)
	}
	return PickOption{
		Question: "Read the Runes — discard a card unless you sacrifice a permanent",
		Options:  options,
		Then:     readTheRunesAnswered(branches, remaining-1),
	}.Apply(ctx)
}

// readTheRunesAnswered runs the chosen branch and then asks the next
// question. Everything it captures is a scalar or a frozen slice.
func readTheRunesAnswered(branches []string, rest int) func(ctx *Context, index int) error {
	return func(ctx *Context, index int) error {
		next := func(ctx *Context) error { return readTheRunesStep(ctx, rest) }
		if index >= 0 && index < len(branches) && branches[index] == readTheRunesSacrifice {
			return SacrificeChoice{
				Player:     ctx.Controller(),
				Candidates: PermanentsControlledBy(ctx.Game, ctx.Controller()),
				Question:   "Read the Runes — sacrifice a permanent",
				Then:       next,
			}.Apply(ctx)
		}
		// The discard, and the "no question could be asked" case,
		// which is the same outcome: an index the branch list does
		// not have means the chooser left between the question being
		// built and it being queued, and the rest of the run still
		// happens.
		player := ctx.Controller()
		if p := ctx.PlayerByID(player); p == nil || p.Hand == nil || p.Hand.Size() == 0 {
			return next(ctx)
		}
		item := ctx.Item
		ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{
			Player:   player,
			Source:   ctx.Source(),
			N:        1,
			Question: "Read the Runes — discard a card",
			Then: func(g *game.Game, _ uuid.UUID, _ []uuid.UUID) error {
				return readTheRunesStep(NewContext(g, item), rest)
			},
		})
		return nil
	}
}
