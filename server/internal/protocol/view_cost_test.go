package protocol

import (
	"fmt"
	"math/rand/v2"
	"sync/atomic"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// view_cost_test.go pins #1261: building a view is LINEAR in the size
// of the board.
//
// Two stamps broke that. stampManaConditions asked
// AnyActivationRestrictionsForEffect — a walk of the battlefield —
// once per permanent, and viewOfActivatedAbilities once per card with
// an ability (#1210, landed the day before the nightly bot soak went
// red); and stampNoUntap gathered the board's untap-step restrictions
// — another walk — once per permanent. Each is correct and
// each is invisible to a correctness test: a view of a thirty-land
// board is the same view either way, only nine hundred hook calls
// dearer. So these tests count the catalog hook each walk goes
// through, at two board sizes, and fail when doubling the board more
// than doubles the calls.
//
// They count calls rather than timing anything, so they are
// deterministic on any machine.

// scanCounter wraps a board-walk catalog hook and counts its calls.
// ViewOfGame runs on one goroutine, but the counter is atomic anyway
// so a future parallel stamp cannot make the test flaky.
type scanCounter struct{ n atomic.Int64 }

func countActivationScans(t *testing.T) *scanCounter {
	t.Helper()
	c := &scanCounter{}
	prev := game.CatalogActivationRestrictions
	game.CatalogActivationRestrictions = func(id string) []game.ActivationRestriction {
		c.n.Add(1)
		if prev != nil {
			return prev(id)
		}
		return nil
	}
	t.Cleanup(func() { game.CatalogActivationRestrictions = prev })
	return c
}

func countUntapScans(t *testing.T) *scanCounter {
	t.Helper()
	c := &scanCounter{}
	prev := game.CatalogUntapStepRestrictions
	game.CatalogUntapStepRestrictions = func(key string) []game.UntapStepRestriction {
		c.n.Add(1)
		if prev != nil {
			return prev(key)
		}
		return nil
	}
	t.Cleanup(func() { game.CatalogUntapStepRestrictions = prev })
	return c
}

// boardOf returns an active game with n catalogued permanents on the
// battlefield, each carrying a mana ability and an activated ability
// — every per-card stamp this file is about runs for each of them.
// The oracle IDs are what make each permanent visible to the catalog
// hooks; nothing is registered under them.
func boardOf(t *testing.T, n int) *game.Game {
	t.Helper()
	g := buildActiveGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() {
		for i := 0; i < n; i++ {
			g.Battlefield.PushTop(game.Card{
				InstanceID: uuid.New(),
				Name:       fmt.Sprintf("Probe %d", i),
				TypeLine:   "Artifact",
				OracleID:   fmt.Sprintf("view-cost-probe-%d", i),
				Owner:      me.ID,
				Controller: me.ID,
				ManaAbilities: []game.ManaAbilityShape{
					{TapCost: true, Produced: "{C}", Label: "Add {C}"},
				},
				ActivatedAbilities: []game.ActivatedAbilityShape{
					{Label: "{T}: Draw a card.", Cost: game.AbilityCost{Tap: true}},
				},
			})
		}
	})
	return g
}

// scansFor builds one full view of an n-permanent board and returns
// how many times the hook behind `counter` was called while doing so.
func scansFor(t *testing.T, n int, counter func(*testing.T) *scanCounter) int64 {
	t.Helper()
	g := boardOf(t, n)
	c := counter(t)
	_ = ViewOfGame(g)
	return c.n.Load()
}

// assertLinear fails when doubling the board more than ~doubles the
// hook calls. A per-permanent walk of the board makes the count
// quadratic — ratio ~4 — and a walk taken once per view keeps it
// linear — ratio ~2. The threshold sits between.
func assertLinear(t *testing.T, what string, small, large int64) {
	t.Helper()
	t.Logf("%s: %d hook calls at 20 permanents, %d at 40", what, small, large)
	if small == 0 {
		t.Fatalf("%s: the hook was never called — the probe board does not reach the stamp", what)
	}
	if ratio := float64(large) / float64(small); ratio > 3 {
		t.Errorf("%s: doubling the board multiplied the hook calls by %.1f; want ~2 (one walk per view, not one per permanent)", what, ratio)
	}
}

func TestViewWalksTheBoardForActivationRestrictionsOncePerView(t *testing.T) {
	assertLinear(t, "activation restrictions",
		scansFor(t, 20, countActivationScans), scansFor(t, 40, countActivationScans))
}

