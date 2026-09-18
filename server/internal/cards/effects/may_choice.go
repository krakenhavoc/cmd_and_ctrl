package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// may_choice.go — the free yes/no a RESOLVING effect asks (#796,
// CR 608.2).
//
// # The gap this closes
//
// Three yes/no prompts already existed and none of them was this one.
// A trigger's CR 603.5 "you may" is asked before the ability goes on
// the stack, by the harvester, and is over by the time anything
// resolves. `PayUnless` / `MayPay` are asked about a MANA payment and
// speak the client's "Pay {2}" vocabulary. Search, scry and
// put-from-hand each carry their own decline because declining is part
// of that instruction.
//
// What nothing could ask is the plain shape "you may [do X]. If you
// do, [Y]" in the middle of a resolution, where X is neither a search
// nor a cost the engine already prompts for:
//
//   - Eden, Seat of the Sanctum — "Mill two cards. Then you may
//     sacrifice this land." The decision comes AFTER the mill, which
//     is the whole of why the card is played.
//   - Combustible Gearhulk — "target opponent MAY have you draw three
//     cards", asked of a seat that is not the controller.
//   - Charismatic Conqueror, Painful Quandary — the same, asked of the
//     player an opponent's trigger is about.
//
// # No new prompt kind
//
// It is built on PendingChoiceConfirm (game/chained_choice.go), which
// is already exactly this: a two-way question whose branches are plain
// continuations supplied by the card, addressed by Chooser, classified
// in the choice gate, enumerated by `internal/legal`, and rendered by
// the client modal. A `may` kind would have been a fourth spelling of
// a question the queue can already ask, and every consumer would have
// needed a case for it.
//
// What this file adds is the CARD-FACING shape: default the chooser to
// the controller, default the labels to Yes / No, and keep the two
// branches as ctx-taking effects so a card writes them the way it
// writes every other effect.
//
// # Addressed by seat
//
// Player is any seat. "You may" defaults it to the effect's
// controller; #568's opponent-facing cards set it to the opponent, and
// the prompt is otherwise identical — which is the point, and the
// reason #568 needed no second yes/no.
//
// # Undo safety
//
// The branches capture no *Game and no pointer into a zone. They
// receive the live game through a fresh Context bound to the same
// stack item, exactly as PayUnless.OnDecline does, so a branch run
// after an undo resolves against the restored game. The stack item
// itself is already detached from the stack by the time an effect
// resolves, and carries only Controller / SourceCardID / Targets.

// MayChoice asks Player "you may <Question>" during resolution and
// runs OnYes or OnNo when they answer.
//
// The calling effect has already done everything before the "you may";
// what remains is the answer, which arrives later via resolve_choice.
// Anything printed AFTER the decision — "If you do, ..." — goes in
// OnYes, not after this primitive returns, for the reason Scry.Then
// exists: Apply only queues the prompt.
type MayChoice struct {
	// Player is asked. Zero means the resolving effect's controller,
	// which is what "you may" means.
	Player uuid.UUID

	// Question is the prompt's header, written the way the card is —
	// "Eden, Seat of the Sanctum — sacrifice it?".
	Question string

	// YesLabel / NoLabel are the card's own words for the two
	// branches. Empty renders as Yes / No, which is what a prompt
	// that really is a yes/no wants.
	YesLabel, NoLabel string

	// LifeCost is the life the YES branch charges, declared for the
	// wire and for a bot's pricing (#547). The branch still performs
	// and re-checks the payment; this is a declaration, not a
	// deduction. Zero on most cards.
	LifeCost int

	// OnYes / OnNo are the two branches. Either may be nil — a "you
	// may" whose decline does nothing should not have to write an
	// empty closure to say so. Both run with g.mu held and may queue
	// further choices, which is how a chain continues.
	OnYes, OnNo func(ctx *Context) error
}

func (m MayChoice) Apply(ctx *Context) error {
	chooser := m.Player
	if chooser == uuid.Nil {
		chooser = ctx.Controller()
	}
	item := ctx.Item
	onYes, onNo := m.OnYes, m.OnNo
	ctx.Game.QueueConfirmForEffect(game.ConfirmPrompt{
		Chooser:      chooser,
		Source:       ctx.Source(),
		Question:     m.Question,
		AcceptLabel:  m.YesLabel,
		DeclineLabel: m.NoLabel,
		LifeCost:     m.LifeCost,
		OnAccept:     mayChoiceBranch(onYes, item),
		OnDecline:    mayChoiceBranch(onNo, item),
	})
	return nil
}

// mayChoiceBranch adapts a ctx-taking branch to the engine's
// `func(*game.Game) error` continuation contract, binding a FRESH
// Context to the game the resolver hands back rather than to the one
// that queued the prompt. That is the whole undo story: after an undo
// the restored game is a different object, and a branch closed over
// the old one would mutate a game nobody is looking at.
//
// A nil branch stays nil, so the engine's own "this branch does
// nothing" fast path is preserved rather than being buried under a
// closure that returns nil.
func mayChoiceBranch(branch func(ctx *Context) error, item *game.StackItem) func(g *game.Game) error {
	if branch == nil {
		return nil
	}
	return func(g *game.Game) error {
		return branch(NewContext(g, item))
	}
}
