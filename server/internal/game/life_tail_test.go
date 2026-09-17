package game

import (
	"reflect"
	"testing"

	"github.com/google/uuid"
)

// life_tail_test.go is the #482 regression suite. Two things are being
// pinned:
//
//  1. WHO RUNS THE WINDOW. Every writer of a life total — the public
//     sandbox verb, ChangePlayerLifeForEffect (every catalog GainLife,
//     drain and pay-life cost) and the CR 702.15b lifelink credit —
//     runs the CR 614 replacement window. Damage does NOT: it reduces
//     life directly once the DAMAGE replacements have settled
//     (CR 120.3), so a life replacement must not see it.
//
//  2. WHERE IT LANDS. When two life replacements apply, CR 616 asks
//     the affected player to order them and the change lands from the
//     resume rather than from the entry point. Following
//     damage_tail_test.go, every scenario is run twice — paused, and
//     with one combined replacement that never prompts — and the whole
//     observable outcome is compared.

// --- who runs the window --------------------------------------------

// doubleGainReplacement is Rhox Faithmender in one line: a positive
// life change becomes twice as big. Registered per test so the effect
// path, the lifelink path and the damage path can each be asked
// whether the window ran.
func doubleGainReplacement() ReplacementEffect {
	return ReplacementEffect{
		Watches: []EventKind{EventChangeLife},
		AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
			return ev.Kind == RepEventLife && ev.LifeDelta > 0
		},
		Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
			ev.LifeDelta *= 2
			return nil
		},
		Label: "gain twice that much",
	}
}

// lifeEventRecorder captures EVERY EventChangeLife the engine emits,
// so a test can assert on what a "whenever you gain life" or "whenever
// you lose life" trigger would see rather than only on the resulting
// total. noncombat_keywords_test.go's recorder keeps the positive ones
// only, which is the right filter for a damage path and the wrong one
// here.
type lifeEventRecorder struct{ events []Event }

func (r *lifeEventRecorder) OnEvent(_ *Game, ev Event) {
	if ev.Kind == EventChangeLife {
		r.events = append(r.events, ev)
	}
}

// TestChangePlayerLifeForEffectRunsTheLifeWindow — the headline bug.
// Every catalog GainLife goes through this entry point and it used to
// call ChangeLife directly, so a life-change replacement only ever saw
// a life total typed in by hand.
func TestChangePlayerLifeForEffectRunsTheLifeWindow(t *testing.T) {
	g := newActiveGame(t)
	seen := &lifeEventRecorder{}
	src := uuid.New()
	g.WithWriteLock(func() {
		g.Listeners = append(g.Listeners, seen)
		g.RegisterReplacementForTest(doubleGainReplacement())
		if err := g.ChangePlayerLifeForEffect(src, g.Seats[0].ID, 3); err != nil {
			t.Fatalf("ChangePlayerLifeForEffect: %v", err)
		}
	})
	if got := g.Seats[0].Life; got != StartingLife+6 {
		t.Errorf("life = %d, want %d — a catalog GainLife of 3 under a doubler is 6", got, StartingLife+6)
	}
	if len(seen.events) != 1 {
		t.Fatalf("%d EventChangeLife emitted, want exactly 1", len(seen.events))
	}
	ev := seen.events[0]
	if ev.Amount != 6 {
		t.Errorf("Amount = %d, want 6 — the event a lifegain trigger sees is the REPLACED amount", ev.Amount)
	}
	if ev.Target != g.Seats[0].ID || ev.Source != src {
		t.Errorf("Target/Source = %s/%s, want %s/%s", ev.Target, ev.Source, g.Seats[0].ID, src)
	}
}

// TestLifeLossThroughTheEffectPathRunsTheWindowToo — a drain and a
// pay-life cost are the same entry point, and a life-loss replacement
// (Bloodletter of Aclazotz) watches the same event. The doubler here
// declines negative deltas, which is exactly what Rhox Faithmender's
// AppliesTo does, so the loss lands untouched.
func TestLifeLossThroughTheEffectPathRunsTheWindowToo(t *testing.T) {
	g := newActiveGame(t)
	var saw []int
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(ReplacementEffect{
			Watches: []EventKind{EventChangeLife},
			AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
				if ev.Kind != RepEventLife {
					return false
				}
				saw = append(saw, ev.LifeDelta)
				return ev.LifeDelta < 0
			},
			Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
				ev.LifeDelta *= 2
				return nil
			},
			Label: "lose twice that much",
		})
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, g.Seats[0].ID, -2); err != nil {
			t.Fatalf("ChangePlayerLifeForEffect: %v", err)
		}
	})
	if len(saw) == 0 {
		t.Fatal("a life LOSS through the effect path never reached the CR 614 window")
	}
	if got := g.Seats[0].Life; got != StartingLife-4 {
		t.Errorf("life = %d, want %d", got, StartingLife-4)
	}
}

