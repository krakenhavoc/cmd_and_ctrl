package testsp

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// grantedRowsOf is the labels of a permanent's GRANTED activated rows.
func grantedRowsOf(t *testing.T, g *game.Game, id uuid.UUID) []string {
	t.Helper()
	abs, origins := game.ActivatedAbilitiesWithOrigins(effectstest.GaLayered(t, g, id))
	var out []string
	for i, ab := range abs {
		if origins.At(i).Granted() {
			out = append(out, ab.Label)
		}
	}
	return out
}

// --- 1. a loyalty ability granted to planeswalkers you control ------------

// "Planeswalkers you control have '[−8]: …'" reaches every planeswalker its
// controller controls, one that arrives later, and none an opponent controls;
// and it ends the moment the grantor does.
func TestKioraGrantsTheLoyaltyAbilityToEveryPlaneswalkerYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.FixtureWalker(g, me.ID, effectstest.Pw2797WalkerOracle, "Planeswalker — Fix", 4)
	theirs := effectstest.FixtureWalker(g, opp.ID, effectstest.Pw2797WalkerOracle, "Planeswalker — Fix", 4)
	notAWalker := effectstest.PushCatalogPermanent(g, me.ID, "Bear", "Creature — Bear", "", false)

	if rows := grantedRowsOf(t, g, a); len(rows) != 0 {
		t.Fatalf("a walker with no grantor on the board already has granted rows: %v", rows)
	}
	kiora := effectstest.PushCatalogPermanent(g, me.ID, "Kiora of Salt and Sand", "Legendary Creature — Merfolk Noble", effectstest.KioraOracle2797, false)

	want := "−8: Create an 8/8 blue Leviathan creature token with hexproof."
	if rows := grantedRowsOf(t, g, a); len(rows) != 1 || rows[0] != want {
		t.Errorf("my walker's granted rows = %v, want [%q]", rows, want)
	}
	if rows := grantedRowsOf(t, g, theirs); len(rows) != 0 {
		t.Errorf("an opponent's walker got the grant: %v", rows)
	}
	if rows := grantedRowsOf(t, g, notAWalker); len(rows) != 0 {
		t.Errorf("a creature got a loyalty ability: %v", rows)
	}

	late := effectstest.FixtureWalker(g, me.ID, effectstest.Pw2797WalkerOracle, "Planeswalker — Fix", 3)
	if rows := grantedRowsOf(t, g, late); len(rows) != 1 {
		t.Errorf("a walker that arrived after the grantor has granted rows %v, want one", rows)
	}

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(kiora) })
	for _, id := range []uuid.UUID{a, late} {
		if rows := grantedRowsOf(t, g, id); len(rows) != 0 {
			t.Errorf("the grant outlived Kiora on %s: %v", id, rows)
		}
	}
}

// The granted −8 is the walker's: it costs eight of THIS walker's counters
// (CR 606.6), makes an 8/8 blue Leviathan with hexproof for the activator,
// and sends a walker paid down to zero to the graveyard (CR 704.5i).
func TestKioraGrantedMinusEightMakesALeviathan(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushCatalogPermanent(g, me.ID, "Kiora of Salt and Sand", "Legendary Creature — Merfolk Noble", effectstest.KioraOracle2797, false)
	short := effectstest.FixtureWalker(g, me.ID, effectstest.Pw2797WalkerOracle, "Planeswalker — Fix", 7)
	full := effectstest.FixtureWalker(g, me.ID, effectstest.Pw2797WalkerOracle, "Planeswalker — Fix", 8)

	idx, ref := effectstest.GrantedLoyaltyRow(t, g, short, "−8")
	if err := g.ActivateCatalogAbility(me.ID, short, idx, game.ActivateAbilityParams{Ref: ref}); err != game.ErrInsufficientLoyalty {
		t.Fatalf("−8 with 7 loyalty: got %v, want ErrInsufficientLoyalty", err)
	}
	if got := effectstest.LoyaltyCount(g, short); got != 7 {
		t.Fatalf("a refused −8 moved the loyalty to %d", got)
	}

	idx, ref = effectstest.GrantedLoyaltyRow(t, g, full, "−8")
	effectstest.B16Activate(t, g, me.ID, full, idx, game.ActivateAbilityParams{Ref: ref})
	effectstest.PassPriorityAroundTable(t, g)

	var leviathan *game.Card
	for i := range g.Battlefield.Cards {
		if c := &g.Battlefield.Cards[i]; c.Name == "Leviathan" {
			leviathan = c
		}
	}
	if leviathan == nil {
		t.Fatal("no Leviathan on the battlefield")
	}
	if leviathan.Controller != me.ID || leviathan.Power != 8 || leviathan.Toughness != 8 {
		t.Errorf("Leviathan = %d/%d controlled by %s, want 8/8 mine", leviathan.Power, leviathan.Toughness, leviathan.Controller)
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, leviathan.InstanceID, "hexproof") {
		t.Errorf("Leviathan abilities %v, want hexproof", effectstest.EffectiveAbilities(t, g, leviathan.InstanceID))
	}
	if !strings.Contains(leviathan.TypeLine, "Leviathan") {
		t.Errorf("Leviathan type line = %q", leviathan.TypeLine)
	}
	if effectstest.OnBattlefieldNow(g, full) {
		t.Error("the walker paid all eight loyalty and is still on the battlefield (CR 704.5i)")
	}
}

