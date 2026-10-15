package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// any_player_pays_each_test.go — ADR 0187 §2 (#3572): one any-player
// offer for each object, paid in life. Cleansing's shape, measured at
// the engine with no card in the way.

type eachOutcome struct {
	runs    int
	unsaved []uuid.UUID
}

// eachFixture is a four-seat game with two lands per seat, the
// effect's controller being seat 2.
type eachFixture struct {
	g     *Game
	src   uuid.UUID
	lands [4][]uuid.UUID
	out   *eachOutcome
}

func newEachFixture(t *testing.T) *eachFixture {
	t.Helper()
	return newEachFixtureOn(t, newFourPlayerActiveGame(t))
}

// leg is the open choose_cards leg, asserting it is addressed to seat.
func (f *eachFixture) leg(t *testing.T, seat int) *PendingChoice {
	t.Helper()
	c := onlyPendingChoice(t, f.g)
	if c.Kind != PendingChoiceChooseCards || c.Chooser != f.g.Seats[seat].ID {
		t.Fatalf("open prompt is %s for %v, want a choose_cards for seat %d", c.Kind, c.Chooser, seat)
	}
	return c
}

func (f *eachFixture) answer(t *testing.T, seat int, picks ...uuid.UUID) {
	t.Helper()
	c := f.leg(t, seat)
	if err := f.g.ResolveChooseCards(c.ID, c.Chooser, picks); err != nil {
		t.Fatalf("seat %d ResolveChooseCards: %v", seat, err)
	}
}

func lifeLossEvents(g *Game, seat uuid.UUID) int {
	n := 0
	for _, ev := range g.Events {
		if ev.Kind == EventChangeLife && ev.Target == seat && ev.Amount < 0 {
			n++
		}
	}
	return n
}

// The legs run in APNAP order from the ACTIVE player, whoever controls
// the source, the controller included.
func TestAnyPlayerPaysForEachAsksInAPNAPOrderFromTheActivePlayer(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	g.Turn.ActiveSeat = 1
	f := newEachFixtureOn(t, g)
	for i, seat := range []int{1, 2, 3, 0} {
		c := f.leg(t, seat)
		if o := c.AnyPlayerOffer; o == nil || !sameIDs(o.Order, seatIDs(g, 1, 2, 3, 0)) || o.Controller != g.Seats[2].ID || o.EachLife != 1 {
			t.Fatalf("leg %d offer = %+v, want APNAP order from seat 1, controller seat 2, 1 life each", i, o)
		}
		f.answer(t, seat)
	}
	if f.out.runs != 1 {
		t.Fatalf("OnUnsaved ran %d times, want once after the last leg", f.out.runs)
	}
}

// Each leg offers every land still there, including ones an earlier
// player chose, and Max is min(lands, life / EachLife).
func TestAnyPlayerPaysForEachOffersEveryLandAndCapsMaxByLife(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	g.Seats[0].Life = 3
	g.Seats[1].Life = 40
	f := newEachFixtureOn(t, g)
	first := f.leg(t, 0)
	if first.ChooseMax != 3 || first.ChooseMin != 0 || len(first.ChooseCards) != 8 {
		t.Fatalf("seat 0 leg: min %d max %d over %d lands, want 0, 3 (its life) over 8", first.ChooseMin, first.ChooseMax, len(first.ChooseCards))
	}
	f.answer(t, 0, f.lands[0][0], f.lands[1][0])
	second := f.leg(t, 1)
	if second.ChooseMax != 8 || len(second.ChooseCards) != 8 {
		t.Fatalf("seat 1 leg: max %d over %d lands, want 8 over 8", second.ChooseMax, len(second.ChooseCards))
	}
	for _, id := range []uuid.UUID{f.lands[0][0], f.lands[1][0]} {
		if !containsID(second.ChooseCards, id) {
			t.Fatalf("a land seat 0 chose is missing from seat 1's leg: %v", id)
		}
	}
}

// newEachFixtureOn is newEachFixture over a game the test has already
// shaped (life totals, the active seat).
func newEachFixtureOn(t *testing.T, g *Game) *eachFixture {
	t.Helper()
	f := &eachFixture{g: g, out: &eachOutcome{}}
	f.src = departureTestSource(g, g.Seats[2].ID, "Cleansing")
	var all []uuid.UUID
	for i := range f.lands {
		for _, name := range []string{"Forest", "Island"} {
			id := pushLand(g, g.Seats[i].ID, name, "Basic Land — "+name)
			f.lands[i] = append(f.lands[i], id)
			all = append(all, id)
		}
	}
	g.WithWriteLock(func() {
		err := g.QueueAnyPlayerPaysForEachForEffect(AnyPlayerPaysForEach{
			Source: f.src, Controller: g.Seats[2].ID,
			Objects: all, EachLife: 1, Noun: "land",
			OnUnsaved: func(_ *Game, unsaved []uuid.UUID) error {
				f.out.runs++
				f.out.unsaved = append([]uuid.UUID(nil), unsaved...)
				return nil
			},
		})
		if err != nil {
			t.Fatalf("QueueAnyPlayerPaysForEachForEffect: %v", err)
		}
	})
	return f
}

