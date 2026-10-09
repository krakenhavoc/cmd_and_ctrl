package effects

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// planeswalker_statics_2797_test.go — #2797: the planeswalker-matters
// engine pieces the Reality Fracture set needs, proved on fixture
// planeswalkers so the tests do not depend on a card that also needs
// "empower Jace" (#2796).
//
//  1. a loyalty ability granted to a CLASS of planeswalkers (Kiora of Salt
//     and Sand is the real card), counted against CR 606.3;
//  2. the CR 704.5i exemption (Sanctum Lurker's static);
//  3. "whenever you put one or more loyalty counters on a planeswalker"
//     (Inspired Tethermage's trigger).

const (
	// A planeswalker with a +2 and a −3 and nothing else.
	pw2797WalkerOracle = "test-2797-walker"
	// A creature whose whole text is the Lurker's static.
	pw2797ExemptOracle = "test-2797-zero-loyalty-exempt"
	// A creature that silences the exempting fixture, for the
	// "lost its abilities" case.
	pw2797SilencerOracle = "test-2797-silencer"
	// A creature with the Tethermage trigger: a +1/+1 counter on itself
	// for each placement.
	pw2797WatcherOracle = "test-2797-loyalty-watcher"
	// A second grantor of a loyalty ability to planeswalkers you control.
	pw2797GranterOracle = "test-2797-granter"
)

func init() {
	ran := func(g *game.Game, item *game.StackItem) error {
		return g.AddCounterForEffect(item.SourceCardID, "ran", 1)
	}
	Register(Spec{
		OracleID: pw2797WalkerOracle,
		Name:     "Fixture Walker",
		// A fixture, but the catalog's census asks every planeswalker.
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{Label: "+2: Fixture.", Cost: LoyaltyCost(2), Effect: ran},
			{Label: "−3: Fixture.", Cost: LoyaltyCost(-3), Effect: ran},
		},
	})
	Register(Spec{
		OracleID: pw2797GranterOracle,
		Name:     "Fixture Granter",
		Grants: []AbilityGrant{{
			Key: "test-2797-granter/minus-one",
			Activated: []ActivatedAbility{{
				Label: "−1: Fixture.", Cost: LoyaltyCost(-1), Effect: ran,
			}},
			Text: "[−1]: Fixture.",
		}},
		Static: []game.StaticAbility{GrantAbilitiesToYourPlaneswalkers("test-2797-granter/minus-one")},
	})
	Register(Spec{
		OracleID:              pw2797ExemptOracle,
		Name:                  "Fixture Lurker",
		ZeroLoyaltyExemptions: PlaneswalkersSurviveZeroLoyalty(),
	})
	Register(Spec{
		OracleID: pw2797SilencerOracle,
		Name:     "Fixture Silencer",
		Static: []game.StaticAbility{{
			Layer:            game.Layer6Ability,
			RemovesAbilities: true,
			AppliesTo: func(target *game.Card, _ *game.Game, _ *game.Card) bool {
				return target.OracleID == pw2797ExemptOracle
			},
			Apply: func(*game.Characteristic, *game.Card, *game.Game, *game.Card) {},
		}},
	})
	Register(Spec{
		OracleID: pw2797WatcherOracle,
		Name:     "Fixture Watcher",
		Triggered: []game.TriggeredAbility{
			WheneverYouPutLoyaltyCountersOnAPlaneswalker("Fixture Watcher — +1/+1 counter",
				func(g *game.Game, item *game.StackItem) error {
					return g.AddCounterForEffect(item.SourceCardID, game.CounterPlusOne, 1)
				}),
		},
	})
}

// fixtureWalker seats a planeswalker with `loyalty` counters placed the way
// the engine places them (an event per placement), so the loyalty trigger's
// reading of "before" has an arrival to anchor on.
func fixtureWalker(g *game.Game, owner uuid.UUID, oracle, typeLine string, loyalty int) uuid.UUID {
	instance := uuid.New()
	// A distinct name per walker, so two legendary fixtures never meet the
	// legend rule (CR 704.5j) and park the game on a prompt.
	id := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: instance, Name: "Fixture Walker " + instance.String()[:8], TypeLine: typeLine,
		OracleID: oracle, Owner: owner, Controller: owner,
	})
	if loyalty > 0 {
		g.WithWriteLock(func() { _ = g.AddCounterForEffect(id, game.CounterLoyalty, loyalty) })
	}
	return id
}