// CR 606.3 counts per PERMANENT, so the granted row, the walker's own rows
// and a second grantor's row all share one activation a turn, in any order.
func TestClassGrantedLoyaltyAbilitySharesTheWalkersOncePerTurn(t *testing.T) {
	t.Run("own first, then granted", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		effectstest.ToMain(t, g)
		me := g.Seats[g.Turn.ActiveSeat]
		effectstest.PushCatalogPermanent(g, me.ID, "Kiora of Salt and Sand", "Legendary Creature — Merfolk Noble", effectstest.KioraOracle2797, false)
		w := effectstest.FixtureWalker(g, me.ID, effectstest.Pw2797WalkerOracle, "Planeswalker — Fix", 9)
		idx, ref := effectstest.GrantedLoyaltyRow(t, g, w, "−8")

		effectstest.B16Activate(t, g, me.ID, w, 0, game.ActivateAbilityParams{}) // +2
		effectstest.PassPriorityAroundTable(t, g)
		if err := g.ActivateCatalogAbility(me.ID, w, idx, game.ActivateAbilityParams{Ref: ref}); err != game.ErrLoyaltyAlreadyActivated {
			t.Errorf("the granted −8 after the walker's +2: got %v, want ErrLoyaltyAlreadyActivated", err)
		}
	})
	t.Run("granted first, then own", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		effectstest.ToMain(t, g)
		me := g.Seats[g.Turn.ActiveSeat]
		effectstest.PushCatalogPermanent(g, me.ID, "Kiora of Salt and Sand", "Legendary Creature — Merfolk Noble", effectstest.KioraOracle2797, false)
		w := effectstest.FixtureWalker(g, me.ID, effectstest.Pw2797WalkerOracle, "Planeswalker — Fix", 10)
		idx, ref := effectstest.GrantedLoyaltyRow(t, g, w, "−8")

		effectstest.B16Activate(t, g, me.ID, w, idx, game.ActivateAbilityParams{Ref: ref})
		effectstest.PassPriorityAroundTable(t, g)
		for _, own := range []int{0, 1} {
			if err := g.ActivateCatalogAbility(me.ID, w, own, game.ActivateAbilityParams{}); err != game.ErrLoyaltyAlreadyActivated {
				t.Errorf("the walker's own row %d after the granted −8: got %v, want ErrLoyaltyAlreadyActivated", own, err)
			}
		}
	})
	t.Run("two grantors, one activation", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		effectstest.ToMain(t, g)
		me := g.Seats[g.Turn.ActiveSeat]
		effectstest.PushCatalogPermanent(g, me.ID, "Kiora of Salt and Sand", "Legendary Creature — Merfolk Noble", effectstest.KioraOracle2797, false)
		effectstest.PushCatalogPermanent(g, me.ID, "Fixture Granter", "Creature — Fixture", effectstest.Pw2797GranterOracle, false)
		w := effectstest.FixtureWalker(g, me.ID, effectstest.Pw2797WalkerOracle, "Planeswalker — Fix", 20)
		abs, origins := game.ActivatedAbilitiesWithOrigins(effectstest.GaLayered(t, g, w))
		var refs []string
		var idxs []int
		for i := range abs {
			if origins.At(i).Granted() {
				refs = append(refs, origins.Ref(i))
				idxs = append(idxs, i)
			}
		}
		if len(refs) != 2 {
			t.Fatalf("two grantors gave %d granted rows, want 2", len(refs))
		}
		effectstest.B16Activate(t, g, me.ID, w, idxs[0], game.ActivateAbilityParams{Ref: refs[0]})
		effectstest.PassPriorityAroundTable(t, g)
		if err := g.ActivateCatalogAbility(me.ID, w, idxs[1], game.ActivateAbilityParams{Ref: refs[1]}); err != game.ErrLoyaltyAlreadyActivated {
			t.Errorf("the second grantor's row after the first's: got %v, want ErrLoyaltyAlreadyActivated", err)
		}
	})
}

