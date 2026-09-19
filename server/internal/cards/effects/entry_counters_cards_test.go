package effects

import (
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// entry_counters_cards_test.go — #1002, the CARD half.
//
// The mechanic is pinned in game/entry_counters_test.go: the counters
// are seeded onto the entry event before the CR 614 pipeline, so they
// land on the permanent between the move and EventETB. What is here is
// the fourteen cards that declare the clause, and the two things they
// are the proof of: every one of them really does enter with its
// counters through that road, and none of them still tells players the
// counters go on a beat early.

// entryCounterCard is one row of the table: the card, the counter it
// enters with, and how many are on it once its own entry triggers have
// finished with them.
type entryCounterCard struct {
	name     string
	typeLine string
	oracle   string
	manaCost string
	kind     string
	// want is the count after the dust settles, 3 for every card that
	// only enters. Voracious Hydra's enters trigger doubles its own
	// counters, and that it CAN is the point: the trigger is harvested
	// off the EventETB the entry emits after the counters have landed.
	want int
}

// everyCardWithCastEntryCounters is the whole population, taken from
// the registry rather than typed out, so a fourteenth card cannot be
// added without this test noticing.
func everyCardWithCastEntryCounters(t *testing.T) []Spec {
	t.Helper()
	var out []Spec
	for _, spec := range All() {
		if len(spec.EntersWithCountersFromCast) > 0 {
			out = append(out, spec)
		}
	}
	if len(out) == 0 {
		t.Fatal("no card declares EntersWithCountersFromCast — the scan has rotted")
	}
	return out
}

// The thirteen X cards. Cast for X=3, each enters with three counters
// ON THE BATTLEFIELD — the assertion the old OnResolve could pass only
// because the counters travelled with the card.
func TestXEntryCounterCardsEnterWithTheirCounters(t *testing.T) {
	cards := []entryCounterCard{
		{"Hangarback Walker", "Artifact Creature — Construct", "dde55256-5259-44e7-a267-fca45a7f0d04", "{X}{X}", game.CounterPlusOne, 3},
		{"Marketback Walker", "Artifact Creature — Construct", "0405e0a9-6d02-4691-bdb8-59c72b824dab", "{X}{X}", game.CounterPlusOne, 3},
		{"Mikaeus, the Lunarch", "Legendary Creature — Human Cleric", "82f3faa8-39fa-450b-843f-d60a4c36d8f7", "{X}{W}", game.CounterPlusOne, 3},
		{"Goldvein Hydra", "Creature — Hydra", "2b62543f-a475-457a-a96b-b5d070383d3c", "{X}{G}", game.CounterPlusOne, 3},
		{"Primordial Hydra", "Creature — Hydra", "1c36ed3a-c806-47e5-83f9-e44999c67fe5", "{X}{G}{G}", game.CounterPlusOne, 3},
		{"Lifeblood Hydra", "Creature — Hydra", "b14d05c0-fe10-4079-a90e-0aea1a8fd375", "{X}{G}{G}{G}", game.CounterPlusOne, 3},
		{"Hydroid Krasis", "Creature — Jellyfish Hydra Beast", "6bd872b2-5c40-4e11-9a7f-0136a51b0642", "{X}{G}{U}", game.CounterPlusOne, 3},
		{"Wildwood Scourge", "Creature — Hydra", "b9dec104-c636-4770-a7fc-7a3331face15", "{X}{G}", game.CounterPlusOne, 3},
		{"Benevolent Hydra", "Creature — Hydra", "01dbf1bc-ca62-4fb6-959c-ef7c0dc03bb0", "{X}{G}{G}", game.CounterPlusOne, 3},
		{"Shivan Devastator", "Creature — Dragon Hydra", "b7daa74c-6142-4107-9355-be98af6ccf13", "{X}{R}", game.CounterPlusOne, 3},
		{"Voracious Hydra", "Creature — Hydra", "ff8f5a4b-112a-425e-b489-7ee26d1d9fb3", "{X}{G}{G}", game.CounterPlusOne, 6},
		{"The Goose Mother", "Legendary Creature — Bird Hydra", "de595f1b-3f7d-45e0-a31b-ed23e5d1ee48", "{X}{G}{U}", game.CounterPlusOne, 3},
		{"Fated Firepower", "Enchantment", "13daa21c-278d-45bd-9a6e-a77d6a558453", "{X}{R}{R}{R}", "fire", 3},
	}
	if len(cards) != 13 {
		t.Fatalf("the table is the thirteen cards of #1002, got %d", len(cards))
	}
	for _, tc := range cards {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			id := castXSpell(t, g, tc.name, tc.typeLine, tc.oracle, tc.manaCost, 3, nil)
			passPriorityAroundTable(t, g)
			if !g.Battlefield.Contains(id) {
				t.Fatalf("%s did not reach the battlefield", tc.name)
			}
			if got := counterCount(g, id, tc.kind); got != tc.want {
				t.Errorf("%s X=3: %d %s counters, want %d", tc.name, got, tc.kind, tc.want)
			}
		})
	}
}