// A player at 0 life, or one whose life can't change (CR 119.8), gets
// no leg, is logged as declining, and shows in Declined on later legs.
func TestAnyPlayerPaysForEachSkipsAPlayerWhoCannotPay(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	g.Seats[0].Life = 0
	lockUntilNextTurn(g, g.Seats[1])
	f := newEachFixtureOn(t, g)
	if got := countEvents(g, EventAnyPlayerDeclined); got != 2 {
		t.Fatalf("declined events = %d, want 2 (seat 0 at 0 life, seat 1 locked)", got)
	}
	c := f.leg(t, 2)
	if !sameIDs(c.AnyPlayerOffer.Declined, seatIDs(g, 0, 1)) {
		t.Fatalf("seat 2's leg lists declined %v, want seats 0 and 1", c.AnyPlayerOffer.Declined)
	}
	if len(c.EarlierPermanentChoicesForEffect()) != 0 {
		t.Fatalf("a seat with no leg made no choice, earlier = %+v", c.EarlierPermanentChoicesForEffect())
	}
	f.answer(t, 2)
	f.answer(t, 3)
	// Seat 0 sits at 0 life, so the state check after the first answer
	// eliminates it and its two lands leave with it (CR 800.4a).
	if f.out.runs != 1 || len(f.out.unsaved) != 6 {
		t.Fatalf("OnUnsaved ran %d times on %d lands, want once on the 6 that remain", f.out.runs, len(f.out.unsaved))
	}
}

// Nothing is paid until the last leg is answered; then each seat pays
// once per land it chose, as separate life-loss events.
func TestAnyPlayerPaysForEachPaysOnlyAfterTheLastLegAsSeparatePayments(t *testing.T) {
	f := newEachFixture(t)
	g := f.g
	f.answer(t, 0, f.lands[0][0], f.lands[0][1], f.lands[1][0])
	f.answer(t, 1, f.lands[1][1])
	f.answer(t, 2)
	for i, seat := range g.Seats {
		if seat.Life != StartingLife {
			t.Fatalf("seat %d paid %d life before the last leg", i, StartingLife-seat.Life)
		}
	}
	if f.out.runs != 0 {
		t.Fatal("OnUnsaved ran before the last leg")
	}
	f.answer(t, 3, f.lands[2][0])
	if g.Seats[0].Life != StartingLife-3 || g.Seats[1].Life != StartingLife-1 ||
		g.Seats[2].Life != StartingLife || g.Seats[3].Life != StartingLife-1 {
		t.Fatalf("life after paying = %d %d %d %d, want 37 39 40 39",
			g.Seats[0].Life, g.Seats[1].Life, g.Seats[2].Life, g.Seats[3].Life)
	}
	if n := lifeLossEvents(g, g.Seats[0].ID); n != 3 {
		t.Fatalf("seat 0 had %d life-loss events for 3 lands, want 3 separate payments", n)
	}
	if paid := countEvents(g, EventAnyPlayerPaid); paid != 3 {
		t.Fatalf("paid events = %d, want 3 (one per payer)", paid)
	}
	if f.out.runs != 1 || len(f.out.unsaved) != 3 {
		t.Fatalf("OnUnsaved ran %d times on %d lands, want once on the 3 nobody paid for", f.out.runs, len(f.out.unsaved))
	}
}

// Two players who choose the same land both pay, and it is saved once.
func TestAnyPlayerPaysForEachTwoPayersForOneLandBothPay(t *testing.T) {
	f := newEachFixture(t)
	g := f.g
	shared := f.lands[3][0]
	f.answer(t, 0, shared)
	f.answer(t, 1, shared)
	f.answer(t, 2)
	f.answer(t, 3)
	if g.Seats[0].Life != StartingLife-1 || g.Seats[1].Life != StartingLife-1 {
		t.Fatalf("life = %d %d, want both payers at 39", g.Seats[0].Life, g.Seats[1].Life)
	}
	if f.out.runs != 1 || len(f.out.unsaved) != 7 || containsID(f.out.unsaved, shared) {
		t.Fatalf("OnUnsaved ran %d times on %d lands (shared among them: %v), want once on the other 7",
			f.out.runs, len(f.out.unsaved), containsID(f.out.unsaved, shared))
	}
}

