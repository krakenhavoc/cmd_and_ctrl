package game

import (
	"fmt"
	"strconv"

	"github.com/google/uuid"
)

// option_pick.go — "choose one of the following", asked of a named
// seat while an effect is resolving (CR 608.2).
//
// # The gap this closes (#568)
//
// The queue could already ask a yes/no with two card-supplied branches
// (PendingChoiceConfirm, chained_choice.go) and a card-set pick
// (PendingChoiceChooseCards). What it could not ask is a choice among
// THREE OR MORE named consequences — Torment of Hailfire's "each
// opponent loses 3 life unless that player sacrifices a nonland
// permanent of their choice or discards a card" is the card that
// forced it, and a confirm cannot express it without nesting two
// yes/nos and asking the second question in an order the card does
// not print.
//
// A modal spell's "choose one —" is NOT this kind. That choice is made
// at announce (CR 601.2b), lives on Spec.Modes, and is over by the time
// the spell resolves. This one happens DURING resolution, is frequently
// addressed to somebody other than the controller, and each branch is a
// plain continuation supplied by the card.
//
// # Addressed by seat
//
// Chooser is any seat, not the resolving effect's controller. That is
// the whole of #568: Fact or Fiction's piles are separated by an
// opponent, Torment's choice is made by each opponent, and the engine
// had no prompt shape that could be pointed at another player without
// being a mana payment (pay_unless) or a trigger's own "you may".
//
// FromPlayer names whose material the options are about, and it is
// load-bearing rather than decorative: protocol's redaction pass reads
// it to decide what the CHOOSER is entitled to see when the pool is not
// theirs (see redactChoiceCards in protocol/view.go).
//
// # Two shapes, one kind
//
// An option may carry cards. That is what makes the second half of a
// pile split — "put one pile into your hand and the other into your
// graveyard" — the same prompt as Torment's three-way question rather
// than a kind of its own: two options, each holding a pile, addressed
// to the controller. The first half of a pile split is an ordinary
// PendingChoiceChooseCards addressed to the splitter, and the two are
// joined by #552's chaining, which is why "pile split" needs no kind
// at all.

// PendingChoiceOptionPick is "choose one of the following", addressed
// to any seat, answered with the INDEX of the chosen option.
//
// Answered with {option_index: N}, routed by the choice's KIND rather
// than by the field's presence — the meaningful value of the payload
// is zero (the first option), so there is nothing to route on. The
// same reason loop_shortcut is routed by kind.
//
// Every offered option is accepted: ResolveOptionPick validates the
// index and nothing else, exactly as ResolveConfirm validates nothing
// about the board — except an option that costs mana (#2854,
// ChoiceOption.ManaCost), which it pays and refuses when it no longer
// can be paid. The legality lives at QUEUE time — an effect builds
// the option list out of what the chooser can actually do, dropping
// "sacrifice a nonland permanent" for a player who controls none — and
// must offer at least one option that always works, which is the one
// the enumerator marks AlwaysLegal. A prompt whose every branch could
// fail is a prompt a seat can be stuck on (#544).
const PendingChoiceOptionPick PendingChoiceKind = "option_pick"

// NoChoiceIndex is the index a frame's continuation is run with when
// the question could not be put, or could not be kept up: "nobody
// chose". Not an answer a client may send — ResolveOptionPick refuses
// a negative index — and not an offset into anything.
//
// It is the #544 rule in one value. A continuation is the rest of the
// card, so an option pick that ends without an answer still has to run
// it; what it must not do is silently pick a branch on the chooser's
// behalf. Every continuation the engine queues already handles it,
// because the card-side primitive has always had to
// (effects.PickOption's Then is documented as "the index of the chosen
// option, or -1 when no question could be asked"): a pile pick takes
// the first pile, Torment of Hailfire moves on to the next victim, a
// player choice records the absence.
//
// Reached from two places, and #1006 is the second. The first is QUEUE
// time — an empty option list, or a chooser who has already left, so
// nothing is queued and the caller runs its own continuation. The
// second is DROP time: a prompt the engine withdraws unanswered runs
// the frame with this (dropDefault, leave_game.go).
//
// A SEAT continuation (optionPickFrame.thenSeat, #994) spells the same
// absence as uuid.Nil rather than as an index, because that is the
// currency it speaks; runWithNoChoice is the one place both are
// written.
const NoChoiceIndex = -1

