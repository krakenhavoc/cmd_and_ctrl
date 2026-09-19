package game

import (
	"testing"

	"github.com/google/uuid"
)

// keyword_action_test.go — #976, the engine half: a keyword action
// with a count is a CR 614 event, opened once per instruction at the
// one entry point of each action.
//
// The card half (Tekuthal's "proliferate twice", the "scry that many
// plus one" shape, and the CR 616 order between them) is in
// cards/effects/keyword_action_replacements_test.go, where a real
// catalog entry can carry the replacement. Everything here is about
// the pipeline itself, which is why the replacements below are
// test-injected: they have no source card, so they can exercise the
// cancel, the pause and the resume without a card to own them.

// keywordCountReplacement is "if you would <action>, <action>
// count(n) instead" as a test injection.
func keywordCountReplacement(action KeywordAction, count func(int) int, label string) ReplacementEffect {
	return ReplacementEffect{
		Watches: []EventKind{EventKeywordAction},
		AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
			return ev.Kind == RepEventKeywordAction && ev.KeywordAction == action
		},
		Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
			ev.KeywordActionCount = count(ev.KeywordActionCount)
			return nil
		},
		Label: label,
	}
}

// keywordCancelReplacement is the CR 614.10 null replacement: the
// action simply is not taken.
func keywordCancelReplacement(action KeywordAction) ReplacementEffect {
	return ReplacementEffect{
		Watches: []EventKind{EventKeywordAction},
		AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
			return ev.Kind == RepEventKeywordAction && ev.KeywordAction == action
		},
		Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
			ev.Cancel()
			return nil
		},
		Label: "no " + string(action),
	}
}

// seedCounterPermanent parks a permanent carrying `n` +1/+1 counters,
// which is the minimum a proliferate needs to have something to
// choose (CR 701.34a).
func seedCounterPermanent(g *Game, owner uuid.UUID, n int) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: id,
		Name:       "Counterful Bear",
		TypeLine:   "Creature — Bear",
		Owner:      owner,
		Controller: owner,
		Counters:   map[string]int{CounterPlusOne: n},
	})
	return id
}

// --- proliferate -----------------------------------------------------

// The floor: with nothing watching, a proliferate happens once.
func TestProliferateHappensOnceWithNoReplacement(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	bear := seedCounterPermanent(g, me.ID, 1)

	g.WithWriteLock(func() {
		if err := g.ProliferateForEffect(me.ID, uuid.Nil, []uuid.UUID{bear}, nil); err != nil {
			t.Fatalf("ProliferateForEffect: %v", err)
		}
	})
	if got := counterOn(g, bear, CounterPlusOne); got != 2 {
		t.Errorf("+1/+1 counters = %d, want 2", got)
	}
}

// "If you would proliferate, proliferate twice instead" is a count of
// TIMES: the whole action runs twice, so a permanent with one counter
// ends with three.
func TestProliferateTwiceTakesTheActionTwice(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	bear := seedCounterPermanent(g, me.ID, 1)

	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(keywordCountReplacement(KeywordActionProliferate, timesTwoForTest, "twice"))
		if err := g.ProliferateForEffect(me.ID, uuid.Nil, []uuid.UUID{bear}, nil); err != nil {
			t.Fatalf("ProliferateForEffect: %v", err)
		}
	})
	if got := counterOn(g, bear, CounterPlusOne); got != 3 {
		t.Errorf("+1/+1 counters = %d, want 3 (one printed, two proliferates)", got)
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("one replacement queued %d prompts, want none", len(g.PendingChoices))
	}
}

// A cancelled keyword action does nothing at all — no counter, no
// error, and the once-per-event map is cleaned up behind it.
func TestCancelledProliferateDoesNothing(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	bear := seedCounterPermanent(g, me.ID, 1)

	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(keywordCancelReplacement(KeywordActionProliferate))
		if err := g.ProliferateForEffect(me.ID, uuid.Nil, []uuid.UUID{bear}, nil); err != nil {
			t.Fatalf("ProliferateForEffect: %v", err)
		}
	})
	if got := counterOn(g, bear, CounterPlusOne); got != 1 {
		t.Errorf("+1/+1 counters = %d, want 1 — the proliferate was replaced away", got)
	}
	if len(g.replacementsAppliedThisEvent) != 0 {
		t.Errorf("once-per-event map still holds %d entries", len(g.replacementsAppliedThisEvent))
	}
}