func TestZeroLoyaltyExemptionKeepsYourPlaneswalkersOnTheBattlefield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Fixture Lurker", "Creature — Horror", effectstest.Pw2797ExemptOracle, false)
	mine := effectstest.FixtureWalker(g, me.ID, effectstest.Pw2797WalkerOracle, "Planeswalker — Fix", 1)
	theirs := effectstest.FixtureWalker(g, opp.ID, effectstest.Pw2797WalkerOracle, "Planeswalker — Fix", 1)

	g.WithWriteLock(func() {
		_ = g.AddCounterForEffect(mine, game.CounterLoyalty, -1)
		_ = g.AddCounterForEffect(theirs, game.CounterLoyalty, -1)
	})
	effectstest.RunStateChecks(g)
	if !effectstest.OnBattlefieldNow(g, mine) {
		t.Error("my 0-loyalty planeswalker was put into the graveyard despite the exemption")
	}
	if effectstest.LoyaltyCount(g, mine) != 0 {
		t.Errorf("loyalty = %d, want it to stay at 0", effectstest.LoyaltyCount(g, mine))
	}
	if effectstest.OnBattlefieldNow(g, theirs) {
		t.Error("an opponent's 0-loyalty planeswalker survived my exemption (it covers planeswalkers I control)")
	}
}

// The exemption is a static of the permanent that prints it: when the last
// exempting permanent leaves, the next state check puts every 0-loyalty walker
// it was covering into its owner's graveyard; two of them compose; and one
// that has lost its abilities exempts nothing.
func TestZeroLoyaltyExemptionIsAStaticOfItsSource(t *testing.T) {
	t.Run("leaving", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		a := effectstest.PushCatalogPermanent(g, me.ID, "Fixture Lurker", "Creature — Horror", effectstest.Pw2797ExemptOracle, false)
		b := effectstest.PushCatalogPermanent(g, me.ID, "Fixture Lurker", "Creature — Horror", effectstest.Pw2797ExemptOracle, false)
		w := effectstest.FixtureWalker(g, me.ID, effectstest.Pw2797WalkerOracle, "Planeswalker — Fix", 2)
		g.WithWriteLock(func() { _ = g.AddCounterForEffect(w, game.CounterLoyalty, -2) })
		effectstest.RunStateChecks(g)
		g.WithWriteLock(func() { _, _ = g.Battlefield.Remove(a) })
		effectstest.RunStateChecks(g)
		if !effectstest.OnBattlefieldNow(g, w) {
			t.Fatal("one of two exempting permanents leaving revoked the other's exemption")
		}
		g.WithWriteLock(func() { _, _ = g.Battlefield.Remove(b) })
		effectstest.RunStateChecks(g)
		if effectstest.OnBattlefieldNow(g, w) {
			t.Error("the last exempting permanent left and the 0-loyalty walker is still here")
		}
		if !me.Graveyard.Contains(w) {
			t.Error("the walker did not go to its owner's graveyard")
		}
	})
	t.Run("lost its abilities", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.PushCatalogPermanent(g, me.ID, "Fixture Lurker", "Creature — Horror", effectstest.Pw2797ExemptOracle, false)
		effectstest.PushCatalogPermanent(g, me.ID, "Fixture Silencer", "Creature — Horror", effectstest.Pw2797SilencerOracle, false)
		w := effectstest.FixtureWalker(g, me.ID, effectstest.Pw2797WalkerOracle, "Planeswalker — Fix", 2)
		g.WithWriteLock(func() { _ = g.AddCounterForEffect(w, game.CounterLoyalty, -2) })
		effectstest.RunStateChecks(g)
		if effectstest.OnBattlefieldNow(g, w) {
			t.Error("an exempting permanent that lost all abilities still exempted")
		}
	})
}

