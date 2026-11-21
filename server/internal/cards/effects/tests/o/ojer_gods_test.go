package testso

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	cardso "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/o"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ojer_gods_test.go — returning a card to the battlefield transformed
// from a graveyard (#1900, ADR 0079 amendment). The engine verb is
// pinned in game/transform_test.go; what is pinned here is that the
// Ojer gods reach it and that the back faces behave.

func ojerPakpatiqRow() cards.Card {
	return effectstest.TransformRow(cardso.OjerPakpatiqOracleID,
		"Ojer Pakpatiq, Deepest Epoch", "Legendary Creature — God", "{2}{U}{U}",
		"Temple of Cyclical Time", "Land", "", "", []string{"U"})
}

func ojerAxonilRow() cards.Card {
	return effectstest.TransformRow(cardso.OjerAxonilOracleID,
		"Ojer Axonil, Deepest Might", "Legendary Creature — God", "{2}{R}{R}",
		"Temple of Power", "Land", "", "", []string{"R"})
}

// The headline: Pakpatiq dies, its trigger goes on the stack, and it
// comes back tapped, transformed, with three time counters.
func TestOjerPakpatiqReturnsAsTheTempleWithThreeTimeCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	god := effectstest.ImportToBattlefield(t, g, ojerPakpatiqRow(), me)

	effectstest.OjerKill(t, g, god)
	if !me.Graveyard.Contains(god) {
		t.Fatal("setup: the God did not die")
	}
	if effectstest.OjerOnBattlefield(g, cardso.OjerPakpatiqOracleID) != nil {
		t.Fatal("it came back before its trigger resolved")
	}
	effectstest.PassPriorityAroundTable(t, g)

	temple := effectstest.OjerOnBattlefield(g, cardso.OjerPakpatiqOracleID)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	god := effectstest.ImportToBattlefield(t, g, ojerAxonilRow(), me)
	effectstest.OjerKill(t, g, god)
	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(god); err != nil {
			t.Fatalf("exile the card in response: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OjerOnBattlefield(g, cardso.OjerAxonilOracleID) != nil {
		t.Error("a card exiled in response still came back")
	}
	if !g.Exile.Contains(god) {
		t.Error("the card left exile")
	}
}

// Axonil returns as Temple of Power with no counters, tapped, and taps
// for {R}.
func TestOjerAxonilReturnsAsTempleOfPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	god := effectstest.ImportToBattlefield(t, g, ojerAxonilRow(), me)
	effectstest.OjerKill(t, g, god)
	effectstest.PassPriorityAroundTable(t, g)

	temple := effectstest.OjerOnBattlefield(g, cardso.OjerAxonilOracleID)
	if temple == nil || temple.ActiveFace != 1 || temple.Name != "Temple of Power" || !temple.Tapped {
		t.Fatalf("returned as %+v, want a tapped Temple of Power", temple)
	}
	if len(temple.Counters) != 0 {
		t.Errorf("counters = %v, want none", temple.Counters)
	}
	id := temple.InstanceID
	g.WithWriteLock(func() { effectstest.OjerOnBattlefield(g, cardso.OjerAxonilOracleID).Tapped = false })
	me.ManaPool.EmptyPool()
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap Temple of Power: %v", err)
	}
	if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != "R" {
		t.Errorf("pool %v, want {R}", got)
	}
}