// grantedRowsOf is the labels of a permanent's GRANTED activated rows.
func grantedRowsOf(t *testing.T, g *game.Game, id uuid.UUID) []string {
	t.Helper()
	abs, origins := game.ActivatedAbilitiesWithOrigins(gaLayered(t, g, id))
	var out []string
	for i, ab := range abs {
		if origins.At(i).Granted() {
			out = append(out, ab.Label)
		}
	}
	return out
}

func onBattlefieldNow(g *game.Game, id uuid.UUID) bool {
	_, ok := battlefieldCard(g, id)
	return ok
}

// --- 1. a loyalty ability granted to planeswalkers you control ------------

// "Planeswalkers you control have '[−8]: …'" reaches every planeswalker its
// controller controls, one that arrives later, and none an opponent controls;
// and it ends the moment the grantor does.
func TestKioraGrantsTheLoyaltyAbilityToEveryPlaneswalkerYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := fixtureWalker(g, me.ID, pw2797WalkerOracle, "Planeswalker — Fix", 4)
	theirs := fixtureWalker(g, opp.ID, pw2797WalkerOracle, "Planeswalker — Fix", 4)
	notAWalker := pushCatalogPermanent(g, me.ID, "Bear", "Creature — Bear", "", false)

	if rows := grantedRowsOf(t, g, a); len(rows) != 0 {
		t.Fatalf("a walker with no grantor on the board already has granted rows: %v", rows)
	}
	kiora := pushCatalogPermanent(g, me.ID, "Kiora of Salt and Sand", "Legendary Creature — Merfolk Noble", kioraOracle2797, false)

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

	late := fixtureWalker(g, me.ID, pw2797WalkerOracle, "Planeswalker — Fix", 3)
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

const kioraOracle2797 = "e70bb7f8-8098-49c5-929c-68f63cd821c5"

// The granted −8 is the walker's: it costs eight of THIS walker's counters
// (CR 606.6), makes an 8/8 blue Leviathan with hexproof for the activator,
// and sends a walker paid down to zero to the graveyard (CR 704.5i).
func TestKioraGrantedMinusEightMakesALeviathan(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	pushCatalogPermanent(g, me.ID, "Kiora of Salt and Sand", "Legendary Creature — Merfolk Noble", kioraOracle2797, false)
	short := fixtureWalker(g, me.ID, pw2797WalkerOracle, "Planeswalker — Fix", 7)
	full := fixtureWalker(g, me.ID, pw2797WalkerOracle, "Planeswalker — Fix", 8)

	idx, ref := grantedLoyaltyRow(t, g, short, "−8")
	if err := g.ActivateCatalogAbility(me.ID, short, idx, game.ActivateAbilityParams{Ref: ref}); err != game.ErrInsufficientLoyalty {
		t.Fatalf("−8 with 7 loyalty: got %v, want ErrInsufficientLoyalty", err)
	}
	if got := loyaltyCount(g, short); got != 7 {
		t.Fatalf("a refused −8 moved the loyalty to %d", got)
	}

	idx, ref = grantedLoyaltyRow(t, g, full, "−8")
	b16Activate(t, g, me.ID, full, idx, game.ActivateAbilityParams{Ref: ref})
	passPriorityAroundTable(t, g)

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
	if !effectiveAbilitiesContain(t, g, leviathan.InstanceID, "hexproof") {
		t.Errorf("Leviathan abilities %v, want hexproof", effectiveAbilities(t, g, leviathan.InstanceID))
	}
	if !strings.Contains(leviathan.TypeLine, "Leviathan") {
		t.Errorf("Leviathan type line = %q", leviathan.TypeLine)
	}
	if onBattlefieldNow(g, full) {
		t.Error("the walker paid all eight loyalty and is still on the battlefield (CR 704.5i)")
	}
}

