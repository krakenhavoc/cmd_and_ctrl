package game

import (
	"testing"

	"github.com/google/uuid"
)

// mill_amount_test.go — #569, the engine half. The catalog half (the
// real Bruvac the Grandiloquent and The Water Crystal) is in
// cards/effects/mill_replacements_test.go, where a card can carry the
// replacement and be scoped to an opponent.
//
// What is here is the engine's own notion of a mill AMOUNT: one CR 614
// window per INSTRUCTION, opened before any card leaves the library,
// only for something the rules call a mill, and able to pause — which
// nothing about a mill could do before a card had already been chosen.
//
// The per-card window is a different thing and was never missing; see
// milled_this_way_test.go, which pins it.

// millAmountReplacement is "if a player would mill, they mill count(n)
// instead" as a test injection. `owner` controls it and is deliberately
// NOT the milling player in most of what follows, because the affected
// player of a mill is the one whose library is read (#982).
func millAmountReplacement(owner uuid.UUID, count func(int) int, label string) ReplacementEffect {
	return ReplacementEffect{
		Watches: []EventKind{EventMill},
		AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
			return ev.Kind == RepEventMill && ev.MillCount > 0
		},
		Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
			ev.MillCount = count(ev.MillCount)
			return nil
		},
		Controller: func(*ReplacementEvent, *Game, *Card) uuid.UUID { return owner },
		Label:      label,
	}
}

// millCancelReplacement is the CR 614.10 null replacement: the mill
// simply does not happen.
func millCancelReplacement(label string) ReplacementEffect {
	return ReplacementEffect{
		Watches: []EventKind{EventMill},
		AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
			return ev.Kind == RepEventMill
		},
		Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
			ev.Cancel()
			return nil
		},
		Label: label,
	}
}

// stockLibrary REPLACES p's library with exactly n plain cards and
// returns them top-first — the order a mill takes them in. Exactly n,
// because half of what follows counts what is left on the library or
// runs it out on purpose, and the opening deal would make both
// arithmetic about the fixture rather than about the mill.
func stockLibrary(p *Player, n int) []uuid.UUID {
	p.Library.Cards = nil
	// PushTop appends, so the first push is the deepest card; read the
	// IDs back reversed to get the order a mill takes them in.
	ids := make([]uuid.UUID, 0, n)
	for i := 0; i < n; i++ {
		ids = append(ids, libraryCard(p, "Stock Bear"))
	}
	// ids is bottom-to-top; reverse into top-first.
	for i, j := 0, len(ids)-1; i < j; i, j = i+1, j-1 {
		ids[i], ids[j] = ids[j], ids[i]
	}
	return ids
}

// millN runs a fire-and-forget mill under the write lock.
func millN(t *testing.T, g *Game, p *Player, n int) []uuid.UUID {
	t.Helper()
	var got []uuid.UUID
	g.WithWriteLock(func() {
		var err error
		got, err = g.MillToZoneForEffect(p.ID, n, ZoneGraveyard)
		if err != nil {
			t.Fatalf("MillToZoneForEffect: %v", err)
		}
	})
	return got
}

// --- the floor -------------------------------------------------------

// With nothing watching, a mill of three mills three.
func TestMillWithNoReplacementMillsWhatItAsksFor(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	g.WithWriteLock(func() { stockLibrary(p, 6) })

	got := millN(t, g, p, 3)

	if len(got) != 3 {
		t.Errorf("milled %d cards, want 3", len(got))
	}
	if len(p.Graveyard.Cards) != 3 {
		t.Errorf("graveyard holds %d, want 3", len(p.Graveyard.Cards))
	}
}

// --- the amount is replaceable ---------------------------------------

// Bruvac's sentence, as the engine sees it: one event per instruction,
// so "mill three" with a doubler out is one event that becomes six —
// not three events of one.
func TestADoublerDoublesTheWholeMillInstruction(t *testing.T) {
	g := newActiveGame(t)
	p, them := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() {
		stockLibrary(p, 12)
		g.RegisterReplacementForTest(millAmountReplacement(them.ID, func(n int) int { return n * 2 }, "twice"))
	})

	got := millN(t, g, p, 3)

	if len(got) != 6 {
		t.Errorf("milled %d cards, want 6 — the instruction is one event and the doubler rewrites its count", len(got))
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("one applicable replacement queued %d prompts, want none (#792)", len(g.PendingChoices))
	}
}

