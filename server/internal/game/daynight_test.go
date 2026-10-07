package game

import (
	"testing"

	"github.com/google/uuid"
)

// daynight_test.go — day and night (CR 731) and daybound / nightbound
// (CR 702.145), ADR 0132, #2561.
//
// The shapes worth failing loudly: the designation changes only on the
// rules' own terms (an untap-step check that reads the PREVIOUS turn,
// never fires from neither, and is not a trigger source for the first
// designation); a werewolf follows the designation without ever being
// a blink (no zone move, no ETB, and an entry at night is a face, not a
// transform); and nothing else can turn one over.

// werewolfFixture is a daybound // nightbound double-faced creature,
// the shape Scryfall gives every Midnight Hunt werewolf: a human front
// with daybound, a wolf back with nightbound, each face carrying its own
// keyword.
func werewolfFixture(owner uuid.UUID) Card {
	c := Card{
		InstanceID: uuid.New(),
		OracleID:   "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
		Owner:      owner,
		Controller: owner,
		Layout:     LayoutTransform,
		Faces: []Face{
			{
				Name: "Fixture Human", TypeLine: "Creature — Human Werewolf",
				ManaCost: "{2}{G}", Colors: []string{"G"}, Power: 2, Toughness: 3,
				Keywords: []string{KeywordDaybound},
			},
			{
				Name: "Fixture Wolf", TypeLine: "Creature — Werewolf",
				Colors: []string{"G"}, Power: 4, Toughness: 5,
				Keywords: []string{KeywordNightbound},
			},
		},
	}
	c.SetFace(0)
	return c
}

// enterFromHand puts the card onto the battlefield through the entry
// pipeline, so the zone-move event the day/night listener reads is
// really emitted.
func enterFromHand(t *testing.T, g *Game, c Card) uuid.UUID {
	t.Helper()
	owner := g.playerByIDLocked(c.Owner)
	owner.Hand.PushTop(c)
	var id uuid.UUID
	var err error
	g.WithWriteLock(func() {
		id, err = g.PutFromHandOntoBattlefieldForEffect(c.InstanceID, HandEntryOptions{})
	})
	if err != nil || id == uuid.Nil {
		t.Fatalf("put %q onto the battlefield: id=%v err=%v", c.Name, id, err)
	}
	return id
}

func setDayNight(g *Game, d DayNightDesignation) {
	g.WithWriteLock(func() { g.DayNight.Designation = d })
}

func dnCount(g *Game, kind EventKind) int {
	n := 0
	for _, ev := range g.Events {
		if ev.Kind == kind {
			n++
		}
	}
	return n
}

func TestGameStartsNeitherDayNorNight(t *testing.T) {
	g := newActiveGame(t)
	if g.IsDay() || g.IsNight() || g.DayNightDesignation() != DesignationNeither {
		t.Fatalf("a fresh game is %q, want neither (CR 731.1)", g.DayNightDesignation())
	}
}

// TestFirstDesignationIsNotAFlip: neither -> day is a gain (Amount 0);
// day -> night is "day becomes night" (Amount 1); a repeat is nothing.
func TestFirstDesignationIsNotAFlip(t *testing.T) {
	g := newActiveGame(t)
	g.WithWriteLock(func() { g.BecomeDayForEffect() })
	if !g.IsDay() {
		t.Fatalf("designation = %q, want day", g.DayNightDesignation())
	}
	evs := dnEvents(g, EventDayNightChanged)
	if len(evs) != 1 || evs[0].Label != "day" || DayNightFlipped(evs[0]) {
		t.Fatalf("first designation events = %+v, want one non-flip 'day'", evs)
	}

	g.WithWriteLock(func() { g.BecomeDayForEffect() })
	if got := dnCount(g, EventDayNightChanged); got != 1 {
		t.Errorf("a second 'it becomes day' emitted again (%d events), want nothing (CR 731.1)", got)
	}

	g.WithWriteLock(func() { g.BecomeNightForEffect() })
	evs = dnEvents(g, EventDayNightChanged)
	if len(evs) != 2 || !DayNightFlipped(evs[1]) || evs[1].Label != "night" {
		t.Fatalf("day -> night events = %+v, want a flip to night", evs)
	}
}

