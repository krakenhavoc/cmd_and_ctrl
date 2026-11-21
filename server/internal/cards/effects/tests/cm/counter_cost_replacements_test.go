package testscm

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// lastCounterPlacedAmount reads the NEW TOTAL (`Event.Amount`, the way
// applyCounterByLocked emits it) off the most recent EventCounterPlaced
// for `cardID` / `kind`, for a permanent that may not survive the
// placement to have its live counters read back (Winding Constrictor's
// extra counter on a 0-toughness-bound Devoted Druid, below).
func lastCounterPlacedAmount(g *game.Game, cardID uuid.UUID, kind string) (int, bool) {
	for i := len(g.Events) - 1; i >= 0; i-- {
		ev := g.Events[i]
		if ev.Kind == game.EventCounterPlaced && ev.Target == cardID && ev.Label == kind {
			return ev.Amount, true
		}
	}
	return 0, false
}

// --- Devoted Druid's -1/-1 cost -----------------------------------------

// TestDevotedDruidPlusVizierOfRemediesUntapsForFree is the
// long-standing infinite-mana combo: Vizier reduces the Druid's own
// -1/-1 counter to zero, so nothing is ever placed and the Druid can
// untap indefinitely.
func TestDevotedDruidPlusVizierOfRemediesUntapsForFree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.CcrPushReplacement(g, me.ID, "Vizier of Remedies", "Creature — Human Monk", effectstest.CcrVizierOracle)
	druid := effectstest.B12Push(g, me.ID, "Devoted Druid", "Creature — Elf Druid", effectstest.DevotedDruidOracle, 0, 2)
	effectstest.AdvanceToMain(t, g)

	for i := 0; i < 3; i++ {
		effectstest.B16Tap(g, druid)
		effectstest.B16Activate(t, g, me.ID, druid, 0, game.ActivateAbilityParams{})
		if got := effectstest.CounterCount(g, druid, game.CounterMinusOne); got != 0 {
			t.Fatalf("round %d: -1/-1 counters on the Druid: %d, want 0 (Vizier of Remedies)", i, got)
		}
		if effectstest.B16Tapped(t, g, druid) {
			t.Fatalf("round %d: the Druid did not untap", i)
		}
	}
}

// TestDevotedDruidPlusDoublingSeasonExactlyOneCounter is CR 614.16's
// other half: Doubling Season names "an effect", so it does not touch
// the Druid's cost.
func TestDevotedDruidPlusDoublingSeasonExactlyOneCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.CcrPushReplacement(g, me.ID, "Doubling Season", "Enchantment", effectstest.CcrDoublingSeason)
	druid := effectstest.B12Push(g, me.ID, "Devoted Druid", "Creature — Elf Druid", effectstest.DevotedDruidOracle, 0, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.B16Tap(g, druid)

	effectstest.B16Activate(t, g, me.ID, druid, 0, game.ActivateAbilityParams{})
	if got := effectstest.CounterCount(g, druid, game.CounterMinusOne); got != 1 {
		t.Errorf("-1/-1 counters under Doubling Season: %d, want 1 — a cost is not an effect", got)
	}
	if effectstest.B16Tapped(t, g, druid) {
		t.Error("the Druid did not untap")
	}
}

