package effects

import (
	"fmt"
	"strconv"

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
// The answer is a registered body (MayAnswer) and captures nothing. It
// receives the live game through a fresh Context bound to a copy of the
// asking item rebuilt from data the prompt carries (game.AskedItem), so
// an answer given after an undo or a restore resolves against the
// restored game.

// MayChoice asks Player "you may <Question>" during resolution and
// runs the registered answer Then when they answer.
//
// The calling effect has already done everything before the "you may";
// what remains is the answer, which arrives later via resolve_choice.
// Anything printed AFTER the decision — "If you do, ..." — goes in
// Then, not after this primitive returns, for the reason Scry.ThenTail
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

	// DeclineEndsLoop is PendingChoice.DeclineEndsLoop (#2040, #4298): saying
	// yes would continue a loop the table is already in. A hint for a bot.
	DeclineEndsLoop bool

	// Draws, when above zero, says the YES branch is a choice to draw
	// that many cards, for Drawer. A player a draw limit stops drawing
	// that many can't choose to (CR 121.2b, CR 121.3), and neither can
	// anyone choosing for them (CR 121.3a, Combustible Gearhulk's
	// opponent), so nothing is asked and Then runs at once as a "no"
	// (#4016).
	Draws int

	// Drawer is who draws on a "yes". Zero means the player asked,
	// which is what "you may draw" and "each player may draw" mean.
	Drawer uuid.UUID

	// Then is the answer, a REGISTERED one (MayAnswer, MayYes, MayYesNo,
	// #4264), and Carry the plain data it is handed: the prompt holds no
	// closure, so a table waiting on it is a restore point. It runs with
	// g.mu held and may queue further choices, which is how a chain
	// continues. When a draw limit means nothing is asked (Draws), Then
	// runs at once as a "no". Required (#4253 retired the closure
	// branches).
	Then  game.ConfirmThen
	Carry []uuid.UUID

	// Word is the one word of the card's text a keyed answer is
	// parameterised by (#4291): an amount ("2"), a token's name. A shared
	// answer ("draw N cards", "gain N life") reads it as r.Word, so many
	// cards use one registered key. Set with Then.
	Word string

	// Objects are the objects, with the epoch each had, a keyed answer is
	// handed beside Carry's IDs (r.Objects): a permanent the card named as it
	// asked, so the answer can tell it from a new object the same card has
	// become (CR 400.7).
	Objects []game.ObjectRef
}

// MayAnswer registers a keyed "you may" answer under `key`, an on-disk
// identity (never renamed, never reused, ledgered "confirm <key>"). Call
// it once, from a package-level var. The body runs against a Context
// bound to a copy of the asking item rebuilt from data the prompt carries
// (game.AskedItem: its controller, source card and source object with its
// epoch, targets, triggering event, X and modes, #4291), because the item
// itself has long resolved by the time the player answers (and a restore
// does not carry it). r.Accepted is the answer, r.Carry the card's own
// Carry and r.Word its Word. A prompt queued without the item gets a
// stand-in with the asking controller and the source card.
func MayAnswer(key string, body func(ctx *Context, r game.ConfirmAnswer) error) game.ConfirmThen {
	return game.RegisterConfirmThen(key, func(g *game.Game, r game.ConfirmAnswer) error {
		// MayChoice.Apply puts the asking effect's controller first; the
		// card's own Carry follows it.
		var ctx *Context
		ctx, r.Carry = askedContext(g, r.Item, r.Chooser, r.Source, r.Carry)
		return body(ctx, r)
	})
}

// MayYes registers a keyed "you may" answer that does something only on a
// yes (the common shape: "you may draw a card"). The body runs as
// MayAnswer's does, and a decline does nothing.
func MayYes(key string, body func(ctx *Context, r game.ConfirmAnswer) error) game.ConfirmThen {
	return MayAnswer(key, func(ctx *Context, r game.ConfirmAnswer) error {
		if !r.Accepted {
			return nil
		}
		return body(ctx, r)
	})
}

// MayYesNo registers a keyed answer with a branch for each reply, for
// "you may X. If you don't, Y". Either may be nil.
func MayYesNo(key string, onYes, onNo func(ctx *Context, r game.ConfirmAnswer) error) game.ConfirmThen {
	return MayAnswer(key, func(ctx *Context, r game.ConfirmAnswer) error {
		branch := onNo
		if r.Accepted {
			branch = onYes
		}
		if branch == nil {
			return nil
		}
		return branch(ctx, r)
	})
}

// wordInt reads the amount a keyed answer was parameterised with
// (MayChoice.Word), 0 when it holds none.
func wordInt(r game.ConfirmAnswer) int {
	n, _ := strconv.Atoi(r.Word)
	return n
}

func (m MayChoice) Apply(ctx *Context) error {
	chooser := m.Player
	if chooser == uuid.Nil {
		chooser = ctx.Controller()
	}
	if m.Then.Key() == "" {
		return fmt.Errorf("%w: MayChoice with no registered answer (MayAnswer)", game.ErrInvalidParam)
	}
	carry := append([]uuid.UUID{ctx.Controller()}, m.Carry...)
	asked := game.NewAskedItem(ctx.Item)
	if m.Draws > 0 {
		drawer := m.Drawer
		if drawer == uuid.Nil {
			drawer = chooser
		}
		if !ctx.Game.CanChooseToDrawForEffect(drawer, m.Draws) {
			return ctx.Game.RunConfirmThenForEffect(m.Then, game.ConfirmAnswer{
				Chooser: chooser, Source: ctx.Source(), Carry: carry,
				LifeCost: m.LifeCost, Word: m.Word, Objects: m.Objects, Item: asked.StackItem(),
			})
		}
	}
	_, err := ctx.Game.QueueConfirmKeyedForEffect(game.ConfirmPrompt{
		Chooser:         chooser,
		Source:          ctx.Source(),
		Question:        m.Question,
		AcceptLabel:     m.YesLabel,
		DeclineLabel:    m.NoLabel,
		LifeCost:        m.LifeCost,
		DeclineEndsLoop: m.DeclineEndsLoop,
		Asked:           asked,
		Word:            m.Word,
		Objects:         m.Objects,
	}, m.Then, carry)
	return err
}