// ChoiceOption is one branch of a PendingChoiceOptionPick: the card's
// own words for it, the cards it is about (a pile, or nothing), and
// what it costs the chooser in life.
//
// It is plain data and goes on the wire. What the option MEANS is the
// frame's business, not this struct's — the index comes back and the
// queuing effect decides, which is what keeps the kind reusable
// instead of being a second modal-spell system.
type ChoiceOption struct {
	// Label is the option's printed sentence — "Lose 3 life",
	// "Sacrifice a nonland permanent", "Put this pile into your
	// hand". Required: it is the whole of what the chooser reads.
	Label string

	// Cards are the cards this option is about, in render order.
	// Empty for an option that names no cards, which is most of them.
	// Projected per viewer through the same redaction pass as every
	// other card list on a prompt, so an option over a pool the
	// chooser does not own shows only what that seat may legally see.
	Cards []uuid.UUID

	// LifeCost is the life this branch charges, for the wire and for
	// a bot's pricing — the same declaration ConfirmPrompt.LifeCost
	// makes, and for the same #547 reason: a policy holding only the
	// wire payload otherwise prices "lose 3 life" exactly like
	// "discard a card", and a bot at 3 life picks the life and dies.
	//
	// The engine does NOT deduct it. The branch does.
	LifeCost int

	// Player is the SEAT this option is about, for an option list
	// whose branches are players rather than consequences — every
	// prompt built by choose_player.go. uuid.Nil on every other
	// option, which is all of them (a pile, a Torment branch): the
	// zero value means "this option is not about a seat" and nothing
	// reads it.
	//
	// It exists because an option has to be able to name its SUBJECT
	// (#994). The engine re-checks an open prompt whenever the board
	// moves under it — reassignChoiceLocked prunes the cards a
	// departed player took with them — and it could do that only for
	// options built out of cards. A seat option carried its seat as
	// rendered TEXT in Label, so a departed seat was identifiable
	// only by string-matching a player's name, and a prompt went on
	// offering a player who was no longer a player (CR 800.4a).
	// With the id here, seats prune exactly the way cards do:
	// pruneDepartedSeatOptionsLocked and reassignChoiceLocked are
	// the two paths, and both key on this field.
	//
	// Label is still what the chooser READS; this is what the engine
	// checks. The two are set together by one function
	// (seatChoiceOptionsLocked) so they cannot disagree.
	Player uuid.UUID

	// ManaCost is the mana this branch costs the chooser, as a cost
	// string ("{1}", "{B}", "{3}"), or "" for a branch that costs no
	// mana (#2854). Winter's Chill: "its controller may pay {1} or
	// {2}" is three branches — pay {2}, pay {1}, pay nothing — and the
	// first two carry their price here.
	//
	// Unlike LifeCost, the ENGINE pays it: the option is a payment
	// (CR 118.12, a cost paid as the effect resolves), so it is paid
	// the way every other resolution-time payment is, through
	// payCostLocked (the pay_unless payer: the floating pool first,
	// then the auto-tapper, no spend-restricted mana, nothing tapped
	// when it cannot be paid in full). An option the chooser cannot
	// afford is not offered (CR 118.3, 608.2d): QueueOptionPickForEffect
	// drops it, and ResolveOptionPick refuses it if the board moved
	// between the question and the answer. No {X}, and never on the
	// first option, which is the always-legal way out.
	ManaCost string `json:",omitempty"`

	// Harm is what this branch costs the chooser besides life and
	// mana, as the card's own estimate in cards (1 is about one card
	// from hand, 2 about a creature), or 0 for a branch that costs
	// nothing a policy could price (#3140). It exists for the prompts
	// whose branches are both bad — a villainous choice (CR 701.55) —
	// so a bot can take the lesser. It is a hint on the wire and
	// nothing more: the engine never reads it, the branch does the
	// harm.
	Harm int `json:",omitempty"`

	// Dungeon and Room name the dungeon card and the room this branch
	// enters, by key, on a venture prompt (CR 701.49, ADR 0203): Room
	// on "choose the next room", Dungeon alone on "choose a dungeon".
	// The venture's keyed continuation reads them, which is what makes
	// the answer independent of the option order, and a bot prices the
	// room by its key. Empty on every other option.
	Dungeon string `json:",omitempty"`
	Room    string `json:",omitempty"`
}

// cloneChoiceOptions deep-copies an option list. Each option owns a
// card slice, so a shallow copy would let an undo snapshot and the
// live game share a backing array — the same rule every other slice
// on a PendingChoice follows (clone.go).
func cloneChoiceOptions(in []ChoiceOption) []ChoiceOption {
	if len(in) == 0 {
		return nil
	}
	out := make([]ChoiceOption, len(in))
	for i, opt := range in {
		out[i] = opt
		out[i].Cards = copyUUIDs(opt.Cards)
	}
	return out
}