func dnEvents(g *Game, kind EventKind) []Event {
	var out []Event
	for _, ev := range g.Events {
		if ev.Kind == kind {
			out = append(out, ev)
		}
	}
	return out
}

func TestBecomeDayIfNeitherLeavesNightAlone(t *testing.T) {
	g := newActiveGame(t)
	setDayNight(g, DesignationNight)
	g.WithWriteLock(func() { g.BecomeDayIfNeitherForEffect() })
	if !g.IsNight() {
		t.Fatalf("'if it's neither day nor night, it becomes day' turned night into %q", g.DayNightDesignation())
	}
}

func TestToggleIsNightToDayElseNight(t *testing.T) {
	for _, tc := range []struct{ from, want DayNightDesignation }{
		{DesignationNight, DesignationDay},
		{DesignationDay, DesignationNight},
		{DesignationNeither, DesignationNight},
	} {
		g := newActiveGame(t)
		setDayNight(g, tc.from)
		g.WithWriteLock(func() { g.ToggleDayNightForEffect() })
		if got := g.DayNightDesignation(); got != tc.want {
			t.Errorf("toggle from %q = %q, want %q", tc.from, got, tc.want)
		}
	}
}

// TestUntapStepCheckReadsThePreviousTurn drives the real turn rotation:
// the spells seat 0 casts in its own turn decide what the designation
// is when seat 1's untap step runs (CR 502.2).
func TestUntapStepCheckReadsThePreviousTurn(t *testing.T) {
	for _, tc := range []struct {
		name  string
		from  DayNightDesignation
		spent int
		want  DayNightDesignation
	}{
		{"day and no spells becomes night", DesignationDay, 0, DesignationNight},
		{"day and one spell stays day", DesignationDay, 1, DesignationDay},
		{"day and many spells stays day", DesignationDay, 3, DesignationDay},
		{"night and two spells becomes day", DesignationNight, 2, DesignationDay},
		{"night and one spell stays night", DesignationNight, 1, DesignationNight},
		{"night and no spells stays night", DesignationNight, 0, DesignationNight},
		{"neither stays neither with no spells", DesignationNeither, 0, DesignationNeither},
		{"neither stays neither with two spells", DesignationNeither, 2, DesignationNeither},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			setDayNight(g, tc.from)
			me := g.Seats[g.Turn.ActiveSeat]
			g.WithWriteLock(func() {
				if tc.spent > 0 {
					g.SpellsCastThisTurn = map[uuid.UUID]CastTally{me.ID: {Total: tc.spent}}
				}
			})
			advanceOneTurn(t, g)
			if got := g.DayNightDesignation(); got != tc.want {
				t.Errorf("after the turn: %q, want %q", got, tc.want)
			}
		})
	}
}

// TestUntapStepCheckCountsOnlyTheActivePlayersSpells: a spell cast by a
// player whose turn it wasn't does not count (CR 502.2 names "the
// previous turn's active player").
func TestUntapStepCheckCountsOnlyTheActivePlayersSpells(t *testing.T) {
	g := newActiveGame(t)
	setDayNight(g, DesignationDay)
	other := g.Seats[1-g.Turn.ActiveSeat]
	g.WithWriteLock(func() {
		g.SpellsCastThisTurn = map[uuid.UUID]CastTally{other.ID: {Total: 5}}
	})
	advanceOneTurn(t, g)
	if !g.IsNight() {
		t.Errorf("an opponent's five spells kept it day; got %q, want night", g.DayNightDesignation())
	}
}

// TestFirstTurnHasNoPreviousTurn: until a turn has ended there is
// nothing for the check to read, so it cannot fire.
func TestFirstTurnHasNoPreviousTurn(t *testing.T) {
	g := newActiveGame(t)
	setDayNight(g, DesignationDay)
	g.WithWriteLock(func() { g.dayNightTurnCheckLocked() })
	if !g.IsDay() {
		t.Fatalf("the check ran with no previous turn and made it %q", g.DayNightDesignation())
	}
}