// The count is the count the window SETTLED on, so a replacement that
// takes it to zero takes the action off the board as surely as a
// cancel does.
func TestProliferateReplacedToZeroDoesNothing(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	bear := seedCounterPermanent(g, me.ID, 1)

	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(keywordCountReplacement(KeywordActionProliferate, zeroForTest, "never"))
		if err := g.ProliferateForEffect(me.ID, uuid.Nil, []uuid.UUID{bear}, nil); err != nil {
			t.Fatalf("ProliferateForEffect: %v", err)
		}
	})
	if got := counterOn(g, bear, CounterPlusOne); got != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", got)
	}
}

// CR 616.1: two DIFFERENT effects in one window is an ordering prompt,
// and the orders differ — ×2 then +1 is three proliferates, +1 then ×2
// is four. Nothing is placed until the order is answered.
func TestProliferateDoublerAndPlusOnePromptForOrder(t *testing.T) {
	for _, tc := range []struct {
		name    string
		order   []string // the answered order, by prompt label
		counter int      // counters on a permanent that started with one
	}{
		{"double then plus one", []string{"twice", "one more"}, 1 + 3},
		{"plus one then double", []string{"one more", "twice"}, 1 + 4},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			me := g.Seats[0]
			bear := seedCounterPermanent(g, me.ID, 1)

			g.WithWriteLock(func() {
				g.RegisterReplacementForTest(keywordCountReplacement(KeywordActionProliferate, timesTwoForTest, "twice"))
				g.RegisterReplacementForTest(keywordCountReplacement(KeywordActionProliferate, plusOneForTest, "one more"))
				if err := g.ProliferateForEffect(me.ID, uuid.Nil, []uuid.UUID{bear}, nil); err != nil {
					t.Fatalf("ProliferateForEffect: %v", err)
				}
			})
			if got := counterOn(g, bear, CounterPlusOne); got != 1 {
				t.Fatalf("+1/+1 counters = %d before the order is answered, want 1", got)
			}
			if len(g.PendingChoices) != 1 {
				t.Fatalf("pending choices = %d, want the one CR 616 prompt", len(g.PendingChoices))
			}
			p := g.PendingChoices[0]
			if p.Kind != PendingChoiceReplacementOrder {
				t.Fatalf("prompt kind = %q, want %q", p.Kind, PendingChoiceReplacementOrder)
			}
			if err := g.ResolveReplacementOrder(p.ID, p.Chooser, orderByLabel(t, g, p, tc.order...)); err != nil {
				t.Fatalf("ResolveReplacementOrder: %v", err)
			}
			if got := counterOn(g, bear, CounterPlusOne); got != tc.counter {
				t.Errorf("+1/+1 counters = %d, want %d", got, tc.counter)
			}
		})
	}
}