// EarlierChoices lists each earlier player in APNAP order, an empty
// list included, and Declined lists those who chose nothing.
func TestAnyPlayerPaysForEachCarriesEarlierChoicesAndDeclined(t *testing.T) {
	f := newEachFixture(t)
	g := f.g
	f.answer(t, 0, f.lands[0][0])
	f.answer(t, 1)
	c := f.leg(t, 2)
	earlier := c.EarlierPermanentChoicesForEffect()
	if len(earlier) != 2 || earlier[0].Seat != g.Seats[0].ID || len(earlier[0].Cards) != 1 ||
		earlier[1].Seat != g.Seats[1].ID || len(earlier[1].Cards) != 0 {
		t.Fatalf("earlier = %+v, want seat 0 with one land then seat 1 with none", earlier)
	}
	if !sameIDs(c.AnyPlayerOffer.Declined, seatIDs(g, 1)) {
		t.Fatalf("declined = %v, want only seat 1", c.AnyPlayerOffer.Declined)
	}
}

// ChooseSuggested is the chooser's own lands that no earlier player
// chose, capped at Max.
func TestAnyPlayerPaysForEachSuggestsOwnLandsNobodyChose(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	g.Seats[1].Life = 1
	f := newEachFixtureOn(t, g)
	first := f.leg(t, 0)
	if !sameIDs(first.ChooseSuggested, f.lands[0]) {
		t.Fatalf("seat 0 suggested %v, want its own two lands", first.ChooseSuggested)
	}
	// Seat 0 pays for one of seat 1's lands, not for its own Forest.
	f.answer(t, 0, f.lands[1][0])
	second := f.leg(t, 1)
	// Seat 1 can pay for one land: its Forest is already chosen, so the
	// suggestion is its Island, and the cap would cut it to one anyway.
	if !sameIDs(second.ChooseSuggested, f.lands[1][1:]) {
		t.Fatalf("seat 1 suggested %v, want only its unchosen Island", second.ChooseSuggested)
	}
}

func TestAnyPlayerPaysForEachSuggestionIsCutToMax(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	g.Seats[0].Life = 1
	f := newEachFixtureOn(t, g)
	first := f.leg(t, 0)
	if first.ChooseMax != 1 || len(first.ChooseSuggested) != 1 || first.ChooseSuggested[0] != f.lands[0][0] {
		t.Fatalf("max %d suggested %v, want 1 and seat 0's first land", first.ChooseMax, first.ChooseSuggested)
	}
}

// OnUnsaved gets only lands that are still the same object: a land that
// left and came back during the picks is a new object and is not
// destroyed with the rest.
func TestAnyPlayerPaysForEachUnsavedExcludesAReplacedObject(t *testing.T) {
	f := newEachFixture(t)
	g := f.g
	f.answer(t, 0)
	movePermanentAwayAndBack(t, g, g.Seats[1].ID, f.lands[1][0])
	f.answer(t, 1)
	f.answer(t, 2)
	f.answer(t, 3)
	if f.out.runs != 1 || len(f.out.unsaved) != 7 || containsID(f.out.unsaved, f.lands[1][0]) {
		t.Fatalf("unsaved = %d lands (replaced land among them: %v), want 7 without it",
			len(f.out.unsaved), containsID(f.out.unsaved, f.lands[1][0]))
	}
}