// optionPickFrame is the continuation behind a
// PendingChoiceOptionPick that is not keyed: ONE closure, given the
// seat the chooser picked.
//
// #4361 slice 6 retired the index closure (`then`), the subject closure
// (`thenSubject`) and the Aura-attach marker: every option pick but one
// is keyed (OptionPickPrompt.ThenKey, option_pick_keyed.go), and an
// Aura's "attach it to a player" question is recognised by its key
// (attachPromptOpenForLocked). What is left is the seat closure of a
// trigger's "of an opponent's choice" walk (pending_choice.go), which
// waits on the pick-target frame becoming data.
//
// It receives the live *Game (not a captured one) on the same
// undo-safety contract confirmFrame and StackItem.Effect follow, and
// runs with g.mu held — so it may queue the next link of a chain.
type optionPickFrame struct {
	// thenSeat is the continuation for an option list whose branches
	// are SEATS. It receives the chosen option's ChoiceOption.Player
	// instead of its index.
	//
	// It reads the seat rather than an index into a captured slice
	// because an INDEX is not a stable answer (#994). The option list of
	// a player prompt is pruned while it is open — a seat that leaves
	// the game comes off it (CR 800.4a) — so index 2 before the prune and
	// index 2 after it are different players. Reading the seat off the
	// option the chooser actually picked cannot drift from what they were
	// shown, because it IS what they were shown.
	thenSeat func(g *Game, seat uuid.UUID) error
}

// subject is what an option is ABOUT: its seat, or its first card, or
// nothing.
func (o ChoiceOption) subject() uuid.UUID {
	if o.Player != uuid.Nil {
		return o.Player
	}
	if len(o.Cards) > 0 {
		return o.Cards[0]
	}
	return uuid.Nil
}

// runWithNoChoice runs the frame as though nobody chose. Reports the
// error the continuation returned, for the caller to report the way it
// reports every other continuation failure.
//
// The one place "nobody chose" is handed to a frame: a SEAT frame is
// run with uuid.Nil (#994's thenSeat), the absence spelled in the
// currency it speaks, as a keyed pick is run with NoChoiceIndex. It is
// a value the continuation already has to handle, and not an answer a
// client can send.
func (f *optionPickFrame) runWithNoChoice(g *Game) error {
	if f == nil || f.thenSeat == nil {
		return nil
	}
	return f.thenSeat(g, uuid.Nil)
}

