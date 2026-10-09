package game

import (
	"fmt"
	"math"
	"math/rand/v2"
	"os"
	"strconv"
	"testing"
)

// shuffle_fairness_test.go measures whether every card in a Commander
// library is equally likely to be drawn, through the paths a real table
// takes: a deck installed in the lobby (ReplaceDeck, which the deck
// upload, a curated bot deck and the MCP seat's set_deck all end in),
// the opening roll, the starting player's choice, the deal, a mulligan,
// and a mid-game "search your library, then shuffle". It does not test
// Zone.Shuffle alone: the question is whether anything AROUND the
// shuffle (a skipped shuffle, a reused seed, a draw from the wrong end)
// makes some cards likelier than others.
//
// The default run is small and deterministic (seeded keys), so it
// cannot flake and costs well under a second. CMDCTRL_SHUFFLE_STATS=N
// runs N games per scenario on fresh crypto-minted keys, the way
// production starts a game, and logs the full numbers:
//
//	CMDCTRL_SHUFFLE_STATS=20000 GO_DOCKER_ENV=CMDCTRL_SHUFFLE_STATS \
//	  scripts/go-docker.sh test ./internal/game -run TestShuffleFairness -v

const (
	fairSeats   = 4
	fairLibrary = 99
	fairTopN    = 15
	// fairMinP is the smallest p-value a check accepts. Each check is a
	// test of a null hypothesis that is true when the shuffle is fair,
	// so a fair run fails one with probability 1e-6.
	fairMinP = 1e-6
)

// fairDeck is a commander plus 99 distinct cards in decklist order, so
// card i is the i-th line of the list (index 0 at the bottom of the
// unshuffled library, index 98 on top).
func fairDeck(seat int) []Card {
	deck := make([]Card, 0, fairLibrary+1)
	deck = append(deck, NewCommander(fmt.Sprintf("Seat %d Commander", seat), [16]byte{}))
	for i := 0; i < fairLibrary; i++ {
		deck = append(deck, NewCard(fmt.Sprintf("Card %02d", i), [16]byte{}))
	}
	return deck
}

func fairIndex(t *testing.T, c Card) int {
	t.Helper()
	var i int
	if _, err := fmt.Sscanf(c.Name, "Card %d", &i); err != nil {
		t.Fatalf("not a fairness-deck card: %q", c.Name)
	}
	return i
}

// fairOrder is a seat's 99 cards in the order they come off the deck:
// the hand in the order drawn, then the library from the top.
func fairOrder(t *testing.T, p *Player) []int {
	t.Helper()
	out := make([]int, 0, fairLibrary)
	for _, c := range p.Hand.Cards {
		out = append(out, fairIndex(t, c))
	}
	for i := len(p.Library.Cards) - 1; i >= 0; i-- {
		out = append(out, fairIndex(t, p.Library.Cards[i]))
	}
	if len(out) != fairLibrary {
		t.Fatalf("seat has %d cards in hand + library, want %d", len(out), fairLibrary)
	}
	return out
}

// fairGame seats four players exactly as the lobby does (AddPlayer at
// join, then ReplaceDeck when the deck is uploaded), starts the table
// on the opening roll, rolls it out, and lets the winner go first.
// key == nil mints the key from crypto/rand, as production does.
func fairGame(t *testing.T, key *rand.Rand) *Game {
	t.Helper()
	g := NewGame()
	for seat := 0; seat < fairSeats; seat++ {
		p, err := g.AddPlayer(fmt.Sprintf("P%d", seat), buildTestDeck("Placeholder"))
		if err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
		if err := g.ReplaceDeck(p.ID, fairDeck(seat)); err != nil {
			t.Fatalf("ReplaceDeck: %v", err)
		}
	}
	if err := g.StartWithOpeningRoll(key); err != nil {
		t.Fatalf("StartWithOpeningRoll: %v", err)
	}
	for i := 0; g.OpeningRoll.Chooser < 0; i++ {
		if i > 64 {
			t.Fatal("the opening roll did not settle")
		}
		round := g.OpeningRoll.Rounds[len(g.OpeningRoll.Rounds)-1]
		for _, seat := range round.Seats {
			if err := g.RollOpening(g.Seats[seat].ID); err != nil {
				t.Fatalf("RollOpening: %v", err)
			}
		}
	}
	chooser := g.OpeningRoll.Chooser
	if err := g.ChooseStartingPlayer(g.Seats[chooser].ID, chooser); err != nil {
		t.Fatalf("ChooseStartingPlayer: %v", err)
	}
	return g
}