func TestDayNightSurvivesCloneAndRestore(t *testing.T) {
	g := newActiveGame(t)
	g.WithWriteLock(func() {
		g.DayNight = DayNightState{Designation: DesignationNight, PrevTurnSpells: 2, PrevTurnKnown: true}
	})
	snap := g.Clone()
	if snap.DayNight != g.DayNight {
		t.Fatalf("Clone dropped the state: %+v", snap.DayNight)
	}
	g.WithWriteLock(func() { g.DayNight = DayNightState{} })
	g.RestoreFrom(snap)
	if g.DayNight.Designation != DesignationNight || g.DayNight.PrevTurnSpells != 2 || !g.DayNight.PrevTurnKnown {
		t.Fatalf("RestoreFrom lost the state: %+v", g.DayNight)
	}
}

// --- daybound / nightbound -------------------------------------------

// TestDayboundEnteringAtNeitherMakesItDay is CR 702.145d.
func TestDayboundEnteringAtNeitherMakesItDay(t *testing.T) {
	g := newActiveGame(t)
	id := enterFromHand(t, g, werewolfFixture(g.Seats[0].ID))
	if !g.IsDay() {
		t.Fatalf("a daybound permanent entered and it is %q, want day", g.DayNightDesignation())
	}
	if c := findBattlefieldCard(g, id); c.ActiveFace != 0 {
		t.Errorf("it is day, so the werewolf should stay front face up; ActiveFace = %d", c.ActiveFace)
	}
}

// TestDayboundEnteringAtNightEntersTransformed is CR 702.145b's first
// ability, and the assertion that matters is what is NOT there: no
// EventTransform (nothing transformed), and the ETB event is harvested
// off the face that entered.
func TestDayboundEnteringAtNightEntersTransformed(t *testing.T) {
	g := newActiveGame(t)
	setDayNight(g, DesignationNight)
	before := dnCount(g, EventTransform)
	id := enterFromHand(t, g, werewolfFixture(g.Seats[0].ID))
	c := findBattlefieldCard(g, id)
	if c.ActiveFace != 1 || c.Name != "Fixture Wolf" {
		t.Fatalf("entered face = %d (%s), want the back face", c.ActiveFace, c.Name)
	}
	if got := dnCount(g, EventTransform) - before; got != 0 {
		t.Errorf("a permanent that enters transformed emitted %d EventTransform; it never transformed (CR 712.18)", got)
	}
	g.RecomputeLayersIfStaleLocked()
	if eff := findBattlefieldCard(g, id).Effective(); eff.Power != 4 || eff.Toughness != 5 {
		t.Errorf("effective P/T = %d/%d, want the wolf's 4/5", eff.Power, eff.Toughness)
	}
	AssertFaceInvariant(t, g)
}

// TestWerewolvesFollowTheDesignation: night turns the front over, day
// turns the back over, and each is a real transform (an in-place one,
// CR 712.18: same object).
func TestWerewolvesFollowTheDesignation(t *testing.T) {
	g := newActiveGame(t)
	id := enterFromHand(t, g, werewolfFixture(g.Seats[0].ID))
	epoch := findBattlefieldCard(g, id).ObjectEpoch

	g.WithWriteLock(func() { g.BecomeNightForEffect() })
	c := findBattlefieldCard(g, id)
	if c.ActiveFace != 1 {
		t.Fatalf("night did not turn the daybound permanent over (ActiveFace %d)", c.ActiveFace)
	}
	if c.ObjectEpoch != epoch {
		t.Errorf("the werewolf became a new object: epoch %d -> %d", epoch, c.ObjectEpoch)
	}
	if got := dnCount(g, EventTransform); got != 1 {
		t.Errorf("EventTransform count = %d, want 1", got)
	}

	g.WithWriteLock(func() { g.BecomeDayForEffect() })
	if c := findBattlefieldCard(g, id); c.ActiveFace != 0 {
		t.Fatalf("day did not turn the nightbound permanent back (ActiveFace %d)", c.ActiveFace)
	}
	AssertFaceInvariant(t, g)
}