// defaultDroppedChoiceLocked runs an option pick's continuation with
// the no-choice outcome because the prompt has been dropped rather
// than answered. It is the dropDefault action of the departure table
// (choiceDepartureDecisions, leave_game.go) and the CR 800.4f/#544
// pair applied to this kind: the question ends, and the rest of the
// card does not.
//
// Before #1006 the drop took the frame with it, and the effect that
// was paused mid-resolution never finished — Fact or Fiction put
// neither pile anywhere, a Torment of Hailfire stopped at the victim
// who left, a "choose a player" never ran the sentence printed after
// it. QueuePileSplitForEffect and QueueChoosePlayerForEffect were
// already careful about exactly this at QUEUE time; nothing was
// careful about it at drop time.
//
// This is the SAME continuation ResolveOptionPick runs for an answer,
// reached from a withdrawal rather than from one — the #808 shape, and
// the same one declineDepartedChoiceLocked is. It cannot re-queue a
// prompt to a departed seat (QueueChoiceForEffect refuses an
// eliminated chooser, #864), and it takes no branch on the chooser's
// behalf, which is why it needs no claim about whose material the
// options were.
//
// Caller must hold g.mu, and must already have taken the prompt out of
// the queue: the continuation may queue the next link of the chain and
// must not land behind the question it is replacing.
// #1019 added the SECOND kind to declare dropDefault and #1027 the
// third, and both are the same idea in a different currency: a
// withdrawn sacrifice or discard prompt is a seat that moved nothing,
// and the RUN waiting on it has to be told so — a run whose last leg
// was dropped silently is a card that stops halfway exactly as a
// dropped option pick was. They are one function rather than three
// drop paths because the departure table has one second column, and a
// kind settled in two places is a kind settled two ways.
//
// The branch is on the PROMPT rather than on the kind (#1027). Two
// kinds carried a run link when that rule was written and the
// discard's is a PendingChoiceChooseCards — the same kind
// Thoughtseize's revealed-hand pick uses, which is no run's leg — so
// "what is waiting on this prompt" is a question about the prompt, not
// about its kind. A fourth verb then needs no fourth case here.
//
// THREE SHAPES A DROPPED PROMPT CAN HAVE, in the order they are asked
// (#1225):
//
//  1. It is one LEG OF A RUN. The leg settles with nothing moved and
//     the run's continuation runs after the last of them. Asked FIRST
//     and exclusively: a run leg's own frame is the leg's sentence and
//     settles the run from inside itself (the discard's
//     discardCardsLocked into opts.then), so running that frame here
//     as well would settle the leg twice and pay the run out early.
//  2. It carries a CARD-SET FRAME whose continuation is a plain
//     closure — a chained pick asked from inside a resolution that is
//     paused waiting for it, which is every mid-card choose_cards:
//     effects.SacrificeChoice's "Torment of Hailfire repeats X times",
//     Gluntch's counters before the second player draws, Ward's
//     sacrifice before the spell is or is not countered. The frame
//     runs with NOTHING PICKED, which is the outcome this payload
//     reserves for "nobody chose" and the one every such continuation
//     already has to handle: with no short-circuit in
//     QueueChooseCardsForEffect, each caller wrote the empty-candidate
//     path itself and it runs the same closure with the same empty
//     answer. Validate is not consulted — it is documented as never
//     being called for an empty pick, because a floor of zero has to
//     keep "choose nothing" as an answer the engine cannot refuse.
//  3. It carries an OPTION-PICK FRAME. NoChoiceIndex, as since #1006.
//
// A prompt with none of the three — a Thoughtseize-shaped pick whose
// caller wanted nothing after it — still runs nothing, which is what
// the nil receivers say.
//
// Caller must hold g.mu, and must already have taken the prompt out of
// the queue: the continuation may queue the next link of the chain and
// must not land behind the question it is replacing.
func (g *Game) defaultDroppedChoiceLocked(c *PendingChoice) {
	if c == nil {
		return
	}
	var err error
	switch {
	case c.promptRun != uuid.Nil:
		err = g.settleRunLegLocked(c.promptRun, c.Chooser, nil)
	case c.chooseCardsResume != nil:
		err = c.chooseCardsResume.runWithNoChoice(g)
	case isKeyedChooseCards(c):
		// #2040: a keyed choose_cards pick, run with nobody's choice.
		err = g.runKeyedChooseCardsLocked(c, nil, false)
	case chooseValueFrameOf(c) != nil:
		// A resolution-time "choose a creature type" whose chooser left
		// (Patriarch's Bidding asks every seat in turn). "" is the
		// absence of a choice, spelled in the currency the frame
		// speaks; an ANSWER can never be empty (CanonicalCreatureType
		// refuses it), so it cannot be mistaken for one.
		err = chooseValueFrameOf(c).runWithNoChoice(g)
	case c.Kind == PendingChoiceOptionPick && c.CouncilVote != nil:
		// ADR 0146: a ballot whose voter left casts nothing, and the
		// vote goes on to the next player (CR 800.4a). When someone was
		// choosing FOR the voter (Illusion of Choice) and it is that
		// chooser who left, the voter has not: their ballot goes back to
		// them, and the next one is asked as an ordinary vote.
		if v := c.CouncilVote; c.Chooser != v.Voter && g.playerInGameLocked(v.Voter) {
			err = g.askNextBallotLocked(v)
		} else {
			err = g.castBallotLocked(v, v.Voter, nil)
		}
	case c.Kind == PendingChoiceOptionPick && c.PickThen != "":
		// #2854: a keyed option pick, run with nobody's choice.
		err = g.withCarriedCauseLocked(carriedCauseOfChoice(c), func() error {
			return runOptionPickThen(g, c.PickThen, OptionPicked{
				Chooser: c.Chooser, Source: c.Source, Index: NoChoiceIndex, Carry: copyUUIDs(c.OptionCarry), dice: c.dice,
				parked: c.opponentTargetResume,
				Asked:  c.Asked.StackItem(),
				Words:  copyStrings(c.PickWords), Nums: copyInts(c.PickNums), Tail: copyKeyedTail(c.pickTail),
			})
		})
	default:
		err = c.optionPickResume.runWithNoChoice(g)
	}
	if err != nil {
		g.emitChoiceEffectErrorLocked(c.Chooser, c.Source, err)
	}
}

