package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// mulligan_london_test.go — the London mulligan (CR 103.5, CR 103.5c,
// ADR 0168): a mulligan draws seven, a keep puts the counted mulligans'
// worth of cards on the bottom in the order the seat chooses, and the
// first mulligan of a multiplayer game is free.

// openChoiceFor returns the one pending choice of `kind` owned by seat,
// or nil.
func openChoiceFor(g *Game, seat int, kind PendingChoiceKind) *PendingChoice {
	id := g.Seats[seat].ID
	for _, c := range g.PendingChoices {
		if c.Chooser == id && c.Kind == kind {
			return c
		}
	}
	return nil
}

// bottomCards answers seat's bottoming prompts: it picks `picks` out of
// its hand (when a pick is asked) and orders them top-first as given.
func bottomCards(t *testing.T, g *Game, seat int, picks []uuid.UUID) {
	t.Helper()
	id := g.Seats[seat].ID
	if c := openChoiceFor(g, seat, PendingChoiceChooseCards); c != nil {
		if err := g.ResolveChooseCards(c.ID, id, picks); err != nil {
			t.Fatalf("seat %d choose bottom cards: %v", seat, err)
		}
	}
	if c := openChoiceFor(g, seat, PendingChoicePutInLibrary); c != nil {
		if err := g.ResolvePutInLibrary(c.ID, id, picks, nil); err != nil {
			t.Fatalf("seat %d order bottom cards: %v", seat, err)
		}
	}
}

// Two players, so no free mulligan (CR 100.1a): one mulligan, then a
// keep, owes one card. The window waits for it, and the card is on the
// bottom of the library afterwards.
func TestLondonMulliganTwoPlayersBottomsOne(t *testing.T) {
	g := startMulliganTable(t, 2, 0)
	mulligan(t, g, 0)
	if got := len(g.Seats[0].Hand.Cards); got != OpeningHandSize {
		t.Fatalf("hand after a mulligan = %d cards, want %d", got, OpeningHandSize)
	}
	if got := g.MulliganBottomCountLocked(g.Seats[0]); got != 1 {
		t.Fatalf("bottom count after one mulligan in a two-player game = %d, want 1", got)
	}
	mulligan(t, g, 1)
	// Round 2: seat 0 keeps and owes a card.
	keep(t, g, 0)
	if openChoiceFor(g, 0, PendingChoiceChooseCards) == nil {
		t.Fatal("no bottoming prompt after the keep")
	}
	// Seat 1 decides while seat 0 is still choosing, and owes one too.
	wantDecider(t, g, 1)
	keep(t, g, 1)
	bottomCards(t, g, 1, handIDs(g.Seats[1], 1))
	if !g.MulligansOpen {
		t.Fatal("the window closed while seat 0 still owed a card")
	}
	pick := handIDs(g.Seats[0], OpeningHandSize)[3]
	bottomCards(t, g, 0, []uuid.UUID{pick})
	if g.MulligansOpen {
		t.Fatal("the window is still open after the last bottom card")
	}
	if got := len(g.Seats[0].Hand.Cards); got != OpeningHandSize-1 {
		t.Fatalf("opening hand = %d cards, want %d", got, OpeningHandSize-1)
	}
	if b, _ := g.Seats[0].Library.Bottom(); b.InstanceID != pick {
		t.Fatalf("library bottom is %v, want the picked card %v", b.InstanceID, pick)
	}
}

// Four players: the first mulligan is free (CR 103.5c), the second
// costs one card.
func TestLondonMulliganFirstIsFreeInMultiplayer(t *testing.T) {
	g := startMulliganTable(t, 4, 0)
	mulligan(t, g, 0)
	if got := g.MulliganBottomCountLocked(g.Seats[0]); got != 0 {
		t.Fatalf("bottom count after the free mulligan = %d, want 0", got)
	}
	for seat := 1; seat < 4; seat++ {
		keep(t, g, seat)
	}
	// Round 2: seat 0 alone. A second mulligan is the first that counts.
	mulligan(t, g, 0)
	if got := g.MulliganBottomCountLocked(g.Seats[0]); got != 1 {
		t.Fatalf("bottom count after two mulligans in a four-player game = %d, want 1", got)
	}
	keep(t, g, 0)
	bottomCards(t, g, 0, []uuid.UUID{handIDs(g.Seats[0], OpeningHandSize)[0]})
	if g.MulligansOpen {
		t.Fatal("the window is still open")
	}
	if got := len(g.Seats[0].Hand.Cards); got != OpeningHandSize-1 {
		t.Fatalf("opening hand = %d cards, want %d", got, OpeningHandSize-1)
	}
}