// The count the window sees is the number the INSTRUCTION asked for,
// not what the library can supply: CR 701.13b's "as many as possible"
// clamp happens after the window settles. A twelve-card mill of a
// five-card library mills five, and a doubler still doubles twelve.
func TestTheReplacedCountIsTheInstructionsNotTheLibrarys(t *testing.T) {
	g := newActiveGame(t)
	p, them := g.Seats[0], g.Seats[1]
	seen := 0
	g.WithWriteLock(func() {
		stockLibrary(p, 5)
		g.RegisterReplacementForTest(ReplacementEffect{
			Watches: []EventKind{EventMill},
			AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
				return ev.Kind == RepEventMill
			},
			Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
				seen = ev.MillCount
				return nil
			},
			Controller: func(*ReplacementEvent, *Game, *Card) uuid.UUID { return them.ID },
			Label:      "watcher",
		})
	})

	got := millN(t, g, p, 12)

	if seen != 12 {
		t.Errorf("the window saw a count of %d, want the instruction's 12", seen)
	}
	if len(got) != 5 {
		t.Errorf("milled %d cards off a five-card library, want 5 (CR 701.13b)", len(got))
	}
	if p.AttemptedEmptyDraw {
		t.Error("milling out set the empty-draw flag — only a DRAW from an empty library loses (#767)")
	}
}

// A count replaced down to zero mills nothing and is not an error.
func TestAMillReplacedToZeroMillsNothing(t *testing.T) {
	g := newActiveGame(t)
	p, them := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() {
		stockLibrary(p, 4)
		g.RegisterReplacementForTest(millAmountReplacement(them.ID, func(int) int { return 0 }, "none"))
	})

	if got := millN(t, g, p, 2); len(got) != 0 {
		t.Errorf("milled %v, want nothing", got)
	}
	if len(p.Graveyard.Cards) != 0 {
		t.Errorf("graveyard holds %d, want 0", len(p.Graveyard.Cards))
	}
}

// --- what does NOT open a window -------------------------------------

// "Exile the top N cards of your library" uses the same helper and is
// not a mill (CR 701.13a defines the keyword action by its
// destination), so a mill doubler must not touch it.
func TestAnExileOfTheTopCardsIsNotAMill(t *testing.T) {
	g := newActiveGame(t)
	p, them := g.Seats[0], g.Seats[1]
	fired := 0
	g.WithWriteLock(func() {
		stockLibrary(p, 8)
		g.RegisterReplacementForTest(millAmountReplacement(them.ID, func(n int) int { fired++; return n * 2 }, "twice"))
	})

	var got []uuid.UUID
	g.WithWriteLock(func() {
		var err error
		got, err = g.MillToZoneForEffect(p.ID, 2, ZoneExile)
		if err != nil {
			t.Fatalf("MillToZoneForEffect: %v", err)
		}
	})

	if fired != 0 {
		t.Errorf("the mill replacement fired %d times on an exile of the top two", fired)
	}
	if len(got) != 2 {
		t.Errorf("exiled %d cards, want 2", len(got))
	}
}