func TestViewWalksTheBoardForUntapRestrictionsOncePerView(t *testing.T) {
	assertLinear(t, "untap-step restrictions",
		scansFor(t, 20, countUntapScans), scansFor(t, 40, countUntapScans))
}

// TestSeatIndexerDoesNotAllocate pins the constant-factor half of
// #1261: publicLogOf asked the seat indexer for every event of the game
// on every frame (since #1401, for every event it folds), and it used
// to format the actor's UUID as a string
// to compare it — one allocation per event per view, ~8% of a bot
// table's CPU. Comparing UUIDs allocates nothing.
func TestSeatIndexerDoesNotAllocate(t *testing.T) {
	g := buildActiveGame(t)
	v := ViewOfGame(g)
	seatOf := seatIndexer(&v)
	actor := g.Seats[1].ID
	if got := seatOf(actor); got != 1 {
		t.Fatalf("seatOf(seat 1) = %d, want 1", got)
	}
	if got := seatOf(uuid.New()); got != NoSeat {
		t.Fatalf("seatOf(stranger) = %d, want NoSeat", got)
	}
	if allocs := testing.AllocsPerRun(100, func() { _ = seatOf(actor) }); allocs != 0 {
		t.Errorf("seatOf allocated %.0f times per call; want 0", allocs)
	}
}

// #1479: the view asks for every card it projects by instance ID —
// castSourceOf, liveCardForView, liveCardForAbilityRows,
// stampActivatedAbilities — and each of those lookups used to walk
// every zone until it found the card. That made a view quadratic in
// the board, and about 7% of a bot table's CPU. The card index
// (game/card_index.go, ADR 0094) answers each one from a table,
// checked against the live zone.
//
// lookupWorkFor counts the cards the lookups of ONE view examine, on a
// board with n permanents on the battlefield and 2n cards in the last
// seat's graveyard — the last zone the old walk reached. The count is
// taken on the second view: the first builds the index, the way the
// first view of a new game does, and every view after it is the one a
// table pays for over and over.
func lookupWorkFor(t *testing.T, n int) (first, steady game.CardLookupWork) {
	t.Helper()
	g := boardOf(t, n)
	g.WithWriteLock(func() {
		last := g.Seats[len(g.Seats)-1]
		for i := 0; i < 2*n; i++ {
			last.Graveyard.PushTop(game.Card{
				InstanceID: uuid.New(),
				Name:       fmt.Sprintf("Buried %d", i),
				TypeLine:   "Creature — Bear",
				// A catalog key is what makes the view look a
				// graveyard card up (liveCardForAbilityRows); nothing
				// is registered under it.
				OracleID:   fmt.Sprintf("view-cost-buried-%d", i),
				Owner:      last.ID,
				Controller: last.ID,
			})
		}
	})
	read, stop := game.CountCardLookupWork()
	defer stop()
	_ = ViewOfGame(g)
	first = read()
	_ = ViewOfGame(g)
	total := read()
	return first, game.CardLookupWork{Lookups: total.Lookups - first.Lookups, Examined: total.Examined - first.Examined}
}

func TestViewCardLookupsAreLinearInTheBoard(t *testing.T) {
	// Quadrupling rather than doubling the board: the view also looks
	// up the seats' own few cards, a linear term big enough to blur a
	// quadratic one at 2x.
	firstSmall, small := lookupWorkFor(t, 10)
	firstLarge, large := lookupWorkFor(t, 40)
	t.Logf("n=10: first view %d lookups examined %d cards, steady view %d examined %d", firstSmall.Lookups, firstSmall.Examined, small.Lookups, small.Examined)
	t.Logf("n=40: first view %d lookups examined %d cards, steady view %d examined %d", firstLarge.Lookups, firstLarge.Examined, large.Lookups, large.Examined)
	if small.Lookups == 0 || large.Lookups <= small.Lookups {
		t.Fatalf("the view looked up %d cards at n=10 and %d at n=40 — the probe board does not reach the lookups", small.Lookups, large.Lookups)
	}
	// The shape: what ONE lookup costs must not grow with the board. A
	// walk of every zone costs more the more cards there are to walk
	// past; a table read costs one card whatever the size.
	perSmall := float64(small.Examined) / float64(small.Lookups)
	perLarge := float64(large.Examined) / float64(large.Lookups)
	if perLarge > 1.5*perSmall {
		t.Errorf("a lookup examined %.1f cards on the small board and %.1f on the 4x board; want the same (a table read), not a walk that grows with the board", perSmall, perLarge)
	}
	// The size: on the fast path a lookup examines exactly one card.
	if perLarge > 1.5 {
		t.Errorf("a steady view's lookups examined %.1f cards each; want ~1 (the fast path)", perLarge)
	}
	// And the whole view's lookup work stays linear in the board: ~4 at
	// 4x, where one walk per card would be ~16.
	if ratio := float64(large.Examined) / float64(small.Examined); ratio > 8 {
		t.Errorf("quadrupling the board multiplied the cards the view's lookups examine by %.1f; want ~4", ratio)
	}
}