// TestLifelinkCreditRunsTheLifeWindow — CR 702.15b makes lifelink life
// GAIN, so it is replaceable. It used to write the total directly,
// which is why Rhox Faithmender did not double its own lifelink even
// though its card comment said it did (#482, the 2026-09-16 note).
func TestLifelinkCreditRunsTheLifeWindow(t *testing.T) {
	atkID := uuid.New()
	g := newActiveGame(t)
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(doubleGainReplacement())
		pushTestCreature(g, atkID, g.Seats[0], 3, 3, "lifelink")
		g.markCombatDamageToPlayerLocked(g.Seats[1].ID, atkID, 3, "")
	})
	if got := g.Seats[0].Life; got != StartingLife+6 {
		t.Errorf("attacker's controller life = %d, want %d — 3 lifelink doubled to 6",
			got, StartingLife+6)
	}
	if got := g.Seats[1].Life; got != StartingLife-3 {
		t.Errorf("defender life = %d, want %d — the DAMAGE is untouched by a life replacement",
			got, StartingLife-3)
	}
}

// TestDamageDoesNotFireTheLifeWindow — the split that keeps the fix
// honest. Damage to a player reduces life inside the damage tail once
// the DAMAGE replacements have settled (CR 120.3); it is not a second,
// separately replaceable life-change event. If it were, every doubler
// in the lifegain deck would also double every Lightning Bolt aimed at
// its controller.
func TestDamageDoesNotFireTheLifeWindow(t *testing.T) {
	fired := 0
	g := newActiveGame(t)
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(ReplacementEffect{
			Watches: []EventKind{EventChangeLife},
			AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
				return ev.Kind == RepEventLife
			},
			Replace: func(*ReplacementEvent, *Game, *Card) error {
				fired++
				return nil
			},
			Label: "count life events",
		})
		if err := g.DealDamageToPlayerForEffect(uuid.Nil, g.Seats[1].ID, 3); err != nil {
			t.Fatalf("DealDamageToPlayerForEffect: %v", err)
		}
	})
	if fired != 0 {
		t.Errorf("a life replacement fired %d times on a damage event, want 0 — "+
			"damage reduces life directly after the DAMAGE replacements (CR 120.3)", fired)
	}
	if got := g.Seats[1].Life; got != StartingLife-3 {
		t.Errorf("life = %d, want %d", got, StartingLife-3)
	}
}

// TestCanceledLifeChangeEmitsNothing — CR 614.10 with a null
// replacement ("your life total can't change", Sulfuric Vortex's
// "can't gain"). No mutation and no event, because nothing happened:
// a "whenever you gain life" trigger must not fire on a gain that was
// replaced away.
func TestCanceledLifeChangeEmitsNothing(t *testing.T) {
	g := newActiveGame(t)
	seen := &lifeEventRecorder{}
	g.WithWriteLock(func() {
		g.Listeners = append(g.Listeners, seen)
		g.RegisterReplacementForTest(ReplacementEffect{
			Watches: []EventKind{EventChangeLife},
			AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
				return ev.Kind == RepEventLife && ev.LifeDelta > 0
			},
			Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
				ev.Cancel()
				return nil
			},
			Label: "you can't gain life",
		})
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, g.Seats[0].ID, 4); err != nil {
			t.Fatalf("ChangePlayerLifeForEffect: %v", err)
		}
	})
	if got := g.Seats[0].Life; got != StartingLife {
		t.Errorf("life = %d, want %d — the gain was replaced away", got, StartingLife)
	}
	if len(seen.events) != 0 {
		t.Errorf("%d EventChangeLife emitted, want 0 — a canceled gain is not a gain", len(seen.events))
	}
}

// --- where it lands: the CR 616 pause -------------------------------

// lifeScenario is one life change, applied to a board. Mirrors
// damageScenario in damage_tail_test.go.
type lifeScenario struct {
	// setUp stocks the board. Runs under the write lock.
	setUp func(g *Game)
	// change fires the life event. Runs under the write lock.
	change func(g *Game)
}