// An `until` run opens ONE amount window PER REPETITION (#1176).
//
// This used to assert the opposite — that a run names no number and so
// opens no window at all — and the engine really did behave that way.
// It was the wrong model: "mills a card, then repeats this process
// until …" gives a one-card mill instruction and repeats it, and a
// mill-amount replacement replaces each of them. The bound is still
// not a number anybody can double (TestBruvacDoublesAMillAmountAndNotHelmsBound
// in the catalog tests pins that), but the repetitions are.
//
// Three cards land here rather than two: the clause wants two, the
// first repetition is doubled to two and reaches it, and the run stops
// with both of them in the graveyard.
func TestAnUntilRunOpensOneAmountWindowPerRepetition(t *testing.T) {
	g := newActiveGame(t)
	p, them := g.Seats[0], g.Seats[1]
	fired := 0
	var sawCounts []int
	g.WithWriteLock(func() {
		stockLibrary(p, 6)
		g.RegisterReplacementForTest(millAmountReplacement(them.ID, func(n int) int {
			fired++
			sawCounts = append(sawCounts, n)
			return n * 2
		}, "twice"))
	})

	var got []uuid.UUID
	g.WithWriteLock(func() {
		err := g.MillToZoneThenForEffect(p.ID, 0, ZoneGraveyard, func(landed []Card) bool {
			return len(landed) >= 2
		}, func(_ *Game, milled []uuid.UUID) error {
			got = milled
			return nil
		})
		if err != nil {
			t.Fatalf("MillToZoneThenForEffect: %v", err)
		}
	})

	if fired != 1 {
		t.Errorf("the mill replacement fired %d times, want 1 — one repetition reached the "+
			"clause, and each repetition is its own instruction", fired)
	}
	for i, n := range sawCounts {
		if n != 1 {
			t.Errorf("repetition %d asked the window about a mill of %d, want 1 — the printed "+
				"instruction is \"mills a card\"", i, n)
		}
	}
	if len(got) != 2 {
		t.Errorf("the until-run milled %d cards, want 2 — one repetition, doubled", len(got))
	}
	if len(p.Graveyard.Cards) != 2 {
		t.Errorf("graveyard holds %d, want 2", len(p.Graveyard.Cards))
	}
}

// The same run with NOTHING watching still mills one card per
// repetition, so a clause that wants two cards gets exactly two.
func TestAnUntilRunWithNoReplacementMillsOneCardPerRepetition(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	g.WithWriteLock(func() { stockLibrary(p, 6) })

	var got []uuid.UUID
	g.WithWriteLock(func() {
		err := g.MillToZoneThenForEffect(p.ID, 0, ZoneGraveyard, func(landed []Card) bool {
			return len(landed) >= 2
		}, func(_ *Game, milled []uuid.UUID) error {
			got = milled
			return nil
		})
		if err != nil {
			t.Fatalf("MillToZoneThenForEffect: %v", err)
		}
	})

	if len(got) != 2 {
		t.Errorf("the until-run milled %d cards, want 2", len(got))
	}
	if len(p.Library.Cards) != 4 {
		t.Errorf("library holds %d, want 4 — the run takes one card per repetition", len(p.Library.Cards))
	}
}

// A run whose clause never accepts ends at the bottom of the library,
// with no error and no loss (CR 701.13b) — and it does so one
// repetition at a time rather than planning the whole library up
// front.
func TestAnUntilRunThatNeverStopsEndsAtTheBottomOfTheLibrary(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	g.WithWriteLock(func() { stockLibrary(p, 5) })

	var got []uuid.UUID
	g.WithWriteLock(func() {
		err := g.MillToZoneThenForEffect(p.ID, 0, ZoneGraveyard, func([]Card) bool { return false },
			func(_ *Game, milled []uuid.UUID) error {
				got = milled
				return nil
			})
		if err != nil {
			t.Fatalf("MillToZoneThenForEffect: %v", err)
		}
	})

	if len(got) != 5 {
		t.Errorf("milled %d cards, want the whole 5-card library", len(got))
	}
	if len(p.Library.Cards) != 0 {
		t.Errorf("library holds %d, want 0", len(p.Library.Cards))
	}
	if p.Eliminated {
		t.Error("running a library out with a mill is not a loss (CR 701.13b)")
	}
}

// --- CR 614.10, and what a cancelled mill still owes ------------------

// A cancelled mill moves nothing and the caller's continuation still
// runs, with an empty list — #808's rule for the life tail, #853's for
// the route, #762's for the tokens.
func TestACancelledMillStillRunsItsContinuation(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	g.WithWriteLock(func() {
		stockLibrary(p, 4)
		g.RegisterReplacementForTest(millCancelReplacement("no mill"))
	})

	ran := 0
	var got []uuid.UUID
	g.WithWriteLock(func() {
		if err := g.MillToZoneThenForEffect(p.ID, 2, ZoneGraveyard, nil, func(_ *Game, milled []uuid.UUID) error {
			ran++
			got = milled
			return nil
		}); err != nil {
			t.Fatalf("MillToZoneThenForEffect: %v", err)
		}
	})

	if ran != 1 {
		t.Errorf("the continuation ran %d times, want exactly 1", ran)
	}
	if len(got) != 0 {
		t.Errorf("the continuation was handed %v, want nothing", got)
	}
	if len(p.Library.Cards) != 4 {
		t.Errorf("the library holds %d cards, want all 4 — a cancelled mill moves nothing", len(p.Library.Cards))
	}
}