// A walker at 0 loyalty is still a walker: it can use a plus ability (CR 606.6
// asks for counters only to pay a minus), and a minus ability is refused.
func TestZeroLoyaltyWalkerCanOnlyUsePlusAbilities(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushCatalogPermanent(g, me.ID, "Fixture Lurker", "Creature — Horror", effectstest.Pw2797ExemptOracle, false)
	w := effectstest.FixtureWalker(g, me.ID, effectstest.Pw2797WalkerOracle, "Planeswalker — Fix", 3)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(w, game.CounterLoyalty, -3) })
	effectstest.RunStateChecks(g)

	if err := g.ActivateCatalogAbility(me.ID, w, 1, game.ActivateAbilityParams{}); err != game.ErrInsufficientLoyalty {
		t.Fatalf("−3 at 0 loyalty: got %v, want ErrInsufficientLoyalty", err)
	}
	effectstest.B16Activate(t, g, me.ID, w, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LoyaltyCount(g, w); got != 2 {
		t.Errorf("loyalty after +2 from 0 = %d, want 2", got)
	}
	if !effectstest.OnBattlefieldNow(g, w) {
		t.Error("the walker left the battlefield")
	}
}

// Not indestructible: every OTHER state-based action still applies, so a
// planeswalker that is also a creature dies to lethal damage.
func TestZeroLoyaltyExemptionDoesNotStopTheOtherStateBasedActions(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Fixture Lurker", "Creature — Horror", effectstest.Pw2797ExemptOracle, false)
	id := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Walking Walker", TypeLine: "Legendary Creature Planeswalker — Fix",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID, DamageMarked: 2,
	})
	effectstest.RunStateChecks(g)
	if effectstest.OnBattlefieldNow(g, id) {
		t.Error("a creature planeswalker with lethal damage survived because of the loyalty exemption")
	}
}

// --- 3. whenever you put one or more loyalty counters on a planeswalker ----

func watcherCounters2797(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	c, ok := effectstest.BattlefieldCard(g, id)
	if !ok {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return c.Counters[game.CounterPlusOne]
}

// One trigger per placement event, not per counter: a +2 loyalty ability puts
// two counters and triggers once; a −3 removes counters and triggers nothing.
func TestLoyaltyCounterTriggerFiresOncePerPlacementNotPerCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	watcher := effectstest.PushCatalogPermanent(g, me.ID, "Fixture Watcher", "Creature — Elf", effectstest.Pw2797WatcherOracle, false)
	w := effectstest.FixtureWalker(g, me.ID, effectstest.Pw2797WalkerOracle, "Planeswalker — Fix", 5)
	effectstest.SettleTriggers(t, g)
	base := watcherCounters2797(t, g, watcher) // seating the walker was itself a placement

	effectstest.B16Activate(t, g, me.ID, w, 0, game.ActivateAbilityParams{}) // +2: two counters, one placement
	effectstest.SettleTriggers(t, g)
	if got := effectstest.LoyaltyCount(g, w); got != 7 {
		t.Fatalf("loyalty after +2 = %d, want 7", got)
	}
	if got := watcherCounters2797(t, g, watcher) - base; got != 1 {
		t.Errorf("a +2 loyalty ability triggered %d times, want 1 — one placement of two counters", got)
	}

	effectstest.AdvanceToNextSeatsTurn(t, g)
	for i := 0; i < 3 && g.Turn.ActiveSeat != 0; i++ {
		effectstest.AdvanceToNextSeatsTurn(t, g)
	}
	effectstest.ToMain(t, g)
	effectstest.B16Activate(t, g, me.ID, w, 1, game.ActivateAbilityParams{}) // −3: a removal
	effectstest.SettleTriggers(t, g)
	if got := watcherCounters2797(t, g, watcher) - base; got != 1 {
		t.Errorf("a −3 loyalty ability triggered the watcher: %d triggers in all, want it to stay 1", got)
	}
}

