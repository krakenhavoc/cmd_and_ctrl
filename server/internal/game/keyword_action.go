package game

import (
	"errors"
	"strconv"

	"github.com/google/uuid"
)

// keyword_action.go is the CR 614 window on a KEYWORD ACTION (#976).
//
// A keyword action is a verb the rules define once and cards then
// print by name — proliferate (CR 701.34), scry (CR 701.22), surveil
// (CR 701.25). Three of them carry a COUNT, and a handful of printed
// cards replace that count: Tekuthal, Inquiry Dominus's "if you would
// proliferate, proliferate twice instead", the Crystal Ball family's
// "if you would scry, scry that many plus one instead".
//
// None of them had a seam. `ProliferateForEffect` applied its choice
// directly and `lookAtTopForEffect` queued its prompt directly, so
// nothing in the CR 614 pipeline ever saw either — Tekuthal shipped
// with the clause declared as a caveat.
//
// # One event, opened once per INSTRUCTION
//
// ADR 0061's shape, and for the same reason: "proliferate" is one
// keyword action however many permanents the choice names, and "scry
// 2" is one keyword action however many cards it looks at, so a
// doubler modifies the INSTRUCTION rather than each counter or each
// card. `RepEventKeywordAction` is opened once at the one entry point
// of each action, carries the action and its count, and the action
// happens only once the window settles:
//
//	proliferate   Count is the number of TIMES the action is taken
//	              (CR 701.34's whole application, choice and all).
//	              Base 1; Tekuthal doubles it.
//	scry/surveil  Count is the number of CARDS looked at. Base N;
//	              a "that many plus one" adds to it, and the prompt
//	              is queued with what the window settled on.
//
// One kind with an `Action` discriminator rather than three kinds,
// because everything downstream of the count is identical: the
// gather, the CR 616.1 apply-loop, the #792 identical-window skip and
// the resume are the same code for all three, and a fourth keyword
// action with a count (investigate, explore) is a constant and an arm
// of one switch rather than a new event.
//
// # Identity, and the order a mixed window asks for
//
// Nothing declared. Two Tekuthals are two objects contributing ONE
// declared effect, so #792's identical-window skip applies and nobody
// is asked to order them — ×2 then ×2 is ×4 either way. A doubler
// beside a "plus one" is two DIFFERENT declared effects and the
// proliferating player really is asked, because CR 616.1's orderings
// differ: ×2 then +1 is 3, +1 then ×2 is 4.
//
// # Pausing
//
// That ordering prompt is a pause, so both halves are carried on the
// event the way every other pipeline in the engine carries them — the
// chosen proliferate lists, and the scry prompt's kind and its "then"
// continuation, on `keywordActionTail`. A paused proliferate has
// placed NO counters and a paused scry has queued NO prompt when the
// entry point returns; the resume does both, through the same
// function the unpaused path runs. `ScryThenForEffect`'s returned
// count is therefore 0 on a pause, the contract
// `CreateTokensForEffect`'s empty ID slice already carries.

// KeywordAction names the keyword action a RepEventKeywordAction is
// about. Only the actions that carry a COUNT are here: a count is
// what a replacement rewrites, and a keyword action without one
// (sacrifice, tap, exile) is already an ordinary mutation with its
// own event.
type KeywordAction string

const (
	// KeywordActionProliferate is CR 701.34. Its count is the number
	// of TIMES the action is taken, not a number of counters: "choose
	// any number of permanents and/or players with counters on them,
	// then give each another counter of each kind already there" is
	// one proliferate whatever it chooses, and "proliferate twice"
	// takes the whole action twice.
	KeywordActionProliferate KeywordAction = "proliferate"

	// KeywordActionScry is CR 701.22. Its count is the number of
	// CARDS looked at.
	KeywordActionScry KeywordAction = "scry"

	// KeywordActionSurveil is CR 701.25. Its count is the number of
	// CARDS looked at, as scry's is.
	KeywordActionSurveil KeywordAction = "surveil"
)