// OptionPickPrompt is the queue-side description of a
// PendingChoiceOptionPick.
type OptionPickPrompt struct {
	// Chooser answers the prompt. Required, and frequently NOT the
	// resolving effect's controller.
	Chooser uuid.UUID

	// FromPlayer owns the material the options are about. Defaults to
	// Chooser when empty. Set it when the options name cards the
	// chooser does not own — the redaction pass reads it.
	FromPlayer uuid.UUID

	// Source is the card asking.
	Source uuid.UUID

	// Question is the prompt's header — the card's own sentence.
	Question string

	// Options are the branches, in the order the card prints them.
	// The FIRST is the one the enumerator marks always-legal, so it
	// must be a branch that can always be taken.
	Options []ChoiceOption

	// ThenSeat is the closure continuation for an option list whose
	// branches are SEATS: it receives the chosen option's Player rather
	// than its index, or uuid.Nil when the prompt ended without an
	// answer. Its one engine caller is a trigger's "of an opponent's
	// choice" walk (pending_choice.go), which waits on the pick-target
	// frame becoming data (#4361); every other option pick is keyed.
	//
	// An index is not a stable answer for a seat list: a seat that
	// leaves the game is pruned off an OPEN prompt (#994, CR 800.4a),
	// which renumbers everything after it. A keyed continuation reads
	// the seat off OptionPicked.Option for the same reason.
	ThenSeat func(g *Game, seat uuid.UUID) error

	// walk and parked are #2055's markers, set only by the "which
	// opponent chooses" question of an "of an opponent's choice" clause
	// (opponent_target.go): walk is the trigger's CR 603.3d target walk
	// the question belongs to, carried as the prompt's pickTargetResume
	// so the drain holds the batch behind it and the prompt belongs to
	// no resolution; parked is the parked cast or activation, carried as
	// its opponentTargetResume so the same announcement is not parked
	// twice. Both are set on the choice BEFORE it is queued, where the
	// queue reads them. Neither is the prompt's continuation.
	walk   *pickTargetFrame
	parked *opponentTargetFrame

	// ThenKey is the REGISTERED continuation (RegisterOptionPickThen,
	// option_pick_keyed.go), and Carry is the plain data it is handed
	// with the answer (#2854). A prompt queued with a key holds no
	// closure, so a table waiting on it is still a restore point: the
	// key and the carried IDs are written with the prompt, and the
	// running binary looks the continuation up when the answer arrives.
	// #4361 slice 6 retired the index closure Then: every engine and
	// card prompt but one sets this. With a key, ThenSeat is ignored.
	ThenKey OptionPickThen
	Carry   []uuid.UUID

	// dice is a paused dice roll (#4307, dice_roll.go), written with the
	// prompt; set only by the dice flow, with a ThenKey.
	dice *diceRollPause

	// Asked is the asking stack item as data (NewAskedItem), handed to a
	// keyed continuation as OptionPicked.Asked (#4253 slice 8): the
	// attack-target pick of a card whose entry rides a tail that reads the
	// item (Ultra Magnus, Tactician). Ignored without ThenKey.
	Asked *AskedItem
	// Words and Nums are a keyed continuation's other plain data (#4253
	// slice 10c, 10b): handed back as OptionPicked.Words and Nums, written
	// with the prompt as PickWords and PickNums. Tail is a registered tail
	// it hands on (OptionPicked.Tail; #4253 slice 10b: a pile split's rest
	// of the card). Ignored without ThenKey.
	Words []string
	Nums  []int
	Tail  KeyedTail
}

// QueueOptionPickForEffect queues a "choose one of the following" and
// returns its ID, or uuid.Nil when it queued nothing — an empty option
// list, or a chooser who has left the game (QueueChoiceForEffect's own
// CR 800.4a guard).
//
// A caller with more of the effect to run after the answer must check
// the return: nothing was queued means nothing will run ThenKey, and a
// chain that assumed otherwise stops halfway. That is the same
// contract QueueDiscardChoiceForEffect documents, pointed the other
// way — a discard with nothing to pitch still runs Then, because CR
// 701.9a says "as many as you can"; a question with no answers was
// never asked at all.
//
// Options that carry a ManaCost (#2854) are offered only when the
// chooser can pay them now (CR 118.3): an unaffordable one is dropped
// here, so the index a keyed continuation is handed is an offset into
// the list the chooser was SHOWN: it reads the chosen option itself
// (OptionPicked.Option).
// A costed option with an unparseable or {X} cost, or a cost on the
// FIRST option, is a bug in the card: it is reported as an effect
// error and nothing is queued.
//
// Caller must hold g.mu.
func (g *Game) QueueOptionPickForEffect(p OptionPickPrompt) uuid.UUID {
	if len(p.Options) == 0 {
		return uuid.Nil
	}
	from := p.FromPlayer
	if from == uuid.Nil {
		from = p.Chooser
	}
	options, err := g.affordableOptionsLocked(p.Chooser, p.Source, p.Options)
	if err != nil {
		g.emitChoiceEffectErrorLocked(p.Chooser, p.Source, err)
		return uuid.Nil
	}
	c := PendingChoice{
		Kind:        PendingChoiceOptionPick,
		Chooser:     p.Chooser,
		FromPlayer:  from,
		Count:       1,
		Source:      p.Source,
		Reason:      p.Question,
		PickOptions: options,
	}
	// #2055's markers ride either form: an opponent-target chooser is
	// keyed (#4361 slice 6), a trigger walk's is still a closure.
	c.pickTargetResume = p.walk
	c.opponentTargetResume = p.parked
	if key := p.ThenKey.key; key != "" {
		c.PickThen = key
		c.OptionCarry = copyUUIDs(p.Carry)
		c.dice = p.dice
		c.Asked = copyAskedItem(p.Asked)
		c.PickWords, c.PickNums = copyStrings(p.Words), copyInts(p.Nums)
		c.pickTail = copyKeyedTail(p.Tail)
		return g.QueueChoiceForEffect(c)
	}
	if p.ThenSeat != nil {
		c.optionPickResume = &optionPickFrame{thenSeat: p.ThenSeat}
	}
	return g.QueueChoiceForEffect(c)
}