// runLifeScenario plays sc out on a fresh two-seat game. When paused is
// true it registers two replacements so CR 616 prompts, then answers
// the prompt in the order offered; otherwise it registers one combined
// replacement doing the same arithmetic and nothing prompts.
//
// The unpaused run gets an explicit state-check sweep at the end,
// because the paused run's answer is an action boundary and gets one
// from the resume. That is the ONLY difference the two paths are
// allowed to have.
func runLifeScenario(t *testing.T, sc lifeScenario, paused bool) (*Game, *lifeEventRecorder) {
	t.Helper()
	g := newActiveGame(t)
	seen := &lifeEventRecorder{}

	g.WithWriteLock(func() {
		g.Listeners = append(g.Listeners, seen)
		if paused {
			// Two applicable replacements on one event is what makes
			// CR 616 prompt. Registration order is gather order, so
			// the prompt offers plus-one-then-double and answering it
			// as offered gives (d+1)*2.
			g.RegisterReplacementForTest(ReplacementEffect{
				Watches:   []EventKind{EventChangeLife},
				AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool { return ev.Kind == RepEventLife },
				Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
					ev.LifeDelta++
					return nil
				},
				Label: "Plus 1",
			})
			g.RegisterReplacementForTest(ReplacementEffect{
				Watches:   []EventKind{EventChangeLife},
				AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool { return ev.Kind == RepEventLife },
				Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
					ev.LifeDelta *= 2
					return nil
				},
				Label: "Double",
			})
		} else {
			g.RegisterReplacementForTest(ReplacementEffect{
				Watches:   []EventKind{EventChangeLife},
				AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool { return ev.Kind == RepEventLife },
				Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
					ev.LifeDelta = (ev.LifeDelta + 1) * 2
					return nil
				},
				Label: "Plus 1 then double",
			})
		}
		sc.setUp(g)
		sc.change(g)
	})

	if !paused {
		if len(g.PendingChoices) != 0 {
			t.Fatalf("unpaused run queued %d choices, want 0", len(g.PendingChoices))
		}
		g.WithWriteLock(g.runStateChecksLocked)
		return g, seen
	}

	if len(g.PendingChoices) != 1 {
		t.Fatalf("paused run queued %d choices, want 1 CR 616 ordering prompt", len(g.PendingChoices))
	}
	prompt := g.PendingChoices[0]
	if prompt.Kind != PendingChoiceReplacementOrder {
		t.Fatalf("prompt kind = %q, want %q", prompt.Kind, PendingChoiceReplacementOrder)
	}
	if err := g.ResolveReplacementOrder(prompt.ID, prompt.Chooser, prompt.ReplacementEffectIDs); err != nil {
		t.Fatalf("ResolveReplacementOrder: %v — the answer must not fail the action", err)
	}
	if len(g.PendingChoices) != 0 {
		t.Fatalf("%d choices still queued after the answer, want 0", len(g.PendingChoices))
	}
	return g, seen
}

// lifeChange is one EventChangeLife with its player expressed as a
// seat index, so two independently-created games are comparable. The
// source is a caller-supplied ID and compares directly.
type lifeChange struct {
	Seat   int
	Amount int
	Source uuid.UUID
}

// lifeOutcome is everything a life change can be observed to have done:
// the totals, who is still in the game, and the events a lifegain or
// life-loss trigger would have keyed off.
type lifeOutcome struct {
	Life       []int
	Eliminated []bool
	Changes    []lifeChange
}

func observeLife(g *Game, seen *lifeEventRecorder) lifeOutcome {
	var out lifeOutcome
	seat := map[uuid.UUID]int{}
	for i, p := range g.Seats {
		seat[p.ID] = i
		out.Life = append(out.Life, p.Life)
		out.Eliminated = append(out.Eliminated, p.Eliminated)
	}
	for _, ev := range seen.events {
		idx, ok := seat[ev.Target]
		if !ok {
			idx = -1
		}
		out.Changes = append(out.Changes, lifeChange{Seat: idx, Amount: ev.Amount, Source: ev.Source})
	}
	return out
}

// assertPausedLifeMatchesUnpaused is the heart of this file: a life
// change that paused for a CR 616 prompt has to end exactly where one
// that never paused would, down to the event a trigger sees.
func assertPausedLifeMatchesUnpaused(t *testing.T, sc lifeScenario) *Game {
	t.Helper()
	pausedGame, pausedSeen := runLifeScenario(t, sc, true)
	unpausedGame, unpausedSeen := runLifeScenario(t, sc, false)
	got, want := observeLife(pausedGame, pausedSeen), observeLife(unpausedGame, unpausedSeen)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("a paused life change ended somewhere else than an unpaused one:\n paused: %+v\n direct: %+v", got, want)
	}
	return pausedGame
}