// --- CR 616.1, and the pause -----------------------------------------

// Two DIFFERENT declared effects in one window is an ordering question,
// because the orderings differ: ×2 then +4 mills 10 from a base of 3,
// +4 then ×2 mills 14. The affected player answers it, and the mill
// happens from the resume with nothing moved before the answer.
func TestADoublerAndAPlusFourAskForAnOrder(t *testing.T) {
	for _, tc := range []struct {
		name  string
		order []string
		want  int
	}{
		{"doubler first", []string{"twice", "plus four"}, 10},
		{"plus four first", []string{"plus four", "twice"}, 14},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			p, them := g.Seats[0], g.Seats[1]
			g.WithWriteLock(func() {
				stockLibrary(p, 20)
				g.RegisterReplacementForTest(millAmountReplacement(them.ID, func(n int) int { return n * 2 }, "twice"))
				g.RegisterReplacementForTest(millAmountReplacement(them.ID, func(n int) int { return n + 4 }, "plus four"))
			})

			got := millN(t, g, p, 3)

			if len(got) != 0 {
				t.Fatalf("the paused mill returned %v, want nothing — no card has moved yet", got)
			}
			if len(p.Graveyard.Cards) != 0 {
				t.Fatalf("%d cards reached the graveyard before the prompt was answered", len(p.Graveyard.Cards))
			}
			c := onlyPrompt(t, g)
			if c.Kind != PendingChoiceReplacementOrder {
				t.Fatalf("prompt kind = %q, want %q", c.Kind, PendingChoiceReplacementOrder)
			}
			// #982: the affected player of a mill is the player
			// MILLING, not whoever controls the replacements — which
			// here is the other seat, twice over.
			if c.Chooser != p.ID {
				t.Fatalf("chooser = %s, want the milling player %s (CR 616.1)", c.Chooser, p.ID)
			}
			if err := g.ResolveReplacementOrder(c.ID, c.Chooser, orderByLabel(t, g, c, tc.order...)); err != nil {
				t.Fatalf("ResolveReplacementOrder: %v", err)
			}
			if got := len(p.Graveyard.Cards); got != tc.want {
				t.Errorf("milled %d cards, want %d", got, tc.want)
			}
		})
	}
}

// The continuation form pauses too, and its list is the one the resume
// produces — the whole reason it is a continuation rather than a return
// value.
func TestAPausedMillHandsItsContinuationTheResumedList(t *testing.T) {
	g := newActiveGame(t)
	p, them := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() {
		stockLibrary(p, 10)
		g.RegisterReplacementForTest(millAmountReplacement(them.ID, func(n int) int { return n * 2 }, "twice"))
		g.RegisterReplacementForTest(millAmountReplacement(them.ID, func(n int) int { return n + 1 }, "plus one"))
	})

	ran := 0
	var got []uuid.UUID
	g.WithWriteLock(func() {
		if err := g.MillToZoneThenForEffect(p.ID, 2, ZoneGraveyard, nil, func(_ *Game, milled []uuid.UUID) error {
			ran++
			got = milled
			return nil
		}); err != nil {
			t.Fatalf("MillToZoneThenForEffect: %v", err)
		}
	})
	if ran != 0 {
		t.Fatalf("the continuation ran %d times with the prompt still open", ran)
	}

	c := onlyPrompt(t, g)
	if err := g.ResolveReplacementOrder(c.ID, c.Chooser, orderByLabel(t, g, c, "twice", "plus one")); err != nil {
		t.Fatalf("ResolveReplacementOrder: %v", err)
	}
	if ran != 1 {
		t.Fatalf("the continuation ran %d times after the answer, want exactly 1", ran)
	}
	if len(got) != 5 {
		t.Errorf("the continuation was handed %d cards, want 5 (2 ×2 then +1)", len(got))
	}
}