// affordableOptionsLocked is the option list as offered: every option
// that costs no mana, and every costed one `chooser` can pay right now
// through the payer ResolveOptionPick will use (canPayCostLocked asks
// payCostLocked on a throwaway clone). Caller must hold g.mu.
func (g *Game) affordableOptionsLocked(chooser, source uuid.UUID, opts []ChoiceOption) ([]ChoiceOption, error) {
	costed := false
	for i, o := range opts {
		if o.ManaCost == "" {
			continue
		}
		if i == 0 {
			return nil, fmt.Errorf("option pick: the first option (%q) is the always-legal one and cannot cost mana", o.Label)
		}
		cost, err := ParseCost(o.ManaCost)
		if err != nil {
			return nil, fmt.Errorf("option pick: unparseable cost %s: %w", o.ManaCost, err)
		}
		if cost.XSlots > 0 {
			return nil, fmt.Errorf("option pick: cost %s has an X", o.ManaCost)
		}
		costed = true
	}
	if !costed {
		return cloneChoiceOptions(opts), nil
	}
	p := g.playerByIDLocked(chooser)
	var out []ChoiceOption
	for _, o := range opts {
		if o.ManaCost != "" {
			cost, _ := ParseCost(o.ManaCost)
			if p == nil || !g.canPayCostLocked(p, cost, source) {
				continue
			}
		}
		out = append(out, o)
	}
	return cloneChoiceOptions(out), nil
}

// payChosenOptionLocked pays the chosen option's ManaCost, if it has
// one. False, with nothing paid, when the chooser can no longer pay it
// (the board moved since the question went up). Caller must hold g.mu.
func (g *Game) payChosenOptionLocked(chooser, source uuid.UUID, o ChoiceOption) bool {
	if o.ManaCost == "" {
		return true
	}
	cost, err := ParseCost(o.ManaCost)
	if err != nil {
		return false
	}
	p := g.playerByIDLocked(chooser)
	if p == nil {
		return false
	}
	return g.payCostLocked(p, cost, source, nil)
}