// A replacement that doubles the placement (Doubling Season's shape) still
// leaves one event, so one trigger.
func TestLoyaltyCounterTriggerIsOneEventWhenTheCountersAreDoubled(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	watcher := effectstest.PushCatalogPermanent(g, me.ID, "Fixture Watcher", "Creature — Elf", effectstest.Pw2797WatcherOracle, false)
	effectstest.SeedReplacementPermanent(g, effectstest.DoublingSeasonOracle, "Doubling Season", me.ID)
	w := effectstest.FixtureWalker(g, me.ID, effectstest.Pw2797WalkerOracle, "Planeswalker — Fix", 1)
	effectstest.SettleTriggers(t, g)
	before := watcherCounters2797(t, g, watcher)

	g.WithWriteLock(func() { _ = g.AddCounterForEffect(w, game.CounterLoyalty, 1) })
	effectstest.SettleTriggers(t, g)
	if got := effectstest.LoyaltyCount(g, w); got < 3 {
		t.Fatalf("one loyalty counter under Doubling Season on a 1-loyalty walker = %d, want at least 3", got)
	}
	// The watcher's own +1/+1 counter is doubled by the same Season, so one
	// trigger is two counters on it; two triggers would be four.
	if got := watcherCounters2797(t, g, watcher) - before; got != 2 {
		t.Errorf("a doubled placement left %d counters on the watcher, want 2 (one trigger, doubled)", got)
	}
}

// "Put a loyalty counter on each planeswalker you control" is one placement per
// PERMANENT, and a trigger on "a planeswalker" fires once for each.
func TestLoyaltyCounterTriggerFiresForEachPlaneswalker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	watcher := effectstest.PushCatalogPermanent(g, me.ID, "Fixture Watcher", "Creature — Elf", effectstest.Pw2797WatcherOracle, false)
	a := effectstest.FixtureWalker(g, me.ID, effectstest.Pw2797WalkerOracle, "Planeswalker — Fix", 2)
	b := effectstest.FixtureWalker(g, me.ID, effectstest.Pw2797WalkerOracle, "Planeswalker — Fix", 2)
	effectstest.SettleTriggers(t, g)
	before := watcherCounters2797(t, g, watcher)

	g.WithWriteLock(func() {
		_ = g.AddCounterByForEffect(me.ID, a, game.CounterLoyalty, 1)
		_ = g.AddCounterByForEffect(me.ID, b, game.CounterLoyalty, 1)
	})
	effectstest.SettleTriggers(t, g)
	if got := watcherCounters2797(t, g, watcher) - before; got != 2 {
		t.Errorf("counters on two planeswalkers triggered %d times, want 2", got)
	}
}

// "You put": counters an opponent puts on a planeswalker do not trigger it,
// and neither do counters of another kind or counters on a non-planeswalker.
func TestLoyaltyCounterTriggerNeedsYouToPutLoyaltyCountersOnAPlaneswalker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	watcher := effectstest.PushCatalogPermanent(g, me.ID, "Fixture Watcher", "Creature — Elf", effectstest.Pw2797WatcherOracle, false)
	mine := effectstest.FixtureWalker(g, me.ID, effectstest.Pw2797WalkerOracle, "Planeswalker — Fix", 2)
	bear := effectstest.PushCatalogPermanent(g, me.ID, "Bear", "Creature — Bear", "", false)
	effectstest.SettleTriggers(t, g)
	before := watcherCounters2797(t, g, watcher)

	g.WithWriteLock(func() {
		_ = g.AddCounterByForEffect(opp.ID, mine, game.CounterLoyalty, 1) // an opponent puts one on my walker
		_ = g.AddCounterByForEffect(me.ID, mine, "charge", 1)             // not loyalty
		_ = g.AddCounterByForEffect(me.ID, bear, game.CounterLoyalty, 1)  // not a planeswalker
	})
	effectstest.SettleTriggers(t, g)
	if got := watcherCounters2797(t, g, watcher) - before; got != 0 {
		t.Errorf("triggered %d times for placements that are not 'you put loyalty counters on a planeswalker', want 0", got)
	}

	g.WithWriteLock(func() { _ = g.AddCounterByForEffect(me.ID, mine, game.CounterLoyalty, 1) })
	effectstest.SettleTriggers(t, g)
	if got := watcherCounters2797(t, g, watcher) - before; got != 1 {
		t.Errorf("I put a loyalty counter on my walker: triggered %d times, want 1", got)
	}
}
