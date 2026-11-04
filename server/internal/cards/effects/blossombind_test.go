package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// blossombind_test.go — #4692's card: "Enchanted creature can't become
// untapped and can't have counters put on it." The engine half is
// game/cant_have_counters_test.go.

const oracleBlossombind = "0ecb35b3-5bd1-4467-a4a3-ce468023ced2"

func TestBlossombindIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(oracleBlossombind)
	if !ok || spec.Name != "Blossombind" || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("Blossombind: registered %v, %+v", ok, spec)
	}
}

// blossombindOn casts Blossombind onto host and resolves it and its entry
// trigger.
func blossombindOn(t *testing.T, g *game.Game, host uuid.UUID) uuid.UUID {
	t.Helper()
	aura := castCatalogSpell(t, g, "Blossombind", auraTypeLine, oracleBlossombind,
		[]game.TargetRef{{Kind: game.TargetCard, ID: host}})
	passPriorityAroundTable(t, g)
	if got := attachmentHostOf(t, g, aura); got.Kind != game.TargetCard || got.ID != host {
		t.Fatalf("Blossombind attached to %+v, want %s", got, host)
	}
	return aura
}

func TestBlossombindStopsCountersUntilItLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	host := pushPermanentForTest(g, me.ID, "Host", "", "Creature — Bear")
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(host, game.CounterPlusOne, 1) })
	// Two counter replacements that would both apply, so a placement that
	// reached the window would open a CR 616 ordering prompt.
	seedReplacementPermanent(g, hardenedScalesOracle, "Hardened Scales", me.ID)
	seedReplacementPermanent(g, doublingSeasonOracle, "Doubling Season", me.ID)
	aura := blossombindOn(t, g, host)
	if !tappedForTest(t, g, host) {
		t.Fatal("the entry trigger did not tap the enchanted creature")
	}
	untapSpell(t, g, host)
	if !tappedForTest(t, g, host) {
		t.Fatal("an untap effect untapped the enchanted creature")
	}

	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(host, game.CounterPlusOne, 1); err != nil {
			t.Fatalf("AddCounterForEffect: %v", err)
		}
	})
	if n := len(g.PendingChoices); n != 0 {
		t.Fatalf("%d prompts opened for a placement that can't happen; no replacement may see it", n)
	}
	if got := plusOnes(g, host); got != 1 {
		t.Fatalf("+1/+1 counters = %d, want the 1 it already had", got)
	}
	var shown bool
	g.ReadSnapshot(func() { shown = g.CantHaveCountersForEffect(host) })
	if !shown {
		t.Error("the restriction is not readable for the view")
	}

	// CR 122.5: a move onto it moves nothing, so the source keeps its counter.
	mage := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Simic Fluxmage", OracleID: oracleSimicFluxmage,
		TypeLine: "Creature — Merfolk Wizard", Power: 1, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	g.WithWriteLock(func() { _, _ = g.AddCounterMustSettleNowForEffect(mage, game.CounterPlusOne, 1) })
	before := plusOnes(g, mage)
	item := &game.StackItem{
		Kind: game.StackItemActivated, Controller: me.ID, Owner: me.ID, SourceCardID: mage,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: host}},
	}
	g.WithWriteLock(func() {
		if err := simicFluxmageMove(g, item); err != nil {
			t.Fatalf("move: %v", err)
		}
	})
	if a, b := plusOnes(g, mage), plusOnes(g, host); a != before || b != 1 {
		t.Errorf("after a move onto it: Fluxmage %d (want %d), host %d (want 1)", a, before, b)
	}

	// Once the Aura leaves, counters go on again.
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, aura); err != nil {
		t.Fatal(err)
	}
	if err := g.AddCounter(host, game.CounterMinusOne, 1); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	if got := counterCount(g, host, game.CounterMinusOne); got == 0 {
		t.Error("no counter landed after Blossombind left")
	}
}

func TestBlossombindStopsDevotedDruidsCost(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	druid := b12Push(g, me.ID, "Devoted Druid", "Creature — Elf Druid", devotedDruidOracle, 0, 2)
	advanceToMain(t, g)
	blossombindOn(t, g, druid)
	if err := g.ActivateCatalogAbility(me.ID, druid, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("Devoted Druid paid a -1/-1 counter cost it can't have put on it (CR 118.3)")
	}
	if got := counterCount(g, druid, game.CounterMinusOne); got != 0 {
		t.Errorf("-1/-1 counters on the Druid: %d, want 0", got)
	}
}