// A free keep in multiplayer asks nothing and closes the window at once.
func TestLondonMulliganFreeKeepAsksNothing(t *testing.T) {
	g := startMulliganTable(t, 3, 0)
	mulligan(t, g, 0)
	keep(t, g, 1)
	keep(t, g, 2)
	keep(t, g, 0)
	if len(g.PendingChoices) != 0 {
		t.Fatalf("a free keep queued %d prompts", len(g.PendingChoices))
	}
	if g.MulligansOpen {
		t.Fatal("the window is still open")
	}
}

// Two cards go on the bottom in the order the seat gives, top-first:
// the last one named is the bottom card.
func TestLondonMulliganBottomsInTheChosenOrder(t *testing.T) {
	g := startMulliganTable(t, 2, 0)
	mulligan(t, g, 0)
	keep(t, g, 1) // round 1 ends; round 2 is seat 0 alone
	mulligan(t, g, 0)
	if got := g.MulliganBottomCountLocked(g.Seats[0]); got != 2 {
		t.Fatalf("bottom count after two mulligans = %d, want 2", got)
	}
	keep(t, g, 0)
	hand := handIDs(g.Seats[0], OpeningHandSize)
	a, b := hand[1], hand[5]
	bottomCards(t, g, 0, []uuid.UUID{a, b})
	lib := g.Seats[0].Library.Cards
	if lib[0].InstanceID != b || lib[1].InstanceID != a {
		t.Fatalf("library bottom two = %v, %v; want %v (bottom) then %v", lib[0].InstanceID, lib[1].InstanceID, b, a)
	}
}

// Inside the window the hand size a client names is ignored: the
// mulligan draws seven. Outside it the sandbox verb still draws N.
func TestLondonMulliganIgnoresHandSizeInTheWindow(t *testing.T) {
	g := startMulliganTable(t, 2, 0)
	if err := g.Mulligan(g.Seats[0].ID, 3); err != nil {
		t.Fatalf("mulligan: %v", err)
	}
	if got := len(g.Seats[0].Hand.Cards); got != OpeningHandSize {
		t.Fatalf("hand = %d cards, want %d", got, OpeningHandSize)
	}
}

// CR 103.5: no further mulligans once the opening hand would be zero
// cards. Two players, so all seven count.
func TestLondonMulliganStopsAtAZeroCardHand(t *testing.T) {
	g := startMulliganTable(t, 2, 0)
	mulligan(t, g, 0)
	keep(t, g, 1) // round 1 ends; seat 0 goes on alone
	for i := 1; i < OpeningHandSize; i++ {
		mulligan(t, g, 0)
	}
	if err := g.Mulligan(g.Seats[0].ID, OpeningHandSize); !errors.Is(err, ErrNoMoreMulligans) {
		t.Fatalf("an eighth mulligan: err = %v, want ErrNoMoreMulligans", err)
	}
	// Keeping puts the whole hand on the bottom: only an order is asked.
	keep(t, g, 0)
	if openChoiceFor(g, 0, PendingChoiceChooseCards) != nil {
		t.Fatal("a pick was asked for the whole hand")
	}
	bottomCards(t, g, 0, handIDs(g.Seats[0], OpeningHandSize))
	if got := len(g.Seats[0].Hand.Cards); got != 0 {
		t.Fatalf("opening hand = %d cards, want 0", got)
	}
	if g.MulligansOpen {
		t.Fatal("the window is still open")
	}
}

// A seat that concedes while it owes bottom cards stops gating the
// window.
func TestLondonMulliganConcedeWhileBottoming(t *testing.T) {
	g := startMulliganTable(t, 3, 0)
	mulligan(t, g, 0)
	keep(t, g, 1)
	keep(t, g, 2)
	mulligan(t, g, 0)
	keep(t, g, 0)
	if !g.MulligansOpen {
		t.Fatal("the window closed before seat 0 bottomed")
	}
	if err := g.Concede(g.Seats[0].ID); err != nil {
		t.Fatalf("concede: %v", err)
	}
	if g.MulligansOpen {
		t.Fatal("the window is still open after the bottoming seat conceded")
	}
}