// maxKeywordActionRepeats caps how many times one settled instruction
// may take its action. The CR 616.1 apply-loop is bounded at 32
// iterations and a doubler is multiplicative, so a board a card bug
// could build would ask for 2^32 proliferates and hang the table;
// nothing printed asks for more than two.
//
// Over the cap the action is taken maxKeywordActionRepeats times and
// the overflow is logged, which is the posture
// ErrReplacementIterationExceeded takes: weaker than asked for, never
// a wedged game.
const maxKeywordActionRepeats = 64

// keywordActionTail is what the entry point still owes once the
// window settles — the keyword-action sibling of zoneRoute,
// damageTail and tokenTail, and it exists for the same reason they
// do: the window can PAUSE on a CR 616 ordering prompt, and the
// resume has to finish the action exactly as the entry point asked
// for it.
//
// Unexported engine plumbing. The catalog never sets or reads it; a
// replacement rewrites the COUNT on the event and nothing else.
type keywordActionTail struct {
	// cards / players are a proliferate's chosen permanents and
	// players (CR 701.34a). The choice is made BEFORE the window
	// opens — by the catalog's Proliferate primitive, or by whoever
	// called with explicit lists — and the same choice is applied
	// each time the settled count asks for.
	//
	// DECLARED SIMPLIFICATION, and a small one: in paper each
	// proliferate of a "proliferate twice" is its own choice. The
	// catalog's chooser is the deterministic beneficial pick, which
	// only ever chooses things that already have counters and only
	// ever adds a counter of a kind already there, so a second pick
	// over the board the first one left returns the same two lists.
	// See cards/effects/proliferate.go, which carries the choice half
	// of the rule and the reason it is not prompted yet.
	cards   []uuid.UUID
	players []uuid.UUID

	// choice is the prompt a settled scry or surveil queues —
	// PendingChoiceScry or PendingChoiceSurveil. The count on the
	// event is what it is queued with.
	choice PendingChoiceKind

	// then is the rest of the sentence after the keyword action:
	// Preordain's "then draw a card". It runs from wherever the
	// action ends — after the player has put the cards back, or
	// immediately when the action looked at nothing or was replaced
	// away entirely — which is the contract lookAtTopForEffect
	// already documents.
	then func(g *Game) error
}

// ProliferateForEffect takes the proliferate keyword action
// (CR 701.34) on behalf of `player`: it opens the CR 614 window on the
// action and, once the window settles, gives each named permanent and
// each named player one additional counter of every kind they already
// have — once per time the window settled on.
//
// `source` is the card whose effect is proliferating, which is what
// "if YOU would proliferate" reads through (the event's Actor is the
// proliferating player; Source is the card).
//
// Both lists may be empty — "any number" includes zero, and a
// proliferate with nothing worth choosing is a legal no-op rather
// than an error. IDs that name something that has left the
// battlefield, or a player who is not seated, are skipped: the
// choice is made when the effect starts resolving and the board can
// have moved underneath it.
//
// It can PAUSE. A window with a doubler and a "plus one" in it queues
// a CR 616 ordering prompt and returns with NO counters placed; the
// resume proliferates when the prompt is answered.
//
// Caller must hold g.mu (it is an effect-time helper).
func (g *Game) ProliferateForEffect(player, source uuid.UUID, cardIDs, playerIDs []uuid.UUID) error {
	_, err := g.runKeywordActionLocked(&ReplacementEvent{
		Kind:               RepEventKeywordAction,
		Actor:              player,
		Source:             source,
		KeywordAction:      KeywordActionProliferate,
		KeywordActionCount: 1,
		keywordAction: &keywordActionTail{
			cards:   cardIDs,
			players: playerIDs,
		},
	})
	return err
}