// TestDevotedDruidPlusWindingConstrictorTwoCounters: Winding
// Constrictor names no effect, so it DOES add one to the Druid's cost
// — which is a real drawback on a printed 0/2, since two -1/-1
// counters is exactly lethal (CR 704.5g). The counter count is read
// off the emitted event rather than the live board, because the
// creature does not survive to have its counters read back.
func TestDevotedDruidPlusWindingConstrictorTwoCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.CcrPushReplacement(g, me.ID, "Winding Constrictor", "Creature — Snake", effectstest.CcrWindingOracle)
	druid := effectstest.B12Push(g, me.ID, "Devoted Druid", "Creature — Elf Druid", effectstest.DevotedDruidOracle, 0, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.B16Tap(g, druid)

	if err := g.ActivateCatalogAbility(me.ID, druid, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	got, ok := lastCounterPlacedAmount(g, druid, game.CounterMinusOne)
	if !ok {
		t.Fatal("no EventCounterPlaced was emitted for the Druid")
	}
	if got != 2 {
		t.Errorf("-1/-1 counters under Winding Constrictor: %d, want 2", got)
	}
	// The whole point: a printed 0/2 with two -1/-1 counters is 0/0 and
	// dies to the state-based check (CR 704.5g) before it can untap.
	if _, ok := effectstest.BattlefieldCard(g, druid); ok {
		t.Error("the Druid survived two -1/-1 counters on a 0/2 body")
	}
}

// --- the loyalty cost ----------------------------------------------------

// TestLoyaltyPlusVorinclexDoublesToPlusTwo: "if you would put one or
// more counters on a permanent or player, put twice that many" names
// no effect, so it doubles a loyalty ability's cost the same as it
// would any other placement.
func TestLoyaltyPlusVorinclexDoublesToPlusTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.CcrPushReplacement(g, me.ID, "Vorinclex, Monstrous Raider", "Legendary Creature — Phyrexian Beast", effectstest.CcrVorinclexOracle)
	pw := effectstest.CcrLoyaltyWalker(g, me.ID, 4, 1)
	effectstest.AdvanceToMain(t, g)

	if err := g.ActivateCatalogAbility(me.ID, pw, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	if got := effectstest.CounterCount(g, pw, game.CounterLoyalty); got != 6 {
		t.Errorf("loyalty after a +1 ability under Vorinclex: %d, want 6 (4 + 2)", got)
	}
}

// TestLoyaltyPlusDoublingSeasonStaysPlusOne is the ruling #1710 asked
// to have checked rather than assumed: Doubling Season's Gatherer
// ruling says a loyalty ability's cost is "put on as a cost, not as an
// effect", so a `+1` stays `+1` — the same reading Devoted Druid's
// cost gets, and the reason `payCostCounterLocked` marks
// CounterFromCost for both.
func TestLoyaltyPlusDoublingSeasonStaysPlusOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.CcrPushReplacement(g, me.ID, "Doubling Season", "Enchantment", effectstest.CcrDoublingSeason)
	pw := effectstest.CcrLoyaltyWalker(g, me.ID, 4, 1)
	effectstest.AdvanceToMain(t, g)

	if err := g.ActivateCatalogAbility(me.ID, pw, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	if got := effectstest.CounterCount(g, pw, game.CounterLoyalty); got != 5 {
		t.Errorf("loyalty after a +1 ability under Doubling Season: %d, want 5 — a cost is not an effect", got)
	}
}

// TestLoyaltyMinusAbilityUnaffectedByVorinclex: a "-N" ability removes
// loyalty counters, and CR 614.1's placement replacements — Vorinclex
// included — have nothing to say about a removal.
func TestLoyaltyMinusAbilityUnaffectedByVorinclex(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.CcrPushReplacement(g, me.ID, "Vorinclex, Monstrous Raider", "Legendary Creature — Phyrexian Beast", effectstest.CcrVorinclexOracle)
	pw := effectstest.CcrLoyaltyWalker(g, me.ID, 6, -3)
	effectstest.AdvanceToMain(t, g)

	if err := g.ActivateCatalogAbility(me.ID, pw, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	if got := effectstest.CounterCount(g, pw, game.CounterLoyalty); got != 3 {
		t.Errorf("loyalty after -3 under Vorinclex: %d, want 3 — a removal is not a placement", got)
	}
}

// TestSandboxActivateLoyaltyPlusVorinclexDoublesToPlusTwo is the same
// ruling through the OTHER loyalty path: `Game.ActivateLoyalty`
// (`mutations.go`) is the manual verb for the planeswalkers with no
// catalog entry, and it paid its counter the same direct-write way
// activated.go's catalogued path used to. #1710 is not just the
// catalogued path — a table using the sandbox verb on an uncatalogued
// walker sees Vorinclex too.
func TestSandboxActivateLoyaltyPlusVorinclexDoublesToPlusTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.CcrPushReplacement(g, me.ID, "Vorinclex, Monstrous Raider", "Legendary Creature — Phyrexian Beast", effectstest.CcrVorinclexOracle)
	pw := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Uncatalogued Walker", TypeLine: "Legendary Planeswalker — Test",
		Owner: me.ID, Controller: me.ID, Counters: map[string]int{game.CounterLoyalty: 4},
	})
	effectstest.AdvanceToMain(t, g)

	if err := g.ActivateLoyalty(me.ID, pw, "+1", 1); err != nil {
		t.Fatalf("ActivateLoyalty: %v", err)
	}
	if got := effectstest.CounterCount(g, pw, game.CounterLoyalty); got != 6 {
		t.Errorf("loyalty after a sandbox +1 under Vorinclex: %d, want 6 (4 + 2)", got)
	}
}