// The undo contract the amount window signs, and the reason
// cloneReplacementResume gives a snapshot its own copy of the mill
// tail: rewinding into the open ordering prompt and answering again
// mills the same cards and reports the same list, rather than finding
// a continuation the live run has already consumed through the shared
// pointer.
func TestUndoAcrossThePausedMillAmountReplays(t *testing.T) {
	g := newActiveGame(t)
	p, them := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() {
		stockLibrary(p, 12)
		g.RegisterReplacementForTest(millAmountReplacement(them.ID, func(n int) int { return n * 2 }, "twice"))
		g.RegisterReplacementForTest(millAmountReplacement(them.ID, func(n int) int { return n + 1 }, "plus one"))
	})

	ran := 0
	var got []uuid.UUID
	g.WithWriteLock(func() {
		if err := g.MillToZoneThenForEffect(p.ID, 2, ZoneGraveyard, nil, func(_ *Game, milled []uuid.UUID) error {
			ran++
			got = append([]uuid.UUID(nil), milled...)
			return nil
		}); err != nil {
			t.Fatalf("MillToZoneThenForEffect: %v", err)
		}
	})
	promptOpen := g.Clone()

	c := onlyPrompt(t, g)
	if err := g.ResolveReplacementOrder(c.ID, c.Chooser, orderByLabel(t, g, c, "twice", "plus one")); err != nil {
		t.Fatalf("ResolveReplacementOrder: %v", err)
	}
	first := append([]uuid.UUID(nil), got...)
	if len(first) != 5 || ran != 1 {
		t.Fatalf("first run milled %d cards in %d continuation runs, want 5 in 1", len(first), ran)
	}

	ran = 0
	got = nil
	g.WithWriteLock(func() { g.RestoreFrom(promptOpen) })
	if n := len(g.Seats[0].Library.Cards); n != 12 {
		t.Fatalf("the rewind left %d cards on the library, want all 12 back", n)
	}

	replay := onlyPrompt(t, g)
	if err := g.ResolveReplacementOrder(replay.ID, replay.Chooser,
		orderByLabel(t, g, replay, "twice", "plus one")); err != nil {
		t.Fatalf("ResolveReplacementOrder (replay): %v", err)
	}
	if ran != 1 {
		t.Errorf("the continuation ran %d times on the replay, want 1", ran)
	}
	if len(got) != len(first) {
		t.Errorf("replayed milled this way = %d cards, want the first run's %d", len(got), len(first))
	}
}

// A mill whose CR 616 prompt is taken AWAY still tells its caller —
// the abandon-side half of the rule above, and the arm #982's
// enumeration test requires in abandonZoneRouteLocked.
func TestADroppedMillPromptStillRunsItsContinuation(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	p, them := g.Seats[1], g.Seats[2]
	g.WithWriteLock(func() {
		stockLibrary(p, 10)
		g.RegisterReplacementForTest(millAmountReplacement(them.ID, func(n int) int { return n * 2 }, "twice"))
		g.RegisterReplacementForTest(millAmountReplacement(them.ID, func(n int) int { return n + 1 }, "plus one"))
	})

	ran := 0
	g.WithWriteLock(func() {
		if err := g.MillToZoneThenForEffect(p.ID, 2, ZoneGraveyard, nil, func(*Game, []uuid.UUID) error {
			ran++
			return nil
		}); err != nil {
			t.Fatalf("MillToZoneThenForEffect: %v", err)
		}
	})
	if c := onlyPrompt(t, g); c.Chooser != p.ID {
		t.Fatalf("chooser = %s, want the milling player %s", c.Chooser, p.ID)
	}

	if err := g.Concede(p.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}

	if len(g.PendingChoices) != 0 {
		t.Fatalf("%d prompts survived the concede", len(g.PendingChoices))
	}
	if ran != 1 {
		t.Errorf("the continuation ran %d times, want exactly 1 — an abandoned mill owes its caller an answer", ran)
	}
}

// --- the per-card window still runs behind the amount one -------------