// Undo across the window: the snapshot is taken while the CR 616
// prompt is open, the live game answers it, and the restore puts the
// table back with the prompt unanswered and no counter placed. The
// replayed answer then lands exactly what the first one did.
func TestUndoAcrossAKeywordActionPrompt(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	bear := seedCounterPermanent(g, me.ID, 1)

	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(keywordCountReplacement(KeywordActionProliferate, timesTwoForTest, "twice"))
		g.RegisterReplacementForTest(keywordCountReplacement(KeywordActionProliferate, plusOneForTest, "one more"))
		if err := g.ProliferateForEffect(me.ID, uuid.Nil, []uuid.UUID{bear}, nil); err != nil {
			t.Fatalf("ProliferateForEffect: %v", err)
		}
	})
	if len(g.PendingChoices) != 1 {
		t.Fatalf("pending choices = %d, want the one CR 616 prompt", len(g.PendingChoices))
	}
	snapshot := g.Clone()

	p := g.PendingChoices[0]
	order := orderByLabel(t, g, p, "twice", "one more")
	if err := g.ResolveReplacementOrder(p.ID, p.Chooser, order); err != nil {
		t.Fatalf("ResolveReplacementOrder: %v", err)
	}
	if got := counterOn(g, bear, CounterPlusOne); got != 4 {
		t.Fatalf("+1/+1 counters after the answer = %d, want 4", got)
	}

	g.RestoreFrom(snapshot)
	if got := counterOn(g, bear, CounterPlusOne); got != 1 {
		t.Errorf("+1/+1 counters after the undo = %d, want 1", got)
	}
	if len(g.PendingChoices) != 1 {
		t.Fatalf("pending choices after the undo = %d, want the prompt back", len(g.PendingChoices))
	}
	p = g.PendingChoices[0]
	if err := g.ResolveReplacementOrder(p.ID, p.Chooser, orderByLabel(t, g, p, "twice", "one more")); err != nil {
		t.Fatalf("replayed ResolveReplacementOrder: %v", err)
	}
	if got := counterOn(g, bear, CounterPlusOne); got != 4 {
		t.Errorf("+1/+1 counters after the replayed answer = %d, want 4", got)
	}
}

// The paused keyword action rides the frame every other paused
// replacement rides, so the persisted snapshot needs no new field: it
// drops the resume and COUNTS it, which is what keeps the snapshot
// honest about not being a restore point.
func TestPausedKeywordActionIsCountedByTheSnapshotCensus(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	bear := seedCounterPermanent(g, me.ID, 1)

	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(keywordCountReplacement(KeywordActionProliferate, timesTwoForTest, "twice"))
		g.RegisterReplacementForTest(keywordCountReplacement(KeywordActionProliferate, plusOneForTest, "one more"))
		if err := g.ProliferateForEffect(me.ID, uuid.Nil, []uuid.UUID{bear}, nil); err != nil {
			t.Fatalf("ProliferateForEffect: %v", err)
		}
	})
	snap := g.CaptureSnapshot()
	if snap.Continuations.ChoiceResumeFrames == 0 {
		t.Errorf("the paused keyword action's resume frame is not counted: %+v", snap.Continuations)
	}
	if snap.Continuations.Empty() {
		t.Error("a snapshot holding a paused replacement must not report an empty census")
	}
}

// --- scry and surveil ------------------------------------------------

// "If you would scry, scry that many plus one instead": the prompt is
// queued with the count the window settled on, and the banner says so.
func TestScryPlusOneQueuesTheBiggerPrompt(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]

	var looked int
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(keywordCountReplacement(KeywordActionScry, plusOneForTest, "scry one more"))
		looked = g.ScryForEffect(me.ID, uuid.Nil, 2)
	})
	if looked != 3 {
		t.Errorf("looked at %d cards, want 3", looked)
	}
	if len(g.PendingChoices) != 1 {
		t.Fatalf("pending choices = %d, want the scry prompt", len(g.PendingChoices))
	}
	p := g.PendingChoices[0]
	if p.Kind != PendingChoiceScry {
		t.Fatalf("prompt kind = %q, want %q", p.Kind, PendingChoiceScry)
	}
	if len(p.ScryCards) != 3 || p.Count != 3 {
		t.Errorf("prompt offers %d cards (count %d), want 3", len(p.ScryCards), p.Count)
	}
	if p.Reason != "Scry 3" {
		t.Errorf("prompt banner = %q, want %q — the copy follows the settled count", p.Reason, "Scry 3")
	}
}

// Surveil is the same action with the other away lane, and the same
// window.
func TestSurveilPlusOneQueuesTheBiggerPrompt(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]

	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(keywordCountReplacement(KeywordActionSurveil, plusOneForTest, "surveil one more"))
		g.SurveilThenForEffect(me.ID, uuid.Nil, 1, nil)
	})
	if len(g.PendingChoices) != 1 {
		t.Fatalf("pending choices = %d, want the surveil prompt", len(g.PendingChoices))
	}
	p := g.PendingChoices[0]
	if p.Kind != PendingChoiceSurveil || len(p.ScryCards) != 2 {
		t.Errorf("prompt = %q over %d cards, want a surveil over 2", p.Kind, len(p.ScryCards))
	}
}