// runKeywordActionLocked runs the CR 614 window for ev and, once it
// settles, takes the action the window left. Shared by the three
// entry points and by the CR 616 resume
// (applyResolvedReplacementEventLocked), so a paused keyword action
// and an unpaused one cannot drift apart.
//
// The returned int is the number of cards a scry or surveil is
// actually looking at — zero for a proliferate, and zero for an
// action that paused, was cancelled or found an empty library.
//
// Caller must hold g.mu.
func (g *Game) runKeywordActionLocked(ev *ReplacementEvent) (int, error) {
	out, err := g.applyReplacementsLocked(ev)
	if errors.Is(err, errReplacementPending) {
		// A prompt is queued and the resume owns the action now.
		// NOTHING has happened: no counter was placed, no prompt was
		// queued, and the caller's continuation is still owed.
		return 0, nil
	}
	if err != nil && !errors.Is(err, ErrReplacementIterationExceeded) {
		g.clearReplacementEventLocked(ev.ID)
		return 0, err
	}
	defer g.clearReplacementEventLocked(ev.ID)
	if out == nil || out.Canceled {
		// CR 614.10 with a null replacement: the action simply is not
		// taken. The rest of the sentence still runs — "scry 2, then
		// draw a card" draws whether or not the scry happened.
		return 0, g.abandonKeywordActionLocked(ev)
	}
	return g.applyResolvedKeywordActionLocked(out)
}

// applyResolvedKeywordActionLocked takes the keyword action a settled
// RepEventKeywordAction describes, with the count the window left on
// it.
//
// Caller must hold g.mu.
func (g *Game) applyResolvedKeywordActionLocked(ev *ReplacementEvent) (int, error) {
	tail := ev.keywordAction
	if tail == nil {
		tail = &keywordActionTail{}
	}
	n := ev.KeywordActionCount
	if n <= 0 {
		// Replaced down to nothing. Not a cancellation — the action
		// was taken, it simply had no count left — but there is
		// nothing to do but tell the caller.
		return 0, g.abandonKeywordActionLocked(ev)
	}
	switch ev.KeywordAction {
	case KeywordActionProliferate:
		times := n
		if times > maxKeywordActionRepeats {
			g.EmitEvent(Event{
				Kind:     EventEffectError,
				Actor:    ev.Actor,
				Source:   ev.Source,
				ErrorMsg: "proliferate repeat count capped at " + strconv.Itoa(maxKeywordActionRepeats),
			})
			times = maxKeywordActionRepeats
		}
		for i := 0; i < times; i++ {
			if err := g.applyProliferateLocked(ev.Actor, tail.cards, tail.players); err != nil {
				return 0, err
			}
		}
		return 0, nil
	case KeywordActionScry, KeywordActionSurveil:
		return g.lookAtTopForEffect(tail.choice, ev.Actor, ev.Source, n, tail.then), nil
	}
	// An action kind nothing takes yet. Telling the caller is the only
	// safe answer: a continuation nobody runs waits forever.
	return 0, g.abandonKeywordActionLocked(ev)
}

// abandonKeywordActionLocked is the terminal outcome of a keyword
// action that did NOTHING — cancelled by a CR 614.10 null
// replacement, or replaced down to a count of zero. It runs the rest
// of the sentence with nothing done.
//
// A caller sequencing work behind the action has to be told even when
// the answer is "none", or it waits forever; that is the call #808
// made for the life tail, #853 for the route tail and #762 for the
// token tail. A proliferate carries no continuation and this is a
// no-op for it.
//
// Caller must hold g.mu.
func (g *Game) abandonKeywordActionLocked(ev *ReplacementEvent) error {
	if ev == nil || ev.keywordAction == nil || ev.keywordAction.then == nil {
		return nil
	}
	// Cleared THROUGH the pointer, so a continuation that re-enters
	// the pipeline on the same tail value cannot run itself twice —
	// and so cloneReplacementResume has to give an undo snapshot its
	// own copy of the tail, which it does.
	then := ev.keywordAction.then
	ev.keywordAction.then = nil
	return then(g)
}
