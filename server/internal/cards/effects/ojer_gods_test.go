package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ojer_gods_test.go — returning a card to the battlefield transformed
// from a graveyard (#1900, ADR 0079 amendment). The engine verb is
// pinned in game/transform_test.go; what is pinned here is that the
// Ojer gods reach it and that the back faces behave.

func ojerPakpatiqRow() cards.Card {
	return transformRow(ojerPakpatiqOracleID,
		"Ojer Pakpatiq, Deepest Epoch", "Legendary Creature — God", "{2}{U}{U}",
		"Temple of Cyclical Time", "Land", "", "", []string{"U"})
}

func ojerAxonilRow() cards.Card {
	return transformRow(ojerAxonilOracleID,
		"Ojer Axonil, Deepest Might", "Legendary Creature — God", "{2}{R}{R}",
		"Temple of Power", "Land", "", "", []string{"R"})
}

// ojerOnBattlefield finds the (single) permanent of the oracle ID.
func ojerOnBattlefield(g *game.Game, oracle string) *game.Card {
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].OracleID == oracle {
			return &g.Battlefield.Cards[i]
		}
	}
	return nil
}

func ojerKill(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	var err error
	g.WithWriteLock(func() { err = g.DestroyPermanentForEffect(id) })
	if err != nil {
		t.Fatalf("DestroyPermanentForEffect: %v", err)
	}
}

// The headline: Pakpatiq dies, its trigger goes on the stack, and it
// comes back tapped, transformed, with three time counters.
func TestOjerPakpatiqReturnsAsTheTempleWithThreeTimeCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	god := importToBattlefield(t, g, ojerPakpatiqRow(), me)

	ojerKill(t, g, god)
	if !me.Graveyard.Contains(god) {
		t.Fatal("setup: the God did not die")
	}
	if ojerOnBattlefield(g, ojerPakpatiqOracleID) != nil {
		t.Fatal("it came back before its trigger resolved")
	}
	passPriorityAroundTable(t, g)

	temple := ojerOnBattlefield(g, ojerPakpatiqOracleID)
	if temple == nil {
		t.Fatal("the God did not return")
	}
	if temple.ActiveFace != 1 || temple.Name != "Temple of Cyclical Time" || !temple.IsLand() {
		t.Errorf("returned as face %d %q land=%v, want the Temple", temple.ActiveFace, temple.Name, temple.IsLand())
	}
	if !temple.Tapped {
		t.Error("it did not enter tapped")
	}
	if temple.Counters["time"] != 3 {
		t.Errorf("time counters = %d, want 3", temple.Counters["time"])
	}
	if temple.Controller != me.ID {
		t.Errorf("controller %v, want its owner", temple.Controller)
	}
}

// If the card leaves the graveyard while the trigger waits (CR 400.7),
// the trigger does nothing: it is not returned from exile.
func TestOjerReturnDoesNothingWhenTheCardLeftTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	god := importToBattlefield(t, g, ojerAxonilRow(), me)
	ojerKill(t, g, god)
	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(god); err != nil {
			t.Fatalf("exile the card in response: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	if ojerOnBattlefield(g, ojerAxonilOracleID) != nil {
		t.Error("a card exiled in response still came back")
	}
	if !g.Exile.Contains(god) {
		t.Error("the card left exile")
	}
}

// Axonil returns as Temple of Power with no counters, tapped, and taps
// for {R}.
func TestOjerAxonilReturnsAsTempleOfPower(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	god := importToBattlefield(t, g, ojerAxonilRow(), me)
	ojerKill(t, g, god)
	passPriorityAroundTable(t, g)

	temple := ojerOnBattlefield(g, ojerAxonilOracleID)
	if temple == nil || temple.ActiveFace != 1 || temple.Name != "Temple of Power" || !temple.Tapped {
		t.Fatalf("returned as %+v, want a tapped Temple of Power", temple)
	}
	if len(temple.Counters) != 0 {
		t.Errorf("counters = %v, want none", temple.Counters)
	}
	id := temple.InstanceID
	g.WithWriteLock(func() { ojerOnBattlefield(g, ojerAxonilOracleID).Tapped = false })
	me.ManaPool.EmptyPool()
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap Temple of Power: %v", err)
	}
	if got := poolColors(me); len(got) != 1 || got[0] != "R" {
		t.Errorf("pool %v, want {R}", got)
	}
}