// Temple of Power (#2662): the transform needs 4 noncombat damage from
// red sources its controller controlled this turn, and sorcery timing.
// A green source's damage does not count; 3 + 1 from red sources does.
func TestTempleOfPowerTransformsAfterFourRedNoncombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	god := effectstest.ImportToBattlefield(t, g, ojerAxonilRow(), me)
	effectstest.OjerKill(t, g, god)
	effectstest.PassPriorityAroundTable(t, g)
	temple := effectstest.OjerOnBattlefield(g, cardso.OjerAxonilOracleID)
	if temple == nil {
		t.Fatal("no Temple")
	}
	id := temple.InstanceID
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)

	red := effectstest.PushVanillaCreature(g, me.ID, "Red Pinger", 1, 1)
	green := effectstest.PushVanillaCreature(g, me.ID, "Green Pinger", 1, 1)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			switch g.Battlefield.Cards[i].InstanceID {
			case red:
				g.Battlefield.Cards[i].Colors = []string{"R"}
			case green:
				g.Battlefield.Cards[i].Colors = []string{"G"}
			}
		}
	})
	ping := func(src uuid.UUID, n int) {
		t.Helper()
		var err error
		g.WithWriteLock(func() { err = g.DealDamageToPlayerForEffect(src, opp.ID, n) })
		if err != nil {
			t.Fatalf("ping: %v", err)
		}
	}
	try := func() error {
		g.WithWriteLock(func() { effectstest.OjerOnBattlefield(g, cardso.OjerAxonilOracleID).Tapped = false })
		me.ManaPool.EmptyPool()
		effectstest.FillPoolColored(me, "R", 3)
		return g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{})
	}

	ping(green, 5)
	ping(red, 3)
	if err := try(); err == nil {
		t.Fatal("transformed with 3 red noncombat damage (and 5 green)")
	}
	ping(red, 1)
	if err := try(); err != nil {
		t.Fatalf("activate with 4 red noncombat damage: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	back := effectstest.OjerOnBattlefield(g, cardso.OjerAxonilOracleID)
	if back == nil || back.ActiveFace != 0 || back.Name != "Ojer Axonil, Deepest Might" {
		t.Fatalf("after the transform: %+v, want the God again", back)
	}
}

// Temple of Cyclical Time: each tap adds {U} and removes a time
// counter; a land with none left still taps and removes nothing; the
// transform needs no counters and sorcery timing.
func TestTempleOfCyclicalTime(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	god := effectstest.ImportToBattlefield(t, g, ojerPakpatiqRow(), me)
	effectstest.OjerKill(t, g, god)
	effectstest.PassPriorityAroundTable(t, g)
	temple := effectstest.OjerOnBattlefield(g, cardso.OjerPakpatiqOracleID)
	if temple == nil {
		t.Fatal("no Temple")
	}
	id := temple.InstanceID
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)

	tap := func() {
		t.Helper()
		g.WithWriteLock(func() { effectstest.OjerOnBattlefield(g, cardso.OjerPakpatiqOracleID).Tapped = false })
		me.ManaPool.EmptyPool()
		if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
			t.Fatalf("tap the Temple: %v", err)
		}
		if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != "U" {
			t.Fatalf("pool %v, want {U}", got)
		}
	}

	tap()
	if got := effectstest.OjerOnBattlefield(g, cardso.OjerPakpatiqOracleID).Counters["time"]; got != 2 {
		t.Fatalf("time counters after one tap = %d, want 2", got)
	}

	// Counters left: the transform ability is not offered.
	g.WithWriteLock(func() { effectstest.OjerOnBattlefield(g, cardso.OjerPakpatiqOracleID).Tapped = false })
	effectstest.FillPoolColored(me, "U", 3)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("transformed with time counters still on it")
	}

	tap()
	tap()
	if got := effectstest.OjerOnBattlefield(g, cardso.OjerPakpatiqOracleID).Counters["time"]; got != 0 {
		t.Fatalf("time counters after three taps = %d, want 0", got)
	}
	tap() // nothing left to remove
	if got := effectstest.OjerOnBattlefield(g, cardso.OjerPakpatiqOracleID).Counters["time"]; got != 0 {
		t.Errorf("time counters went to %d", got)
	}

	g.WithWriteLock(func() { effectstest.OjerOnBattlefield(g, cardso.OjerPakpatiqOracleID).Tapped = false })
	me.ManaPool.EmptyPool()
	effectstest.FillPoolColored(me, "U", 3)
	effectstest.S58p6Activate(t, g, me.ID, id, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	back := effectstest.OjerOnBattlefield(g, cardso.OjerPakpatiqOracleID)
	if back == nil || back.ActiveFace != 0 || back.Name != "Ojer Pakpatiq, Deepest Epoch" {
		t.Fatalf("after the transform: %+v, want the God again", back)
	}
}

// "Whenever you cast an instant spell from your hand, it gains
// rebound": an instant is exiled as it resolves; a sorcery is not.
func TestOjerPakpatiqGivesAnInstantRebound(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	god := effectstest.ImportToBattlefield(t, g, ojerPakpatiqRow(), me)

	id := effectstest.CastPlainSpell(t, g, me, "Instant")
	if n := effectstest.TriggersOnStackFrom(g, god); n != 1 {
		t.Fatalf("Pakpatiq triggers on the stack = %d, want 1", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(id) || me.Graveyard.Contains(id) {
		t.Fatal("the instant was not exiled as it resolved")
	}
	if n := effectstest.ReboundTriggersFor(g, id); n != 1 {
		t.Errorf("rebound triggers = %d, want 1", n)
	}

	sorcery := effectstest.CastPlainSpell(t, g, me, "Sorcery")
	if n := effectstest.TriggersOnStackFrom(g, god); n != 0 {
		t.Errorf("a sorcery triggered Pakpatiq (%d)", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(sorcery) {
		t.Error("the sorcery did not go to the graveyard")
	}
}