// #3387: the house rule. A two-player table with FreeMulligans 1 gets
// the free first mulligan a multiplayer game has by the rules: the
// first mulligan owes nothing, the second owes one card.
func TestLondonMulliganTwoPlayersOneFreeMulligan(t *testing.T) {
	g := startMulliganTable(t, 2, 0)
	// Set directly, as a lobby patch would have before Start.
	g.Settings.FreeMulligans = 1
	if got := g.MulliganNextBottomLocked(g.Seats[0]); got != 0 {
		t.Fatalf("cost of a first mulligan with one free = %d, want 0", got)
	}
	mulligan(t, g, 0)
	if got := g.MulliganBottomCountLocked(g.Seats[0]); got != 0 {
		t.Fatalf("bottom count after a free mulligan = %d, want 0", got)
	}
	if got := g.MulliganNextBottomLocked(g.Seats[0]); got != 1 {
		t.Fatalf("cost of a second mulligan = %d, want 1", got)
	}
	keep(t, g, 1)
	// Round 2: seat 0 keeps after its free mulligan and owes nothing,
	// so no prompt holds the window.
	keep(t, g, 0)
	if openChoiceFor(g, 0, PendingChoiceChooseCards) != nil {
		t.Fatal("a bottoming prompt after a free mulligan")
	}
	if g.MulligansOpen {
		t.Fatal("the window is still open after both seats kept")
	}
	if got := len(g.Seats[0].Hand.Cards); got != OpeningHandSize {
		t.Fatalf("opening hand = %d cards, want %d", got, OpeningHandSize)
	}
}

// A count gives that many free mulligans, and the first counted one
// after them owes one card, at any table size.
func TestLondonMulliganFreeMulligansCount(t *testing.T) {
	for _, seats := range []int{2, 4} {
		g := startMulliganTable(t, seats, 0)
		g.Settings.FreeMulligans = 3
		for i := 0; i < 3; i++ {
			g.Seats[0].MulligansTaken = i
			if got := g.MulliganNextBottomLocked(g.Seats[0]); got != 0 {
				t.Fatalf("%d seats, mulligan %d of 3 free costs %d, want 0", seats, i+1, got)
			}
		}
		g.Seats[0].MulligansTaken = 3
		if got := g.MulliganBottomCountLocked(g.Seats[0]); got != 0 {
			t.Fatalf("%d seats, bottom after 3 free mulligans = %d, want 0", seats, got)
		}
		if got := g.MulliganNextBottomLocked(g.Seats[0]); got != 1 {
			t.Fatalf("%d seats, the fourth mulligan costs %d, want 1", seats, got)
		}
		// Free ones do not count toward the limit (CR 103.5c's logic).
		g.Seats[0].MulligansTaken = 3 + OpeningHandSize - 1
		if !g.MulliganAllowedLocked(g.Seats[0]) {
			t.Fatalf("%d seats: refused with %d counted mulligans", seats, OpeningHandSize-1)
		}
	}
}

// The rules floor: a multiplayer game keeps its free first mulligan
// whatever the count says, so 0 there still means one.
func TestLondonMulliganCountNeverBelowTheRules(t *testing.T) {
	g := startMulliganTable(t, 4, 0)
	g.Settings.FreeMulligans = 0
	if got := g.MulliganNextBottomLocked(g.Seats[0]); got != 0 {
		t.Fatalf("first mulligan at a four-seat table costs %d, want 0", got)
	}
}

// Unlimited: every mulligan is free, nothing is ever bottomed, and the
// zero-card limit never refuses.
func TestLondonMulliganUnlimitedFreeMulligans(t *testing.T) {
	g := startMulliganTable(t, 2, 0)
	g.Settings.FreeMulligans = FreeMulligansUnlimited
	for i := 0; i < 3*OpeningHandSize; i++ {
		g.Seats[0].MulligansTaken = i
		if got := g.MulliganNextBottomLocked(g.Seats[0]); got != 0 {
			t.Fatalf("unlimited: mulligan %d costs %d, want 0", i+1, got)
		}
		if !g.MulliganAllowedLocked(g.Seats[0]) {
			t.Fatalf("unlimited: mulligan %d refused", i+1)
		}
	}
	g.Seats[0].MulligansTaken = 0
	mulligan(t, g, 0)
	mulligan(t, g, 1)
	keep(t, g, 0)
	keep(t, g, 1)
	if g.MulligansOpen {
		t.Fatal("a keep after free mulligans held the window")
	}
}