// CR 606.3 counts per PERMANENT, so the granted row, the walker's own rows
// and a second grantor's row all share one activation a turn, in any order.
func TestClassGrantedLoyaltyAbilitySharesTheWalkersOncePerTurn(t *testing.T) {
	t.Run("own first, then granted", func(t *testing.T) {
		g := newCatalogGame(t)
		toMain(t, g)
		me := g.Seats[g.Turn.ActiveSeat]
		pushCatalogPermanent(g, me.ID, "Kiora of Salt and Sand", "Legendary Creature — Merfolk Noble", kioraOracle2797, false)
		w := fixtureWalker(g, me.ID, pw2797WalkerOracle, "Planeswalker — Fix", 9)
		idx, ref := grantedLoyaltyRow(t, g, w, "−8")

		b16Activate(t, g, me.ID, w, 0, game.ActivateAbilityParams{}) // +2
		passPriorityAroundTable(t, g)
		if err := g.ActivateCatalogAbility(me.ID, w, idx, game.ActivateAbilityParams{Ref: ref}); err != game.ErrLoyaltyAlreadyActivated {
			t.Errorf("the granted −8 after the walker's +2: got %v, want ErrLoyaltyAlreadyActivated", err)
		}
	})
	t.Run("granted first, then own", func(t *testing.T) {
		g := newCatalogGame(t)
		toMain(t, g)
		me := g.Seats[g.Turn.ActiveSeat]
		pushCatalogPermanent(g, me.ID, "Kiora of Salt and Sand", "Legendary Creature — Merfolk Noble", kioraOracle2797, false)
		w := fixtureWalker(g, me.ID, pw2797WalkerOracle, "Planeswalker — Fix", 10)
		idx, ref := grantedLoyaltyRow(t, g, w, "−8")

		b16Activate(t, g, me.ID, w, idx, game.ActivateAbilityParams{Ref: ref})
		passPriorityAroundTable(t, g)
		for _, own := range []int{0, 1} {
			if err := g.ActivateCatalogAbility(me.ID, w, own, game.ActivateAbilityParams{}); err != game.ErrLoyaltyAlreadyActivated {
				t.Errorf("the walker's own row %d after the granted −8: got %v, want ErrLoyaltyAlreadyActivated", own, err)
			}
		}
	})
	t.Run("two grantors, one activation", func(t *testing.T) {
		g := newCatalogGame(t)
		toMain(t, g)
		me := g.Seats[g.Turn.ActiveSeat]
		pushCatalogPermanent(g, me.ID, "Kiora of Salt and Sand", "Legendary Creature — Merfolk Noble", kioraOracle2797, false)
		pushCatalogPermanent(g, me.ID, "Fixture Granter", "Creature — Fixture", pw2797GranterOracle, false)
		w := fixtureWalker(g, me.ID, pw2797WalkerOracle, "Planeswalker — Fix", 20)
		abs, origins := game.ActivatedAbilitiesWithOrigins(gaLayered(t, g, w))
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
		b16Activate(t, g, me.ID, w, idxs[0], game.ActivateAbilityParams{Ref: refs[0]})
		passPriorityAroundTable(t, g)
		if err := g.ActivateCatalogAbility(me.ID, w, idxs[1], game.ActivateAbilityParams{Ref: refs[1]}); err != game.ErrLoyaltyAlreadyActivated {
			t.Errorf("the second grantor's row after the first's: got %v, want ErrLoyaltyAlreadyActivated", err)
		}
	})
}

// --- 2. planeswalkers that aren't put into graveyards for having 0 loyalty --

func runStateChecks(g *game.Game) { g.RunStateChecksForTest() }

// settleTriggers puts triggers a direct test write left pending onto the
// stack (the state check an activation or resolution would have run) and
// resolves them, without moving the step.
func settleTriggers(t *testing.T, g *game.Game) {
	t.Helper()
	runStateChecks(g)
	passPriorityAroundTable(t, g)
}

func TestZeroLoyaltyExemptionKeepsYourPlaneswalkersOnTheBattlefield(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Fixture Lurker", "Creature — Horror", pw2797ExemptOracle, false)
	mine := fixtureWalker(g, me.ID, pw2797WalkerOracle, "Planeswalker — Fix", 1)
	theirs := fixtureWalker(g, opp.ID, pw2797WalkerOracle, "Planeswalker — Fix", 1)

	g.WithWriteLock(func() {
		_ = g.AddCounterForEffect(mine, game.CounterLoyalty, -1)
		_ = g.AddCounterForEffect(theirs, game.CounterLoyalty, -1)
	})
	runStateChecks(g)
	if !onBattlefieldNow(g, mine) {
		t.Error("my 0-loyalty planeswalker was put into the graveyard despite the exemption")
	}
	if loyaltyCount(g, mine) != 0 {
		t.Errorf("loyalty = %d, want it to stay at 0", loyaltyCount(g, mine))
	}
	if onBattlefieldNow(g, theirs) {
		t.Error("an opponent's 0-loyalty planeswalker survived my exemption (it covers planeswalkers I control)")
	}
}