// A chooser who leaves mid-prompt is not replaced: they pay nothing and
// the next player is asked.
func TestAnyPlayerPaysForEachDepartedChooserIsNotReplaced(t *testing.T) {
	f := newEachFixture(t)
	g := f.g
	f.leg(t, 0)
	if err := g.Concede(g.Seats[0].ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	c := f.leg(t, 1)
	if len(c.ChooseCards) != 6 {
		t.Fatalf("seat 1's leg offers %d lands, want 6 (seat 0's two left with it)", len(c.ChooseCards))
	}
	for _, id := range f.lands[0] {
		if containsID(c.ChooseCards, id) {
			t.Fatal("a departed player's land is still offered")
		}
	}
	f.answer(t, 1, f.lands[1][0])
	f.answer(t, 2)
	f.answer(t, 3)
	if g.Seats[0].Life != StartingLife || g.Seats[1].Life != StartingLife-1 {
		t.Fatalf("life = %d %d, want the departed seat unchanged and seat 1 at 39", g.Seats[0].Life, g.Seats[1].Life)
	}
	if f.out.runs != 1 || len(f.out.unsaved) != 5 {
		t.Fatalf("OnUnsaved ran %d times on %d lands, want once on 5 (6 left, 1 paid for)", f.out.runs, len(f.out.unsaved))
	}
}

// A land whose controller leaves during the picks is gone from later
// legs and from the payments: nobody pays for a land that no longer
// exists.
func TestAnyPlayerPaysForEachLandOfADepartedPlayerIsNotPaidFor(t *testing.T) {
	f := newEachFixture(t)
	g := f.g
	f.answer(t, 0, f.lands[3][0], f.lands[0][0])
	if err := g.Concede(g.Seats[3].ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	c := f.leg(t, 1)
	for _, id := range f.lands[3] {
		if containsID(c.ChooseCards, id) {
			t.Fatal("a departed player's land is still offered to a later leg")
		}
	}
	f.answer(t, 1)
	f.answer(t, 2)
	if g.Seats[0].Life != StartingLife-1 {
		t.Fatalf("seat 0 life = %d, want 39: it pays for the land that still exists only", g.Seats[0].Life)
	}
}

// The leg blocks the table, a standing answer never answers it, and
// the table is not a restore point while it is open.
func TestAnyPlayerPaysForEachLegBlocksTheTableAndTakesNoStandingAnswer(t *testing.T) {
	f := newEachFixture(t)
	g := f.g
	c := f.leg(t, 0)
	if err := g.PassPriority(); !errors.Is(err, ErrChoicePending) {
		t.Fatalf("pass_priority with a leg open = %v, want ErrChoicePending", err)
	}
	if c.AutoAnswerKey != "" {
		t.Fatalf("the leg carries a standing-answer key %q", c.AutoAnswerKey)
	}
	if snap := g.CaptureSnapshot(); snap.Restorable() || snap.Continuations.ChoiceResumeFrames != 1 {
		t.Fatalf("a table with a leg open is restorable=%v with %d resume frames, want unrestorable with 1",
			snap.Restorable(), snap.Continuations.ChoiceResumeFrames)
	}
}

// With nobody to ask, OnUnsaved runs at once on every land; with no
// lands it still runs, once, on nothing.
func TestAnyPlayerPaysForEachRunsOnUnsavedAtOnceWhenNobodyCanPay(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	for _, s := range g.Seats {
		s.Life = 0
	}
	f := newEachFixtureOn(t, g)
	if len(g.PendingChoices) != 0 {
		t.Fatalf("%d prompts open with nobody able to pay", len(g.PendingChoices))
	}
	if f.out.runs != 1 || len(f.out.unsaved) != 8 {
		t.Fatalf("OnUnsaved ran %d times on %d lands, want once on all 8", f.out.runs, len(f.out.unsaved))
	}

	empty := newFourPlayerActiveGame(t)
	runs := 0
	empty.WithWriteLock(func() {
		if err := empty.QueueAnyPlayerPaysForEachForEffect(AnyPlayerPaysForEach{
			Source: uuid.New(), Controller: empty.Seats[0].ID, EachLife: 1, Noun: "land",
			OnUnsaved: func(_ *Game, unsaved []uuid.UUID) error { runs++; return nil },
		}); err != nil {
			t.Fatal(err)
		}
	})
	if runs != 1 || len(empty.PendingChoices) != 0 {
		t.Fatalf("empty set: OnUnsaved ran %d times with %d prompts open, want once and none", runs, len(empty.PendingChoices))
	}
}

// A non-positive EachLife is a programming error: reported once, and the
// card still does something.
func TestAnyPlayerPaysForEachNonPositiveLifeReportsAndRunsOnUnsaved(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	land := pushLand(g, g.Seats[0].ID, "Forest", "Basic Land — Forest")
	runs := 0
	g.WithWriteLock(func() {
		if err := g.QueueAnyPlayerPaysForEachForEffect(AnyPlayerPaysForEach{
			Source: uuid.New(), Controller: g.Seats[0].ID, Objects: []uuid.UUID{land}, Noun: "land",
			OnUnsaved: func(_ *Game, unsaved []uuid.UUID) error { runs++; return nil },
		}); err != nil {
			t.Fatal(err)
		}
	})
	if runs != 1 || countEvents(g, EventEffectError) != 1 || len(g.PendingChoices) != 0 {
		t.Fatalf("runs %d, errors %d, prompts %d; want 1, 1, 0", runs, countEvents(g, EventEffectError), len(g.PendingChoices))
	}
}