// TestPausedEffectLifeChangeLandsLikeAnUnpausedOne — the catalog path.
func TestPausedEffectLifeChangeLandsLikeAnUnpausedOne(t *testing.T) {
	src := uuid.New()
	g := assertPausedLifeMatchesUnpaused(t, lifeScenario{
		setUp: func(*Game) {},
		change: func(g *Game) {
			_ = g.ChangePlayerLifeForEffect(src, g.Seats[0].ID, 3)
		},
	})
	if got := g.Seats[0].Life; got != StartingLife+8 {
		t.Errorf("life = %d, want %d — (3+1)*2 = 8", got, StartingLife+8)
	}
}

// TestPausedSandboxLifeChangeLandsLikeAnUnpausedOne — the public verb.
// It has run the window since S17, but it landed the paused case
// through a hand-rolled copy of the tail that emitted no Source.
func TestPausedSandboxLifeChangeLandsLikeAnUnpausedOne(t *testing.T) {
	g := assertPausedLifeMatchesUnpaused(t, lifeScenario{
		setUp: func(*Game) {},
		change: func(g *Game) {
			// The public mutator takes the lock itself, so the
			// scenario's locked callback uses the same body it wraps.
			ev := &ReplacementEvent{Kind: RepEventLife, LifePlayer: g.Seats[0].ID, LifeDelta: 3}
			_, _ = g.changeLifeThroughReplacementsLocked(ev)
		},
	})
	if got := g.Seats[0].Life; got != StartingLife+8 {
		t.Errorf("life = %d, want %d", got, StartingLife+8)
	}
}

// TestPausedLifelinkCreditLandsLikeAnUnpausedOne — the lifelink credit
// pauses in the middle of a damage event, which is the most fragile of
// the three: the damage has already landed and only the life gain is
// waiting on the prompt.
func TestPausedLifelinkCreditLandsLikeAnUnpausedOne(t *testing.T) {
	atkID := uuid.New()
	g := assertPausedLifeMatchesUnpaused(t, lifeScenario{
		setUp: func(g *Game) {
			pushTestCreature(g, atkID, g.Seats[0], 3, 3, "lifelink")
		},
		change: func(g *Game) {
			g.markCombatDamageToPlayerLocked(g.Seats[1].ID, atkID, 3, "")
		},
	})
	if got := g.Seats[0].Life; got != StartingLife+8 {
		t.Errorf("attacker's controller life = %d, want %d — (3+1)*2 lifelink", got, StartingLife+8)
	}
	if got := g.Seats[1].Life; got != StartingLife-3 {
		t.Errorf("defender life = %d, want %d — the damage is not replaced twice", got, StartingLife-3)
	}
}

// TestPausedLifeChangeWhosePlayerLeftDoesNotFailTheAction — the
// counterpart of TestPausedDamageWhoseTargetLeftDoesNotFailTheAction.
// ResolveReplacementOrder dequeues the prompt before it dispatches, so
// an error from the tail takes the prompt away AND fails the action.
func TestPausedLifeChangeWhosePlayerLeftDoesNotFailTheAction(t *testing.T) {
	g := newActiveGame(t)
	victim := g.Seats[1].ID
	g.WithWriteLock(func() {
		for i := 0; i < 2; i++ {
			g.RegisterReplacementForTest(ReplacementEffect{
				Watches:   []EventKind{EventChangeLife},
				AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool { return ev.Kind == RepEventLife },
				Replace:   func(*ReplacementEvent, *Game, *Card) error { return nil },
				Label:     "no-op",
			})
		}
		_ = g.ChangePlayerLifeForEffect(uuid.Nil, victim, 3)
	})
	if len(g.PendingChoices) != 1 {
		t.Fatalf("pending choices = %d, want 1", len(g.PendingChoices))
	}
	prompt := g.PendingChoices[0]

	// The player leaves the table before the chooser answers.
	g.WithWriteLock(func() {
		g.Seats = g.Seats[:1]
	})

	if err := g.ResolveReplacementOrder(prompt.ID, prompt.Chooser, prompt.ReplacementEffectIDs); err != nil {
		t.Errorf("ResolveReplacementOrder = %v, want nil — the prompt is already dequeued, so "+
			"a vanished player must not fail the action too", err)
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("%d choices still queued, want 0", len(g.PendingChoices))
	}
}