// The exemption is a static of the permanent that prints it: when the last
// exempting permanent leaves, the next state check puts every 0-loyalty walker
// it was covering into its owner's graveyard; two of them compose; and one
// that has lost its abilities exempts nothing.
func TestZeroLoyaltyExemptionIsAStaticOfItsSource(t *testing.T) {
	t.Run("leaving", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[0]
		a := pushCatalogPermanent(g, me.ID, "Fixture Lurker", "Creature — Horror", pw2797ExemptOracle, false)
		b := pushCatalogPermanent(g, me.ID, "Fixture Lurker", "Creature — Horror", pw2797ExemptOracle, false)
		w := fixtureWalker(g, me.ID, pw2797WalkerOracle, "Planeswalker — Fix", 2)
		g.WithWriteLock(func() { _ = g.AddCounterForEffect(w, game.CounterLoyalty, -2) })
		runStateChecks(g)
		g.WithWriteLock(func() { _, _ = g.Battlefield.Remove(a) })
		runStateChecks(g)
		if !onBattlefieldNow(g, w) {
			t.Fatal("one of two exempting permanents leaving revoked the other's exemption")
		}
		g.WithWriteLock(func() { _, _ = g.Battlefield.Remove(b) })
		runStateChecks(g)
		if onBattlefieldNow(g, w) {
			t.Error("the last exempting permanent left and the 0-loyalty walker is still here")
		}
		if !me.Graveyard.Contains(w) {
			t.Error("the walker did not go to its owner's graveyard")
		}
	})
	t.Run("lost its abilities", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[0]
		pushCatalogPermanent(g, me.ID, "Fixture Lurker", "Creature — Horror", pw2797ExemptOracle, false)
		pushCatalogPermanent(g, me.ID, "Fixture Silencer", "Creature — Horror", pw2797SilencerOracle, false)
		w := fixtureWalker(g, me.ID, pw2797WalkerOracle, "Planeswalker — Fix", 2)
		g.WithWriteLock(func() { _ = g.AddCounterForEffect(w, game.CounterLoyalty, -2) })
		runStateChecks(g)
		if onBattlefieldNow(g, w) {
			t.Error("an exempting permanent that lost all abilities still exempted")
		}
	})
}

// A walker at 0 loyalty is still a walker: it can use a plus ability (CR 606.6
// asks for counters only to pay a minus), and a minus ability is refused.
func TestZeroLoyaltyWalkerCanOnlyUsePlusAbilities(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	pushCatalogPermanent(g, me.ID, "Fixture Lurker", "Creature — Horror", pw2797ExemptOracle, false)
	w := fixtureWalker(g, me.ID, pw2797WalkerOracle, "Planeswalker — Fix", 3)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(w, game.CounterLoyalty, -3) })
	runStateChecks(g)

	if err := g.ActivateCatalogAbility(me.ID, w, 1, game.ActivateAbilityParams{}); err != game.ErrInsufficientLoyalty {
		t.Fatalf("−3 at 0 loyalty: got %v, want ErrInsufficientLoyalty", err)
	}
	b16Activate(t, g, me.ID, w, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if got := loyaltyCount(g, w); got != 2 {
		t.Errorf("loyalty after +2 from 0 = %d, want 2", got)
	}
	if !onBattlefieldNow(g, w) {
		t.Error("the walker left the battlefield")
	}
}

// Not indestructible: every OTHER state-based action still applies, so a
// planeswalker that is also a creature dies to lethal damage.
func TestZeroLoyaltyExemptionDoesNotStopTheOtherStateBasedActions(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Fixture Lurker", "Creature — Horror", pw2797ExemptOracle, false)
	id := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Walking Walker", TypeLine: "Legendary Creature Planeswalker — Fix",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID, DamageMarked: 2,
	})
	runStateChecks(g)
	if onBattlefieldNow(g, id) {
		t.Error("a creature planeswalker with lethal damage survived because of the loyalty exemption")
	}
}

// --- 3. whenever you put one or more loyalty counters on a planeswalker ----

func watcherCounters2797(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	c, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return c.Counters[game.CounterPlusOne]
}