// ResolveOptionPick answers a PendingChoiceOptionPick: `index` is the
// offset into the prompt's option list.
//
// The choice is dequeued BEFORE the branch runs, for chaining's sake —
// the branch queues the next prompt, and a queue that still held this
// one would hand the client two open modals and the enumerator two
// owed choices for the same seat (ResolveConfirm's reason, verbatim).
//
// An out-of-range index is refused with the prompt still open, so a
// client that sends a stale index can try again. A branch that ERRORS
// emits EventEffectError and the prompt is still gone: a failed
// continuation must not leave an unanswerable prompt in the queue,
// which is the #544 shape.
//
// Caller must NOT hold g.mu.
func (g *Game) ResolveOptionPick(choiceID, chooserID uuid.UUID, index int) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	idx, choice := g.findChoiceLocked(choiceID)
	if idx < 0 {
		return ErrPendingChoiceNotFound
	}
	if choice.Kind != PendingChoiceOptionPick {
		return ErrInvalidParam
	}
	if choice.Chooser != chooserID {
		return ErrNotTheChooser
	}
	if index < 0 || index >= len(choice.PickOptions) {
		return ErrInvalidParam
	}
	frame := choice.optionPickResume
	source := choice.Source
	// Read off the option the chooser actually picked, BEFORE the
	// dequeue takes the list away. A seat continuation is answered with
	// this rather than with the index (#994): the list can have been
	// pruned since it was built, so the index is only meaningful
	// against the list the chooser was shown, which is this one.
	seat := choice.PickOptions[index].Player
	if frame != nil && frame.thenSeat != nil && seat == uuid.Nil {
		// A seat continuation answered with an option that names no
		// seat is a bug in whoever built the list, and it must not
		// reach the frame: uuid.Nil is how a DROP says nobody was
		// chosen (#1006, runWithNoChoice), so passing it here would
		// report an answer as an absence. Refused before the dequeue,
		// like an out-of-range index, so the prompt stays open.
		return ErrInvalidParam
	}
	// #2854: an option that costs mana is paid now, before the prompt
	// goes, through the pay_unless payer. The option was affordable
	// when it was offered; if the board has moved since (a land left,
	// the pool emptied), it is REFUSED with the prompt still open and
	// nothing paid — CR 118.3: a cost that cannot be paid cannot be
	// chosen — and the chooser answers again. The free first option is
	// always there to answer with.
	chosen := choice.PickOptions[index]
	if !g.payChosenOptionLocked(chooserID, source, chosen) {
		return ErrInsufficientMana
	}
	if v := choice.CouncilVote; v != nil {
		// ADR 0146: a ballot. Cast it, then ask the next one.
		g.dequeueChoiceLocked(idx)
		g.emitChoiceEffectErrorLocked(chooserID, source, g.castBallotLocked(v, v.Voter, &chosen))
		g.runStateChecksLocked()
		return nil
	}
	if key := choice.PickThen; key != "" {
		carry := copyUUIDs(choice.OptionCarry)
		dice := choice.dice
		parked := choice.opponentTargetResume
		asked := choice.Asked.StackItem()
		words, nums, tail := copyStrings(choice.PickWords), copyInts(choice.PickNums), copyKeyedTail(choice.pickTail)
		cause := carriedCauseOfChoice(choice) // #4281
		g.dequeueChoiceLocked(idx)
		g.emitChoiceEffectErrorLocked(chooserID, source, g.withCarriedCauseLocked(cause, func() error {
			return runOptionPickThen(g, key, OptionPicked{
				Chooser: chooserID, Source: source, Index: index, Option: &chosen, Carry: carry, dice: dice,
				parked: parked,
				Asked:  asked,
				Words:  words, Nums: nums, Tail: tail,
			})
		}))
		g.runStateChecksLocked()
		return nil
	}
	g.dequeueChoiceLocked(idx)
	if frame == nil || frame.thenSeat == nil {
		// Nothing to run. The prompt is gone either way rather than
		// stuck: an option pick with no frame is a bug in whoever
		// queued it, and refusing the answer would wedge the seat.
		return nil
	}
	if err := frame.thenSeat(g, seat); err != nil {
		g.EmitEvent(Event{
			Kind:     EventEffectError,
			Actor:    chooserID,
			Source:   source,
			ErrorMsg: err.Error(),
		})
	}
	g.runStateChecksLocked()
	return nil
}

// QueuePileSplitForEffect is the two-prompt "separate these cards into
// two piles; somebody else takes one of them" shape (CR 608.2), built
// out of the two kinds above rather than a kind of its own.
//
// Fact or Fiction is the card: "Reveal the top five cards of your
// library. An opponent separates those cards into two piles. Put one
// pile into your hand and the other into your graveyard."
//
// Two chained prompts to two different seats (#552):
//
//  1. PendingChoiceRevealPick to the SPLITTER — "choose the cards for
//     the first pile", floor zero, ceiling all of them. An empty pile
//     is legal and is frequently the right split — "two piles" does
//     not require both to hold a card, and splitting 5-0 is a real
//     Fact or Fiction play.
//  2. PendingChoiceOptionPick to the CHOOSER — two options, each
//     carrying its pile's cards, so the client renders the two piles
//     rather than quoting names into a sentence.
//
// The first leg was a plain PendingChoiceChooseCards until #1214 and
// is now a reveal_pick, which is the same question stated once: "an
// opponent picks from a set you revealed". Two things follow, and the
// second is a bug this move fixes. The prompt no longer hides its
// bounds from the rest of the table — the five cards are PUBLIC, so
// "separate these five" is public too, where a choose_cards over a
// hand rightly is not. And the split leg is now one leg of a RUN, so a
// splitter who leaves AFTER the question goes up settles it with
// nothing separated and the controller still gets to take a pile;
// before, the departure table's dropDefault found no run and no option
// frame on the prompt, ran nothing, and Fact or Fiction put neither
// pile anywhere — which is precisely the failure #1006 named and fixed
// one leg further along.
//
// ThenTail is told the pile the chooser TOOK first and the other one
// second, both in the order they were split. A splitter who has left
// the game, or a card list with nothing in it, runs it with everything
// in the first pile and nothing in the second: the reveal still
// happened and the rest of the card still has to resolve.
//
// Caller must hold g.mu.
func (g *Game) QueuePileSplitForEffect(p PileSplitPrompt) {
	if p.ThenTail.IsZero() {
		// Nothing hangs off the split: nothing to ask (#4253 retired the
		// closure Then; a split with neither asked nothing either).
		return
	}
	// #4253 slice 10b: both links and the rest are data
	// (pile_split_keyed.go).
	g.queuePileSplitKeyedLocked(p)
}