// Both windows, in order: the amount doubles, and then each of the
// cards the settled amount asks for takes the ordinary exit, where a
// graveyard replacement can still redirect it. "Milled this way" counts
// what landed (CR 400.7, §5l), so the redirected card is not in the
// list even though the doubled count paid for it.
func TestTheAmountWindowRunsBeforeThePerCardOne(t *testing.T) {
	g := newActiveGame(t)
	p, them := g.Seats[0], g.Seats[1]
	var stolen uuid.UUID
	g.WithWriteLock(func() {
		stockLibrary(p, 6)
		stolen = p.Library.Cards[len(p.Library.Cards)-1].InstanceID
		g.RegisterReplacementForTest(millAmountReplacement(them.ID, func(n int) int { return n * 2 }, "twice"))
		g.RegisterReplacementForTest(intoGraveyardReplacement(stolen,
			"if it would be put into a graveyard, exile it instead", func(ev *ReplacementEvent) {
				ev.NewZone = ZoneExile
				ev.NewZoneOwner = uuid.Nil
			}))
	})

	got := millN(t, g, p, 2)

	if len(got) != 3 {
		t.Errorf("milled this way = %d cards, want 3 — four came off the library and one went to exile", len(got))
	}
	if !g.Exile.Contains(stolen) {
		t.Error("the redirected card is in exile")
	}
	if len(p.Graveyard.Cards) != 3 {
		t.Errorf("graveyard holds %d, want 3", len(p.Graveyard.Cards))
	}
}

// An `until` run whose every repetition is replaced away TERMINATES.
//
// The termination guard, and the one hazard the per-repetition model
// introduced (#1176): a repetition is a real instruction now, so
// CR 614.10's null replacement can cancel it, and a run that asked for
// a repetition which moves nothing would ask for it forever. The old
// model could not reach this — a run named no number, so it opened no
// amount window and nothing could cancel it.
//
// The run ends with an empty landed list, the library untouched, and
// the caller's continuation run exactly once.
func TestAnUntilRunEndsWhenEveryRepetitionIsReplacedAway(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	g.WithWriteLock(func() {
		stockLibrary(p, 6)
		g.RegisterReplacementForTest(millCancelReplacement("no mill"))
	})

	calls := 0
	var got []uuid.UUID
	g.WithWriteLock(func() {
		err := g.MillToZoneThenForEffect(p.ID, 0, ZoneGraveyard, func([]Card) bool { return false },
			func(_ *Game, milled []uuid.UUID) error {
				calls++
				got = milled
				return nil
			})
		if err != nil {
			t.Fatalf("MillToZoneThenForEffect: %v", err)
		}
	})

	if calls != 1 {
		t.Errorf("the continuation ran %d times, want exactly 1", calls)
	}
	if len(got) != 0 {
		t.Errorf("%d cards landed, want none", len(got))
	}
	if len(p.Library.Cards) != 6 {
		t.Errorf("library holds %d, want the untouched 6", len(p.Library.Cards))
	}
	if len(p.Graveyard.Cards) != 0 {
		t.Errorf("graveyard holds %d, want none", len(p.Graveyard.Cards))
	}
}

// The guard measures the library's DEPTH, not the landed list, so a run
// that lands NOTHING because every card is being diverted still walks
// the whole library — Helm of Obedience under Rest in Peace, which is
// the famous combo and must not be mistaken for a stalled run.
func TestAnUntilRunThatLandsNothingStillWalksTheLibrary(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	g.WithWriteLock(func() {
		stockLibrary(p, 5)
		// Every card on its way to the graveyard is exiled instead —
		// Rest in Peace's clause, as a test injection.
		g.RegisterReplacementForTest(ReplacementEffect{
			Watches: []EventKind{EventZoneMove},
			AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
				return ev.Kind == RepEventMove && ev.NewZone == ZoneGraveyard
			},
			Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
				ev.NewZone = ZoneExile
				return nil
			},
			Label: "exile it instead",
		})
	})

	var got []uuid.UUID
	g.WithWriteLock(func() {
		err := g.MillToZoneThenForEffect(p.ID, 0, ZoneGraveyard, func([]Card) bool { return false },
			func(_ *Game, milled []uuid.UUID) error {
				got = milled
				return nil
			})
		if err != nil {
			t.Fatalf("MillToZoneThenForEffect: %v", err)
		}
	})

	if len(got) != 0 {
		t.Errorf("%d cards landed in the graveyard, want none — every one was exiled instead", len(got))
	}
	if len(p.Library.Cards) != 0 {
		t.Errorf("library holds %d, want 0 — landing nothing is not the same as moving nothing", len(p.Library.Cards))
	}
}