// One trigger per placement event, not per counter: a +2 loyalty ability puts
// two counters and triggers once; a −3 removes counters and triggers nothing.
func TestLoyaltyCounterTriggerFiresOncePerPlacementNotPerCounter(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	watcher := pushCatalogPermanent(g, me.ID, "Fixture Watcher", "Creature — Elf", pw2797WatcherOracle, false)
	w := fixtureWalker(g, me.ID, pw2797WalkerOracle, "Planeswalker — Fix", 5)
	settleTriggers(t, g)
	base := watcherCounters2797(t, g, watcher) // seating the walker was itself a placement

	b16Activate(t, g, me.ID, w, 0, game.ActivateAbilityParams{}) // +2: two counters, one placement
	settleTriggers(t, g)
	if got := loyaltyCount(g, w); got != 7 {
		t.Fatalf("loyalty after +2 = %d, want 7", got)
	}
	if got := watcherCounters2797(t, g, watcher) - base; got != 1 {
		t.Errorf("a +2 loyalty ability triggered %d times, want 1 — one placement of two counters", got)
	}

	advanceToNextSeatsTurn(t, g)
	for i := 0; i < 3 && g.Turn.ActiveSeat != 0; i++ {
		advanceToNextSeatsTurn(t, g)
	}
	toMain(t, g)
	b16Activate(t, g, me.ID, w, 1, game.ActivateAbilityParams{}) // −3: a removal
	settleTriggers(t, g)
	if got := watcherCounters2797(t, g, watcher) - base; got != 1 {
		t.Errorf("a −3 loyalty ability triggered the watcher: %d triggers in all, want it to stay 1", got)
	}
}

// A replacement that doubles the placement (Doubling Season's shape) still
// leaves one event, so one trigger.
func TestLoyaltyCounterTriggerIsOneEventWhenTheCountersAreDoubled(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	watcher := pushCatalogPermanent(g, me.ID, "Fixture Watcher", "Creature — Elf", pw2797WatcherOracle, false)
	seedReplacementPermanent(g, doublingSeasonOracle, "Doubling Season", me.ID)
	w := fixtureWalker(g, me.ID, pw2797WalkerOracle, "Planeswalker — Fix", 1)
	settleTriggers(t, g)
	before := watcherCounters2797(t, g, watcher)

	g.WithWriteLock(func() { _ = g.AddCounterForEffect(w, game.CounterLoyalty, 1) })
	settleTriggers(t, g)
	if got := loyaltyCount(g, w); got < 3 {
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	watcher := pushCatalogPermanent(g, me.ID, "Fixture Watcher", "Creature — Elf", pw2797WatcherOracle, false)
	a := fixtureWalker(g, me.ID, pw2797WalkerOracle, "Planeswalker — Fix", 2)
	b := fixtureWalker(g, me.ID, pw2797WalkerOracle, "Planeswalker — Fix", 2)
	settleTriggers(t, g)
	before := watcherCounters2797(t, g, watcher)

	g.WithWriteLock(func() {
		_ = g.AddCounterByForEffect(me.ID, a, game.CounterLoyalty, 1)
		_ = g.AddCounterByForEffect(me.ID, b, game.CounterLoyalty, 1)
	})
	settleTriggers(t, g)
	if got := watcherCounters2797(t, g, watcher) - before; got != 2 {
		t.Errorf("counters on two planeswalkers triggered %d times, want 2", got)
	}
}

// "You put": counters an opponent puts on a planeswalker do not trigger it,
// and neither do counters of another kind or counters on a non-planeswalker.
func TestLoyaltyCounterTriggerNeedsYouToPutLoyaltyCountersOnAPlaneswalker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	watcher := pushCatalogPermanent(g, me.ID, "Fixture Watcher", "Creature — Elf", pw2797WatcherOracle, false)
	mine := fixtureWalker(g, me.ID, pw2797WalkerOracle, "Planeswalker — Fix", 2)
	bear := pushCatalogPermanent(g, me.ID, "Bear", "Creature — Bear", "", false)
	settleTriggers(t, g)
	before := watcherCounters2797(t, g, watcher)

	g.WithWriteLock(func() {
		_ = g.AddCounterByForEffect(opp.ID, mine, game.CounterLoyalty, 1) // an opponent puts one on my walker
		_ = g.AddCounterByForEffect(me.ID, mine, "charge", 1)             // not loyalty
		_ = g.AddCounterByForEffect(me.ID, bear, game.CounterLoyalty, 1)  // not a planeswalker
	})
	settleTriggers(t, g)
	if got := watcherCounters2797(t, g, watcher) - before; got != 0 {
		t.Errorf("triggered %d times for placements that are not 'you put loyalty counters on a planeswalker', want 0", got)
	}

	g.WithWriteLock(func() { _ = g.AddCounterByForEffect(me.ID, mine, game.CounterLoyalty, 1) })
	settleTriggers(t, g)
	if got := watcherCounters2797(t, g, watcher) - before; got != 1 {
		t.Errorf("I put a loyalty counter on my walker: triggered %d times, want 1", got)
	}
}