// PileSplitPrompt is the queue-side description of a pile split.
type PileSplitPrompt struct {
	// Splitter separates the cards — the opponent, on every printed
	// card of this family. Required.
	Splitter uuid.UUID
	// Chooser takes one of the two piles. Required; the resolving
	// effect's controller on every printed card of this family.
	Chooser uuid.UUID
	// Owner owns the cards. Defaults to Chooser. Read by the
	// redaction pass to decide what the splitter may see.
	Owner uuid.UUID
	// Source is the card asking.
	Source uuid.UUID
	// SplitQuestion / PickQuestion are the two prompts' headers.
	SplitQuestion, PickQuestion string
	// Cards are the cards being separated, in the order the table saw
	// them revealed.
	Cards []uuid.UUID
	// ThenTail is the rest of the card (#4253 slice 10b): a registered
	// tail told the pile the chooser took and then the pile they left as
	// its IDs, with how many were taken as its Amount. It runs with g.mu
	// held. Both prompts are keyed, so a table waiting on either is a
	// restore point. Required: a zero tail queues nothing.
	ThenTail KeyedTail
	// FaceDown makes this the "face-down pile and a face-up pile" split
	// (Sauron's Ransom, Fortune's Favor, Atris, Riddles in the Dark,
	// Curator of Destinies; #2147). The splitter's cards are picked as the
	// FACE-DOWN pile, the rest are the face-up pile and are revealed to the
	// whole table the moment the split is answered, and the pick is offered
	// face-up pile first. The caller must already have made the splitter a
	// knower of every card (a "look at", CR 701.20a): the face-down pile
	// is never revealed, so only the knower set keeps it off the chooser's
	// wire and the table's.
	FaceDown bool
}

// facePileLabel names a face-up or face-down pile for its option
// button: "Take the face-down pile (3 cards)".
func facePileLabel(kind string, cards []uuid.UUID) string {
	switch len(cards) {
	case 0:
		return "Take the " + kind + " pile (empty)"
	case 1:
		return "Take the " + kind + " pile (1 card)"
	default:
		return "Take the " + kind + " pile (" + strconv.Itoa(len(cards)) + " cards)"
	}
}

// partitionPiles splits `all` into the cards named by `first` and the
// rest, each in `all`'s order. A card named in `first` that is not in
// `all` is ignored — the answer was validated against the candidate
// list before it got here, and this function is the one place the two
// piles are derived, so it does not get to disagree with that.
func partitionPiles(all, first []uuid.UUID) (one, two []uuid.UUID) {
	inFirst := make(map[uuid.UUID]bool, len(first))
	for _, id := range first {
		inFirst[id] = true
	}
	for _, id := range all {
		if inFirst[id] {
			one = append(one, id)
			continue
		}
		two = append(two, id)
	}
	return one, two
}

// pileLabel names a pile for the option button: "Pile 1 (3 cards)".
// The COUNT is on the label because the client renders the pile's
// cards beside it and a zero-card pile would otherwise be a button
// with nothing under it — an empty pile is a legal split and has to
// look deliberate rather than broken.
func pileLabel(n int, cards []uuid.UUID) string {
	switch len(cards) {
	case 0:
		return "Take pile " + strconv.Itoa(n) + " (empty)"
	case 1:
		return "Take pile " + strconv.Itoa(n) + " (1 card)"
	default:
		return "Take pile " + strconv.Itoa(n) + " (" + strconv.Itoa(len(cards)) + " cards)"
	}
}

// emitChoiceEffectErrorLocked reports a continuation that failed while
// a prompt was being answered or skipped. Every resolver in the queue
// does this by hand; the pile split does it from three places, so it
// is one function here rather than three copies.
//
// Caller must hold g.mu.
func (g *Game) emitChoiceEffectErrorLocked(actor, source uuid.UUID, err error) {
	if err == nil {
		return
	}
	g.EmitEvent(Event{
		Kind:     EventEffectError,
		Actor:    actor,
		Source:   source,
		ErrorMsg: err.Error(),
	})
}