// #1498: a card off the battlefield has no layer pass, so Effective()
// used to rebuild its printed characteristic — a type-line parse, a
// colour slice, the keyword merge — on every call, and the view calls
// it several times per card (the card view, CurrentPower,
// PowerForComparison, CurrentToughness), plus a parse of its own in
// effectiveTypeLine and one more for every land's intrinsic mana
// ability. ParseTypeLine alone was 5% of a bot table's CPU. The
// printed-characteristic cache (game/printed_cache.go) builds it once,
// when the card enters its zone.
//
// offBattlefieldTable returns a two-seat game with n cards in every
// hand and graveyard and 2n in exile, a mix of the shapes that took
// each parse: a creature with subtypes, a basic land (the intrinsic-
// ability parse), a legendary one (supertypes) and an instant.
func offBattlefieldTable(t testing.TB, n int) *game.Game {
	t.Helper()
	g := game.NewGame()
	for i := range 2 {
		deck := []game.Card{game.NewCommander(fmt.Sprintf("Commander %d", i+1), uuid.Nil)}
		for j := range 10 {
			deck = append(deck, game.NewCard(fmt.Sprintf("Filler %d", j+1), uuid.Nil))
		}
		if _, err := g.AddPlayer(fmt.Sprintf("P%d", i+1), deck); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(1, 2))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	shapes := []game.Card{
		{Name: "Elvish Visionary", TypeLine: "Creature — Elf Shaman", ManaCost: "{1}{G}", Power: 1, Toughness: 1},
		basicLand("Forest", "Forest"),
		{Name: "Isamaru", TypeLine: "Legendary Creature — Dog", ManaCost: "{W}", Power: 2, Toughness: 2},
		{Name: "Shock", TypeLine: "Instant", ManaCost: "{R}"},
	}
	g.WithWriteLock(func() {
		for _, p := range g.Seats {
			for i := 0; i < n; i++ {
				shape := shapes[i%len(shapes)]
				shape.Name = fmt.Sprintf("%s %d", shape.Name, i)
				put(p.Hand, p, shape)
				put(p.Graveyard, p, shape)
				put(g.Exile, p, shape)
			}
		}
	})
	return g
}

// parsesPerSteadyView builds two views of the table and returns how
// many times the SECOND one called ParseTypeLine: the first is the one
// a new game pays once, the second is the one a table pays every frame.
func parsesPerSteadyView(t *testing.T, n int) int64 {
	t.Helper()
	g := offBattlefieldTable(t, n)
	_ = ViewOfGame(g)
	before := game.ParseTypeLineCallsForTest()
	_ = ViewOfGame(g)
	return game.ParseTypeLineCallsForTest() - before
}

func TestSteadyViewDoesNotReparseOffBattlefieldTypeLines(t *testing.T) {
	small, large := parsesPerSteadyView(t, 8), parsesPerSteadyView(t, 32)
	t.Logf("ParseTypeLine calls per steady view: %d with 8 cards per zone, %d with 32", small, large)
	// The shape: quadrupling the cards off the battlefield must not
	// change the count. Before the cache it was several parses per
	// card per view, growing with the zones.
	if large > small {
		t.Errorf("quadrupling the off-battlefield cards took ParseTypeLine from %d to %d calls per view; want no growth (the cached parse)", small, large)
	}
	// The size: the card view, the P/T accessors, the type-line render
	// and the land's intrinsic ability all read the cache.
	if large > 4 {
		t.Errorf("a steady view called ParseTypeLine %d times; want ~0", large)
	}
}

// BenchmarkViewOfGameOffBattlefield measures one view of a table whose
// cards are mostly off the battlefield — 32 per hand and graveyard and
// 64 in exile per seat — the shape the printed-characteristic cache is
// for. BenchmarkViewOfGame (legal_moves_test.go) is the busy-
// battlefield one.
func BenchmarkViewOfGameOffBattlefield(b *testing.B) {
	g := offBattlefieldTable(b, 32)
	_ = ViewOfGame(g)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		_ = ViewOfGame(g)
	}
}
