package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
// none (CR 701.9a). "Sacrifice a permanent" is offered only when
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
		// The first question waits for a prompt the draw asks (a dredge offer; CR 608.2c, #4155).
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error { return readTheRunesDraw(ctx) }),
			Instruction(func(ctx *Context) error {
				return readTheRunesStep(ctx, ctx.StepRecord(readTheRunesRecord).N())
			}),
		),
	})
}

// readTheRunesRecord is how many cards the draw puts in the hand, the number
// of questions owed.
const readTheRunesRecord = "read-the-runes/drawn"

// readTheRunesDraw draws X cards and files how many really come, for the
// chain of questions that follows as the next step.
//
// Caller holds g.mu.
func readTheRunesDraw(ctx *Context) error {
	x := ctx.X()
	if x <= 0 {
		ctx.RecordStep(readTheRunesRecord, nil, 0)
		return nil
	}
	drawn := x
	if p := ctx.PlayerByID(ctx.Controller()); p != nil && p.Library != nil && p.Library.Size() < drawn {
		drawn = p.Library.Size()
	}
	ctx.RecordStep(readTheRunesRecord, nil, drawn)
	return DrawCards{N: x}.Apply(ctx)
}

// readTheRunesLeft is how many questions are still owed, the one being asked
// included. It is filed before each question, so the answer (which runs
// after a restore, from a registered body) reads how many follow it.
const readTheRunesLeft = "read-the-runes/left"

// readTheRunesStep asks the question owed for one drawn card; its answer asks
// the next, so the count is filed with the resolution (readTheRunesLeft).
//
// Caller holds g.mu.
func readTheRunesStep(ctx *Context, remaining int) error {
	if remaining <= 0 {
		return nil
	}
	ctx.RecordStep(readTheRunesLeft, nil, remaining)
	options := []game.ChoiceOption{{Label: "Discard a card"}}
	if len(PermanentsControlledBy(ctx.Game, ctx.Controller())) > 0 {
		options = append(options, game.ChoiceOption{Label: "Sacrifice a permanent"})
	}
	return PickOption{
		Question: "Read the Runes — discard a card unless you sacrifice a permanent",
		Options:  options,
		ThenKey:  readTheRunesAnswered,
		WithItem: true,
	}.Apply(ctx)
}

// readTheRunesAfter asks the next question once this one is settled.
func readTheRunesAfter(ctx *Context) error {
	return readTheRunesStep(ctx, ctx.StepRecord(readTheRunesLeft).N()-1)
}

// The three registered continuations of the chain (#4253), each an on-disk
// key (never renamed, never reused) registered in init because the chain asks
// again from inside them.
var (
	// readTheRunesAnswered is the answer to a question: index 1 is the
	// sacrifice, offered only when there is a permanent to sacrifice.
	readTheRunesAnswered game.OptionPickThen
	// readTheRunesSacrificed is the sacrifice gone, and the next question.
	readTheRunesSacrificed game.TailRef
	// readTheRunesDiscarded is the discard made, and the next question.
	readTheRunesDiscarded game.DiscardThen
)

func init() {
	readTheRunesAnswered = OptionPickThen("option-pick/read-the-runes-answer", func(ctx *Context, r game.OptionPicked) error {
		return readTheRunesBranch(ctx, r.Index == 1)
	})
	readTheRunesSacrificed = CardTail("read-the-runes/sacrificed", func(ctx *Context, _ game.TailOutcome) error {
		return readTheRunesAfter(ctx)
	})
	readTheRunesDiscarded = game.RegisterDiscardThen("choose-cards/read-the-runes-discarded", func(g *game.Game, r game.DiscardsDone) error {
		ctx, _ := askedContext(g, r.Item, r.Player, r.Source, nil)
		return readTheRunesAfter(ctx)
	})
}

// readTheRunesBranch is the answer's branch: the sacrifice when it was taken,
// otherwise the discard, which is also the "no question could be asked" case:
// an index the option list does not have means the chooser left between the
// question being built and it being queued, and the rest of the run still
// happens.
func readTheRunesBranch(ctx *Context, sacrifice bool) error {
	if !sacrifice {
		return readTheRunesDiscardCard(ctx)
	}
	return SacrificeChoice{
		Player:     ctx.Controller(),
		Candidates: PermanentsControlledBy(ctx.Game, ctx.Controller()),
		Question:   "Read the Runes — sacrifice a permanent",
		ThenTail:   ctx.ItemTail(readTheRunesSacrificed),
	}.Apply(ctx)
}

// readTheRunesDiscardCard is the discard branch: the prompt, or the next
// question at once when the hand is empty. The branches are exclusive, so it
// is a function of its own rather than the tail of the sacrifice's `if`.
func readTheRunesDiscardCard(ctx *Context) error {
	player := ctx.Controller()
	if p := ctx.PlayerByID(player); p == nil || p.Hand == nil || p.Hand.Size() == 0 {
		return readTheRunesAfter(ctx)
	}
	return ctx.Game.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
		Player:   player,
		Source:   ctx.Source(),
		N:        1,
		Question: "Read the Runes — discard a card",
		Asked:    game.NewAskedItem(ctx.Item),
	}, readTheRunesDiscarded, nil)
}