// TestOnlyTheDayNightRulesTransformAWerewolf is CR 702.145b / 702.145e's
// "can't transform except due to its daybound / nightbound ability".
func TestOnlyTheDayNightRulesTransformAWerewolf(t *testing.T) {
	g := newActiveGame(t)
	id := enterFromHand(t, g, werewolfFixture(g.Seats[0].ID)) // day, front up
	before := dnCount(g, EventTransform)
	var err error
	g.WithWriteLock(func() { err = g.TransformPermanentForEffect(id) })
	if err != nil {
		t.Fatalf("TransformPermanentForEffect: %v (an instruction that does nothing is not an error)", err)
	}
	if c := findBattlefieldCard(g, id); c.ActiveFace != 0 {
		t.Fatalf("a daybound permanent transformed on an instruction (ActiveFace %d)", c.ActiveFace)
	}
	g.WithWriteLock(func() { g.BecomeNightForEffect() }) // back face, nightbound
	g.WithWriteLock(func() { err = g.TransformPermanentForEffect(id) })
	if err != nil {
		t.Fatal(err)
	}
	if c := findBattlefieldCard(g, id); c.ActiveFace != 1 {
		t.Fatalf("a nightbound permanent transformed on an instruction (ActiveFace %d)", c.ActiveFace)
	}
	if got := dnCount(g, EventTransform) - before; got != 1 {
		t.Errorf("transform events = %d, want only the day/night one", got)
	}
}

// TestUntapStepTurnsTheWerewolvesOver is the two halves together on the
// real turn rotation: a quiet turn makes it night and the werewolf turns
// over before anything untaps.
func TestUntapStepTurnsTheWerewolvesOver(t *testing.T) {
	g := newActiveGame(t)
	id := enterFromHand(t, g, werewolfFixture(g.Seats[0].ID))
	advanceOneTurn(t, g) // seat 0 cast nothing
	if !g.IsNight() {
		t.Fatalf("a turn with no spells left it %q, want night", g.DayNightDesignation())
	}
	if c := findBattlefieldCard(g, id); c.ActiveFace != 1 {
		t.Fatalf("the werewolf did not turn over with night (ActiveFace %d)", c.ActiveFace)
	}
}

// TestADayboundPermanentThatIsNotADoubleFacedCardDoesNotTransform: a
// token copy of a werewolf is not represented by a double-faced card
// (CR 712.9, 702.145b), so night leaves it alone; it still starts the
// clock (CR 702.145d reads "a permanent with daybound").
func TestADayboundPermanentThatIsNotADoubleFacedCardDoesNotTransform(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	c := Card{
		InstanceID: uuid.New(), Name: "Token Werewolf", TypeLine: "Creature — Werewolf",
		Owner: me.ID, Controller: me.ID, Power: 2, Toughness: 2,
		Keywords: []string{KeywordDaybound},
	}
	id := enterFromHand(t, g, c)
	if !g.IsDay() {
		t.Fatalf("a daybound permanent entered and it is %q, want day", g.DayNightDesignation())
	}
	g.WithWriteLock(func() { g.BecomeNightForEffect() })
	if got := dnCount(g, EventTransform); got != 0 {
		t.Errorf("a permanent with no second face emitted %d transform events", got)
	}
	if findBattlefieldCard(g, id) == nil {
		t.Fatal("the permanent vanished")
	}
}

// TestNightboundAloneMakesItNight is CR 702.145g.
func TestNightboundAloneMakesItNight(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	c := Card{
		InstanceID: uuid.New(), Name: "Lone Wolf", TypeLine: "Creature — Werewolf",
		Owner: me.ID, Controller: me.ID, Power: 3, Toughness: 3,
		Keywords: []string{KeywordNightbound},
	}
	enterFromHand(t, g, c)
	if !g.IsNight() {
		t.Fatalf("a nightbound permanent with no daybound one entered; it is %q, want night", g.DayNightDesignation())
	}
}