// fairTally accumulates one scenario's draws.
type fairTally struct {
	hand    int // cards counted as "in hand"
	inHand  [fairLibrary]float64
	inTop   [fairLibrary]float64
	posSum  [fairLibrary]float64
	samples int // seat-orders recorded
}

func (f *fairTally) add(order []int, hand int) {
	f.hand = hand
	f.samples++
	for pos, card := range order {
		if pos < hand {
			f.inHand[card]++
		} else if pos < hand+fairTopN {
			f.inTop[card]++
		}
		f.posSum[card] += float64(pos)
	}
}

// chiSquareP is the upper-tail p-value of a chi-square statistic,
// by the Wilson-Hilferty cube-root approximation (accurate to a few
// parts in a thousand at df ~ 100, which is far finer than a 1e-6
// threshold needs).
func chiSquareP(x float64, df int) float64 {
	k := float64(df)
	z := (math.Cbrt(x/k) - (1 - 2/(9*k))) / math.Sqrt(2/(9*k))
	return 0.5 * math.Erfc(z/math.Sqrt2)
}

// uniformChi is the goodness-of-fit of counts against "every card
// equally often".
func uniformChi(counts []float64) (stat, p, lo, hi, want float64) {
	var total float64
	for _, c := range counts {
		total += c
	}
	want = total / float64(len(counts))
	lo, hi = math.Inf(1), math.Inf(-1)
	for _, c := range counts {
		d := c - want
		stat += d * d / want
		lo = math.Min(lo, c)
		hi = math.Max(hi, c)
	}
	return stat, chiSquareP(stat, len(counts)-1), lo, hi, want
}

// checkUniform fails if the counts are not plausibly uniform and logs
// the statistic, the p-value and the extremes as rates.
func (f *fairTally) checkUniform(t *testing.T, label string, counts []float64, perSample float64) {
	t.Helper()
	stat, p, lo, hi, want := uniformChi(counts)
	n := float64(f.samples)
	t.Logf("%-48s chi2=%7.2f df=%d p=%.4f  min %.5f  max %.5f  expected %.5f (per card per deal)",
		label, stat, len(counts)-1, p, lo/n, hi/n, perSample/fairLibrary)
	if p < fairMinP {
		t.Errorf("%s: not uniform, chi2=%.2f p=%.2g (min %.0f, max %.0f, expected %.1f)", label, stat, p, lo, hi, want)
	}
}

// checkMeanPosition fails if any card's mean position is implausibly
// far from the middle of the deck. A uniform position on 0..98 has
// mean 49 and variance (99^2-1)/12.
func (f *fairTally) checkMeanPosition(t *testing.T, label string) {
	t.Helper()
	n := float64(f.samples)
	sd := math.Sqrt((fairLibrary*fairLibrary-1)/12.0) / math.Sqrt(n)
	worst, worstCard := 0.0, -1
	lo, hi := math.Inf(1), math.Inf(-1)
	for card, s := range f.posSum {
		m := s / n
		lo, hi = math.Min(lo, m), math.Max(hi, m)
		if z := math.Abs(m-49) / sd; z > worst {
			worst, worstCard = z, card
		}
	}
	// Bonferroni over 99 cards at fairMinP: |z| > ~5.7.
	limit := math.Sqrt2 * erfcInv(fairMinP/fairLibrary)
	t.Logf("%-48s mean position min %.2f max %.2f (expected 49.00, sd of a mean %.2f), worst |z| %.2f (Card %02d)",
		label, lo, hi, sd, worst, worstCard)
	if worst > limit {
		t.Errorf("%s: Card %02d's mean position is %.2f sd from the middle", label, worstCard, worst)
	}
}