// A scry replacement must not fire on a surveil, and vice versa: one
// event kind, three actions, and the action is on the event.
func TestScryReplacementDoesNotTouchSurveil(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]

	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(keywordCountReplacement(KeywordActionScry, plusOneForTest, "scry one more"))
		g.SurveilThenForEffect(me.ID, uuid.Nil, 1, nil)
	})
	if len(g.PendingChoices) != 1 || len(g.PendingChoices[0].ScryCards) != 1 {
		t.Fatalf("a scry replacement changed a surveil: %+v", g.PendingChoices)
	}
}

// "Look at the top N cards of your library, then put them back in any
// order" is not a keyword action, so there is nothing to replace.
func TestLookAtTopIsNotAKeywordAction(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]

	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(keywordCountReplacement(KeywordActionScry, plusOneForTest, "scry one more"))
		g.RegisterReplacementForTest(keywordCountReplacement(KeywordActionSurveil, plusOneForTest, "surveil one more"))
		g.LookAtTopThenForEffect(me.ID, uuid.Nil, 2, nil)
	})
	if len(g.PendingChoices) != 1 || len(g.PendingChoices[0].ScryCards) != 2 {
		t.Fatalf("a keyword-action replacement reached look-at-top: %+v", g.PendingChoices)
	}
}

// A cancelled scry queues nothing — and still runs the rest of the
// sentence. Preordain's "then draw a card" is not conditional on the
// scry having happened.
func TestCancelledScryQueuesNothingAndStillRunsThen(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]

	ran := 0
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(keywordCancelReplacement(KeywordActionScry))
		g.ScryThenForEffect(me.ID, uuid.Nil, 2, func(*Game) error {
			ran++
			return nil
		})
	})
	if len(g.PendingChoices) != 0 {
		t.Errorf("a cancelled scry queued %d prompts, want none", len(g.PendingChoices))
	}
	if ran != 1 {
		t.Errorf("the continuation ran %d times, want exactly 1", ran)
	}
}

// A scry that PAUSES has queued nothing yet and reports nothing
// looked at; the resume queues the prompt with the settled count when
// the order is answered.
func TestPausedScryQueuesItsPromptFromTheResume(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]

	ran := 0
	var looked int
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(keywordCountReplacement(KeywordActionScry, plusOneForTest, "one more"))
		g.RegisterReplacementForTest(keywordCountReplacement(KeywordActionScry, timesTwoForTest, "twice"))
		looked = g.ScryThenForEffect(me.ID, uuid.Nil, 1, func(*Game) error {
			ran++
			return nil
		})
	})
	if looked != 0 {
		t.Errorf("a paused scry reported %d cards looked at, want 0", looked)
	}
	if len(g.PendingChoices) != 1 || g.PendingChoices[0].Kind != PendingChoiceReplacementOrder {
		t.Fatalf("want the one CR 616 ordering prompt, got %+v", g.PendingChoices)
	}
	if ran != 0 {
		t.Fatalf("the continuation ran before the scry did")
	}

	p := g.PendingChoices[0]
	// +1 then ×2 over a printed 1: scry 4.
	if err := g.ResolveReplacementOrder(p.ID, p.Chooser, orderByLabel(t, g, p, "one more", "twice")); err != nil {
		t.Fatalf("ResolveReplacementOrder: %v", err)
	}
	if len(g.PendingChoices) != 1 || g.PendingChoices[0].Kind != PendingChoiceScry {
		t.Fatalf("the resume did not queue the scry: %+v", g.PendingChoices)
	}
	if got := len(g.PendingChoices[0].ScryCards); got != 4 {
		t.Errorf("the resumed scry looks at %d cards, want 4 ((1+1)*2)", got)
	}
	if ran != 0 {
		t.Errorf("the continuation ran before the player answered the scry")
	}
}

func timesTwoForTest(n int) int { return n * 2 }

func plusOneForTest(n int) int { return n + 1 }

func zeroForTest(int) int { return 0 }