// Etched Oracle is the fourteenth card in the slot — the same clause
// counting colours of mana spent instead of X (CR 702.44a). Its proof
// is TestEtchedOracleEntersWithSunburstCountersAndDrawsThree in
// counter_costs_and_mana_spent_test.go, which needs a strict payment
// to have colours to count and was already asserting the counters are
// on the permanent when it lands.

// Not one of the fourteen still tells players the counters go on a
// beat before the permanent enters. #412's lesson: a caveat goes stale
// the day somebody lands the mechanic, in a file the author never
// opens, so the sweep is part of the landing.
func TestNoEntryCounterCardStillDeclaresTheOnResolveGap(t *testing.T) {
	for _, spec := range everyCardWithCastEntryCounters(t) {
		for _, c := range spec.Caveats {
			if strings.Contains(c, "a beat before") || strings.Contains(c, "a moment before") {
				t.Errorf("%s still declares the OnResolve timing gap #1002 closed: %q", spec.Name, c)
			}
		}
	}
}

// Every card in the slot declares a counter kind and a count function.
// A half-written clause would silently put nothing on.
func TestEveryCastEntryCounterClauseIsComplete(t *testing.T) {
	for _, spec := range everyCardWithCastEntryCounters(t) {
		for i, clause := range spec.EntersWithCountersFromCast {
			if clause.Kind == "" {
				t.Errorf("%s clause %d has no counter kind", spec.Name, i)
			}
			if clause.Count == nil {
				t.Errorf("%s clause %d has no count", spec.Name, i)
			}
		}
	}
}

// THE PAYOFF, on real cards. Doubling Season is a counter replacement,
// and the entry counters now go through the counter window on a
// permanent, so it doubles them.
func TestDoublingSeasonDoublesGoldveinHydrasEntryCounters(t *testing.T) {
	const goldvein = "2b62543f-a475-457a-a96b-b5d070383d3c"
	const doublingSeason = "01546b7d-a233-4176-8843-d732074dc5b6"
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushCatalogPermanent(g, me.ID, "Doubling Season", "Enchantment", doublingSeason, false)

	id := castXSpell(t, g, "Goldvein Hydra", "Creature — Hydra", goldvein, "{X}{G}", 3, nil)
	passPriorityAroundTable(t, g)

	if got := counterCount(g, id, game.CounterPlusOne); got != 6 {
		t.Errorf("Goldvein Hydra X=3 under Doubling Season: %d counters, want 6", got)
	}
}

// Hardened Scales, the +1 shape, on the same road.
func TestHardenedScalesAddsToGoldveinHydrasEntryCounters(t *testing.T) {
	const goldvein = "2b62543f-a475-457a-a96b-b5d070383d3c"
	const hardenedScales = "a1f3da21-af6d-450e-bf0b-985d158418e6"
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushCatalogPermanent(g, me.ID, "Hardened Scales", "Enchantment", hardenedScales, false)

	id := castXSpell(t, g, "Goldvein Hydra", "Creature — Hydra", goldvein, "{X}{G}", 3, nil)
	passPriorityAroundTable(t, g)

	if got := counterCount(g, id, game.CounterPlusOne); got != 4 {
		t.Errorf("Goldvein Hydra X=3 under Hardened Scales: %d counters, want 4", got)
	}
}

// The sentence every one of the thirteen caveats was about: "whenever
// you put one or more counters on a permanent" now sees the entry
// counters, because there is a permanent by the time they land. All
// Will Be One turns three +1/+1 counters into three damage.
func TestAllWillBeOneSeesTheEntryCounters(t *testing.T) {
	const goldvein = "2b62543f-a475-457a-a96b-b5d070383d3c"
	const allWillBeOne = "477374dc-042c-48f7-9ebe-99c15d8ae04f"
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "All Will Be One", "Enchantment", allWillBeOne, false)
	before := opp.Life

	castXSpell(t, g, "Goldvein Hydra", "Creature — Hydra", goldvein, "{X}{G}", 3, nil)
	passPriorityAroundTable(t, g)
	// The trigger targets: answer the pick with the opponent, then
	// resolve it.
	b04WaitForPick(t, g, me.ID)
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)

	if opp.Life != before-3 {
		t.Errorf("All Will Be One on three entry counters: life %d -> %d, want -3", before, opp.Life)
	}
}
