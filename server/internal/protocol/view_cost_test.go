package protocol

import (
	"fmt"
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
// #1261: publicLogOf asks the seat indexer for every event of the game
// on every frame, and it used to format the actor's UUID as a string
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