// erfcInv inverts math.Erfc by bisection; it is only ever asked for a
// handful of thresholds.
func erfcInv(y float64) float64 {
	lo, hi := 0.0, 10.0
	for range 200 {
		mid := (lo + hi) / 2
		if math.Erfc(mid) > y {
			lo = mid
		} else {
			hi = mid
		}
	}
	return (lo + hi) / 2
}

func (f *fairTally) report(t *testing.T, label string) {
	t.Helper()
	f.checkUniform(t, label+": in the opening hand", f.inHand[:], float64(f.hand))
	f.checkUniform(t, label+": in the next 15", f.inTop[:], fairTopN)
	f.checkMeanPosition(t, label)
}

// handOverlap is how many card indices two hands share.
func handOverlap(a, b []int, hand int) int {
	var in [fairLibrary]bool
	for _, c := range a[:hand] {
		in[c] = true
	}
	n := 0
	for _, c := range b[:hand] {
		if in[c] {
			n++
		}
	}
	return n
}

// overlapCheck tests that two hands drawn from identically listed decks
// share as many card lines as independent hands would: hypergeometric,
// mean h*h/99. A reused seed would make them identical.
type overlapCheck struct {
	sum, n float64
	same   int
}

func (o *overlapCheck) add(a, b []int, hand int) {
	k := handOverlap(a, b, hand)
	o.sum += float64(k)
	o.n++
	if k == hand {
		o.same++
	}
}

func (o *overlapCheck) check(t *testing.T, label string, hand int) {
	t.Helper()
	h, N := float64(hand), float64(fairLibrary)
	mean := h * h / N
	variance := h * (h / N) * ((N - h) / N) * ((N - h) / (N - 1))
	got := o.sum / o.n
	z := (got - mean) / math.Sqrt(variance/o.n)
	p := math.Erfc(math.Abs(z) / math.Sqrt2)
	t.Logf("%-48s mean shared cards %.4f (expected %.4f), z=%+.2f p=%.4f, identical hands %d of %.0f",
		label, got, mean, z, p, o.same, o.n)
	if p < fairMinP {
		t.Errorf("%s: hands are correlated, mean overlap %.4f vs %.4f (z=%.2f)", label, got, mean, z)
	}
}

func fairGames(t *testing.T) (n int, fresh bool) {
	if s := os.Getenv("CMDCTRL_SHUFFLE_STATS"); s != "" {
		v, err := strconv.Atoi(s)
		if err != nil || v < 100 {
			t.Fatalf("CMDCTRL_SHUFFLE_STATS=%q: want a game count of at least 100", s)
		}
		return v, true
	}
	return 600, false
}