// Temple of Cyclical Time: each tap adds {U} and removes a time
// counter; a land with none left still taps and removes nothing; the
// transform needs no counters and sorcery timing.
func TestTempleOfCyclicalTime(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	god := importToBattlefield(t, g, ojerPakpatiqRow(), me)
	ojerKill(t, g, god)
	passPriorityAroundTable(t, g)
	temple := ojerOnBattlefield(g, ojerPakpatiqOracleID)
	if temple == nil {
		t.Fatal("no Temple")
	}
	id := temple.InstanceID
	advanceTo(t, g, game.StepPrecombatMain)

	tap := func() {
		t.Helper()
		g.WithWriteLock(func() { ojerOnBattlefield(g, ojerPakpatiqOracleID).Tapped = false })
		me.ManaPool.EmptyPool()
		if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
			t.Fatalf("tap the Temple: %v", err)
		}
		if got := poolColors(me); len(got) != 1 || got[0] != "U" {
			t.Fatalf("pool %v, want {U}", got)
		}
	}

	tap()
	if got := ojerOnBattlefield(g, ojerPakpatiqOracleID).Counters["time"]; got != 2 {
		t.Fatalf("time counters after one tap = %d, want 2", got)
	}

	// Counters left: the transform ability is not offered.
	g.WithWriteLock(func() { ojerOnBattlefield(g, ojerPakpatiqOracleID).Tapped = false })
	fillPoolColored(me, "U", 3)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("transformed with time counters still on it")
	}

	tap()
	tap()
	if got := ojerOnBattlefield(g, ojerPakpatiqOracleID).Counters["time"]; got != 0 {
		t.Fatalf("time counters after three taps = %d, want 0", got)
	}
	tap() // nothing left to remove
	if got := ojerOnBattlefield(g, ojerPakpatiqOracleID).Counters["time"]; got != 0 {
		t.Errorf("time counters went to %d", got)
	}

	g.WithWriteLock(func() { ojerOnBattlefield(g, ojerPakpatiqOracleID).Tapped = false })
	me.ManaPool.EmptyPool()
	fillPoolColored(me, "U", 3)
	s58p6Activate(t, g, me.ID, id, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	back := ojerOnBattlefield(g, ojerPakpatiqOracleID)
	if back == nil || back.ActiveFace != 0 || back.Name != "Ojer Pakpatiq, Deepest Epoch" {
		t.Fatalf("after the transform: %+v, want the God again", back)
	}
}

// "Whenever you cast an instant spell from your hand, it gains
// rebound": an instant is exiled as it resolves; a sorcery is not.
func TestOjerPakpatiqGivesAnInstantRebound(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceTo(t, g, game.StepPrecombatMain)
	god := importToBattlefield(t, g, ojerPakpatiqRow(), me)

	id := castPlainSpell(t, g, me, "Instant")
	if n := triggersOnStackFrom(g, god); n != 1 {
		t.Fatalf("Pakpatiq triggers on the stack = %d, want 1", n)
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(id) || me.Graveyard.Contains(id) {
		t.Fatal("the instant was not exiled as it resolved")
	}
	if n := reboundTriggersFor(g, id); n != 1 {
		t.Errorf("rebound triggers = %d, want 1", n)
	}

	sorcery := castPlainSpell(t, g, me, "Sorcery")
	if n := triggersOnStackFrom(g, god); n != 0 {
		t.Errorf("a sorcery triggered Pakpatiq (%d)", n)
	}
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(sorcery) {
		t.Error("the sorcery did not go to the graveyard")
	}
}