func TestShuffleFairness(t *testing.T) {
	games, fresh := fairGames(t)
	keyFor := func(i int) *rand.Rand {
		if fresh {
			return nil // crypto/rand, as the lobby starts a game
		}
		return rand.New(rand.NewPCG(0x5eed, uint64(i)))
	}
	t.Logf("%d games of %d seats per scenario, fresh crypto keys: %v", games, fairSeats, fresh)

	t.Run("opening hand", func(t *testing.T) {
		var all, seat0 fairTally
		var seats, nextGame overlapCheck
		var prev []int
		for i := range games {
			g := fairGame(t, keyFor(i))
			orders := make([][]int, fairSeats)
			for s, p := range g.Seats {
				orders[s] = fairOrder(t, p)
				all.add(orders[s], OpeningHandSize)
			}
			seat0.add(orders[0], OpeningHandSize)
			for a := 0; a < fairSeats; a++ {
				for b := a + 1; b < fairSeats; b++ {
					seats.add(orders[a], orders[b], OpeningHandSize)
				}
			}
			if prev != nil {
				nextGame.add(prev, orders[0], OpeningHandSize)
			}
			prev = orders[0]
		}
		all.report(t, "every seat")
		seat0.report(t, "seat 0 alone")
		seats.check(t, "seat vs seat, same game", OpeningHandSize)
		nextGame.check(t, "seat 0, game k vs game k+1", OpeningHandSize)
	})

	// The starting seat mulligans to 6 (game.Mulligan: the hand goes
	// back, the library is shuffled, six are drawn). The deal is the
	// one that was mulliganed, so the second hand is measured against
	// the first as well as against uniform.
	t.Run("mulligan to 6", func(t *testing.T) {
		var tally fairTally
		var vsFirst overlapCheck
		for i := range games {
			g := fairGame(t, keyFor(i))
			p := g.Seats[g.MulliganDecider()]
			first := fairOrder(t, p)
			if err := g.Mulligan(p.ID, 6); err != nil {
				t.Fatalf("Mulligan: %v", err)
			}
			second := fairOrder(t, p)
			tally.add(second, 6)
			vsFirst.add(first, second, 6)
		}
		tally.report(t, "after a mulligan to 6")
		vsFirst.check(t, "mulligan hand vs the hand before", 6)
	})

	// Two mulligans back to back: the free mulligan and the next one,
	// as the client sends them (hand_size 7). The two redraws use the
	// same seat's stream on consecutive counters.
	t.Run("two mulligans to 7", func(t *testing.T) {
		var tally fairTally
		var vsPrev overlapCheck
		for i := range games {
			g := fairGame(t, keyFor(i))
			p := g.Seats[g.MulliganDecider()]
			if err := g.Mulligan(p.ID, 7); err != nil {
				t.Fatalf("Mulligan: %v", err)
			}
			first := fairOrder(t, p)
			// The next round reaches this seat after everyone else
			// keeps.
			for g.MulliganDecider() != p.Seat {
				if err := g.KeepHand(g.Seats[g.MulliganDecider()].ID); err != nil {
					t.Fatalf("KeepHand: %v", err)
				}
			}
			if err := g.Mulligan(p.ID, 7); err != nil {
				t.Fatalf("second Mulligan: %v", err)
			}
			second := fairOrder(t, p)
			tally.add(second, 7)
			vsPrev.add(first, second, 7)
		}
		tally.report(t, "after a second mulligan")
		vsPrev.check(t, "second mulligan vs the first", 7)
	})

	// Mid-game: put the library back into decklist order (the worst
	// case for a shuffle that does not happen), then resolve a search
	// that finds nothing and shuffles, the shape of a fetch land that
	// whiffs. The next seven cards off the top are measured.
	t.Run("search then shuffle", func(t *testing.T) {
		var tally fairTally
		for i := range games {
			g := fairGame(t, keyFor(i))
			for g.MulligansOpen {
				if err := g.KeepHand(g.Seats[g.MulliganDecider()].ID); err != nil {
					t.Fatalf("KeepHand: %v", err)
				}
			}
			p := g.Seats[0]
			g.mu.Lock()
			// Hand back on the library, decklist order, Card 98 on top.
			cards := append(p.Library.Cards, p.Hand.Cards...)
			p.Hand.Cards = nil
			sorted := make([]Card, fairLibrary)
			for _, c := range cards {
				sorted[fairIndex(t, c)] = c
			}
			p.Library.Cards = sorted
			err := g.SearchLibraryThenForEffect(SearchLibrarySpec{
				Player:  p.ID,
				Pred:    func(Card) bool { return false },
				Dest:    ZoneHand,
				Shuffle: true,
			})
			g.mu.Unlock()
			if err != nil {
				t.Fatalf("SearchLibraryThenForEffect: %v", err)
			}
			order := make([]int, 0, fairLibrary)
			for i := len(p.Library.Cards) - 1; i >= 0; i-- {
				order = append(order, fairIndex(t, p.Library.Cards[i]))
			}
			tally.add(order, 7)
		}
		tally.report(t, "after search then shuffle")
	})
}
