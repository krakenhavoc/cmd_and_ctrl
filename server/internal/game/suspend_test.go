package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// suspend_test.go — suspend (CR 702.62), #659.
//
// The verb's timing table is pinned in special_action_test.go and the
// permission model in cast_permission_test.go. What is tested here is
// suspend's own machinery: the counters as the suspended state, the
// exile-zone countdown, the free cast on the last tick and the rules
// that ride it — X locked at 0, haste, and the Path-to-Exile'd copy
// that must never be castable.

const suspendOracle = "test-suspend-oracle"

// withSuspendCard declares suspend N—cost on one oracle key, and
// wires the countdown trigger the way buildDef does in production:
// from the declaration, not from the card file.
func withSuspendCard(t *testing.T, n int, cost string) {
	t.Helper()
	withCatalogSpecialActions(t, func(id string) []SpecialAction {
		if id != suspendOracle {
			return nil
		}
		return []SpecialAction{{
			Kind: SpecialActionSuspend, Cost: cost, Counters: n, Label: SuspendLabel(n, cost),
		}}
	})
	withZoneTriggerCard(t, suspendOracle, []TriggeredAbility{SuspendUpkeepTrigger()})
}

// suspendIt seeds a suspendable card in the seat's hand and takes the
// special action. `typeLine` decides the timing window, so the caller
// has to have the cursor somewhere legal for it.
func suspendIt(t *testing.T, g *Game, p *Player, name, typeLine, manaCost string) uuid.UUID {
	t.Helper()
	card := seedHandCard(p, name, suspendOracle, typeLine, manaCost)
	if err := g.PerformSpecialAction(p.ID, card.InstanceID, SpecialActionSuspend, SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("suspend: %v", err)
	}
	return card.InstanceID
}

// setHandOrExilePT stamps power and toughness on a card wherever it
// currently sits. A 0/0 creature dies to CR 704.5f the moment it
// reaches the battlefield, which has nothing to do with suspend.
func setHandOrExilePT(g *Game, p *Player, id uuid.UUID, power, toughness int) {
	for _, z := range []*Zone{p.Hand, g.Exile} {
		if z == nil {
			continue
		}
		for i := range z.Cards {
			if z.Cards[i].InstanceID == id {
				z.Cards[i].Power, z.Cards[i].Toughness = power, toughness
				return
			}
		}
	}
}

// timeCounters reads the time counters on a card in exile, or -1 when
// the card is not there.
func timeCounters(g *Game, id uuid.UUID) int {
	c := exiledCardByIDLocked(g, id)
	if c == nil {
		return -1
	}
	return c.Counters[CounterTime]
}

// upkeepFor walks the cursor to the named seat's upkeep and drains
// whatever the step queued onto the stack.
func upkeepFor(t *testing.T, g *Game, seat int) {
	t.Helper()
	// Always move at least one step, so a caller standing IN the
	// upkeep it just resolved walks on to the NEXT one rather than
	// returning immediately.
	for i := 0; i < 60; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
		if g.Turn.Step == StepUpkeep && g.Turn.ActiveSeat == seat {
			return
		}
	}
	t.Fatalf("never reached seat %d's upkeep", seat)
}

// --- the special action ----------------------------------------------

// CR 702.62a: pay the suspend cost, exile the card with N time
// counters. Face UP — suspend hides nothing — and out of the hand.
func TestSuspendExilesWithTimeCounters(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	withSuspendCard(t, 3, "{R}")
	advanceTo(t, g, StepPrecombatMain)
	me.ManaPool.AddMana(ManaToken{Color: "R"})

	id := suspendIt(t, g, me, "Rift Bolt", "Sorcery", "{2}{R}")

	if len(me.ManaPool) != 0 {
		t.Errorf("mana pool: got %d tokens left, want 0 — the suspend cost was not paid", len(me.ManaPool))
	}
	if handCard(me, id) != nil {
		t.Error("the suspended card is still in hand")
	}
	c := exiledCardByIDLocked(g, id)
	if c == nil {
		t.Fatal("the suspended card is not in exile")
	}
	if c.FaceDown {
		t.Error("suspend exiled the card face down; CR 702.62b exiles it face up")
	}
	if n := c.Counters[CounterTime]; n != 3 {
		t.Errorf("time counters: got %d, want 3", n)
	}
	if !CardIsSuspended(*c) {
		t.Error("a card with time counters in exile is not reported suspended")
	}
	if g.Stack != nil && len(g.Stack.Cards) > 0 {
		t.Error("suspending put something on the stack")
	}
}

// --- the countdown ----------------------------------------------------

// CR 702.62b: at the beginning of YOUR upkeep, remove a time counter.
// From exile, through #925's zone dimension — and only on your own
// upkeep, which is what stops a four-player table ticking a suspended
// card four times a round.
func TestTimeCountersTickOnYourUpkeepOnly(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0]
	withSuspendCard(t, 3, "{R}")
	advanceTo(t, g, StepPrecombatMain)
	me.ManaPool.AddMana(ManaToken{Color: "R"})
	id := suspendIt(t, g, me, "Rift Bolt", "Sorcery", "{2}{R}")

	// Three other seats take a turn each: no ticks.
	for seat := 1; seat < 4; seat++ {
		upkeepFor(t, g, seat)
		settleStack(t, g)
		if n := timeCounters(g, id); n != 3 {
			t.Fatalf("after seat %d's upkeep: %d time counters, want 3", seat, n)
		}
	}
	upkeepFor(t, g, 0)
	settleStack(t, g)
	if n := timeCounters(g, id); n != 2 {
		t.Errorf("after the owner's upkeep: %d time counters, want 2", n)
	}
}

// A copy of the same card that reached exile some OTHER way is not
// suspended: no counters, so no trigger, no countdown, and — the bug
// the retired card-level `CastableZones: exile` shape would have
// shipped — no cast.
func TestAPathExiledCopyNeverTicksAndIsNotCastable(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	withSuspendCard(t, 1, "{R}")
	advanceTo(t, g, StepPrecombatMain)

	pathed := NewCard("Rift Bolt", me.ID)
	pathed.OracleID = suspendOracle
	pathed.TypeLine = "Sorcery"
	pathed.ManaCost = "{2}{R}"
	g.Exile.PushTop(pathed)

	upkeepFor(t, g, 0)
	settleStack(t, g)

	if n := timeCounters(g, pathed.InstanceID); n != 0 {
		t.Errorf("a Path to Exile'd copy has %d time counters, want none", n)
	}
	if perm := grantOn(g, me.ID, pathed.InstanceID, ZoneExile); perm != nil {
		t.Error("a Path to Exile'd copy has a cast permission")
	}
	me.ManaPool.AddMana(ManaToken{Color: "C"}, ManaToken{Color: "C"}, ManaToken{Color: "R"})
	err := g.CastSpell(me.ID, pathed.InstanceID, CastSpellParams{Strict: true, FromZone: "exile"})
	if !errors.Is(err, ErrNoPlayPermission) {
		t.Fatalf("cast of a Path to Exile'd copy: got %v, want ErrNoPlayPermission", err)
	}
}

// --- the free cast ----------------------------------------------------

// CR 702.62b/c: when the LAST counter comes off, and only then, the
// owner is offered the cast. Taking it stamps a free, flash-timed
// permission over that one object.
func TestTheLastTickOffersTheFreeCast(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	withSuspendCard(t, 2, "{R}")
	advanceTo(t, g, StepPrecombatMain)
	me.ManaPool.AddMana(ManaToken{Color: "R"})
	id := suspendIt(t, g, me, "Rift Bolt", "Sorcery", "{2}{R}")

	// First tick: 2 → 1, no offer.
	upkeepFor(t, g, 0)
	settleStack(t, g)
	if n := timeCounters(g, id); n != 1 {
		t.Fatalf("after one tick: %d time counters, want 1", n)
	}
	if pendingChoiceOfKind(g, PendingChoiceMayCast) != nil {
		t.Fatal("the cast was offered before the last counter came off")
	}

	// Second tick: 1 → 0, and the offer.
	upkeepFor(t, g, 0)
	settleStack(t, g)
	if n := timeCounters(g, id); n != 0 {
		t.Fatalf("after two ticks: %d time counters, want 0", n)
	}
	offer := pendingChoiceOfKind(g, PendingChoiceMayCast)
	if offer == nil {
		t.Fatal("the last tick offered no cast")
	}
	if offer.Chooser != me.ID {
		t.Errorf("offer chooser = %s, want the owner %s", offer.Chooser, me.ID)
	}
	if err := g.ResolveMayCast(offer.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveMayCast: %v", err)
	}

	perm := grantOn(g, me.ID, id, ZoneExile)
	if perm == nil {
		t.Fatal("accepting the offer granted no cast permission")
	}
	if perm.Cost != "{0}" {
		t.Errorf("granted price: got %q, want {0} — without paying its mana cost", perm.Cost)
	}
	if perm.Timing != TimingFlash {
		t.Errorf("granted timing: got %q, want flash — the trigger resolves in an upkeep", perm.Timing)
	}
	// A sorcery, in an upkeep, for nothing.
	if err := g.CastSpell(me.ID, id, CastSpellParams{Strict: true, FromZone: "exile"}); err != nil {
		t.Fatalf("the free cast: %v", err)
	}
	if !g.Stack.Contains(id) {
		t.Error("the free cast did not reach the stack")
	}
}

// Declining leaves the card stranded in exile with no counters and no
// permission, which is the printed outcome.
func TestDecliningTheFreeCastStrandsTheCard(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	withSuspendCard(t, 1, "{R}")
	advanceTo(t, g, StepPrecombatMain)
	me.ManaPool.AddMana(ManaToken{Color: "R"})
	id := suspendIt(t, g, me, "Rift Bolt", "Sorcery", "{2}{R}")

	upkeepFor(t, g, 0)
	settleStack(t, g)
	offer := pendingChoiceOfKind(g, PendingChoiceMayCast)
	if offer == nil {
		t.Fatal("no cast offer on the last tick")
	}
	if err := g.ResolveMayCast(offer.ID, me.ID, false); err != nil {
		t.Fatalf("ResolveMayCast(decline): %v", err)
	}
	if exiledCardByIDLocked(g, id) == nil {
		t.Error("a declined suspend cast left exile anyway")
	}
	if perm := grantOn(g, me.ID, id, ZoneExile); perm != nil {
		t.Error("a declined offer still granted a cast permission")
	}
}

// CR 107.3b, through the shared rule #873 wrote: a spell cast without
// paying its mana cost has exactly one legal X, and it is 0. The
// suspend grant needs no rule of its own for it — CastCostFor answers
// from the "{0}" price.
func TestTheFreeCastLocksXAtZero(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	withSuspendCard(t, 1, "{R}")
	advanceTo(t, g, StepPrecombatMain)
	me.ManaPool.AddMana(ManaToken{Color: "R"})
	id := suspendIt(t, g, me, "Test Fireball", "Sorcery", "{X}{R}")

	upkeepFor(t, g, 0)
	settleStack(t, g)
	offer := pendingChoiceOfKind(g, PendingChoiceMayCast)
	if offer == nil {
		t.Fatal("no cast offer on the last tick")
	}
	if err := g.ResolveMayCast(offer.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveMayCast: %v", err)
	}
	if err := g.CastSpell(me.ID, id, CastSpellParams{Strict: true, FromZone: "exile", XValue: 5}); !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("X=5 on a free cast: got %v, want ErrInvalidParam", err)
	}
	if err := g.CastSpell(me.ID, id, CastSpellParams{Strict: true, FromZone: "exile", XValue: 0}); err != nil {
		t.Fatalf("X=0 on a free cast: %v", err)
	}
}

// CR 702.62e: a creature cast this way has haste. The grant carries
// the fact and CastSpell registers the layer-6 keyword against the
// one object it opened.
func TestASuspendedCreatureHasHaste(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	withSuspendCard(t, 1, "{R}")
	advanceTo(t, g, StepPrecombatMain)
	me.ManaPool.AddMana(ManaToken{Color: "R"})
	id := suspendIt(t, g, me, "Test Goblin", "Creature — Goblin", "{2}{R}")
	// A 0/0 would die to CR 704.5f the moment it landed.
	setHandOrExilePT(g, me, id, 2, 2)

	upkeepFor(t, g, 0)
	settleStack(t, g)
	offer := pendingChoiceOfKind(g, PendingChoiceMayCast)
	if offer == nil {
		t.Fatal("no cast offer on the last tick")
	}
	if err := g.ResolveMayCast(offer.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveMayCast: %v", err)
	}
	if err := g.CastSpell(me.ID, id, CastSpellParams{Strict: true, FromZone: "exile"}); err != nil {
		t.Fatalf("the free cast: %v", err)
	}
	settleStack(t, g)

	c, ok := battlefieldCardByID(g, id)
	if !ok {
		t.Fatal("the suspended creature is not on the battlefield")
	}
	if !HasKeyword(&c, "haste") {
		t.Error("a creature cast from a suspended card has no haste (CR 702.62e)")
	}
	if HasSummoningSickness(&c) {
		t.Error("the suspended creature is summoning sick despite its haste")
	}
	// CR 400.7: the permanent that arrived is a new object and carries
	// none of the time counters the exiled card had.
	if n := c.Counters[CounterTime]; n != 0 {
		t.Errorf("the permanent entered with %d time counters, want none", n)
	}
}

// A hard-cast copy of the same card gets no haste: the haste belongs
// to the PERMISSION, not to the card.
func TestAHardCastCopyGetsNoHaste(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	withSuspendCard(t, 1, "{R}")
	advanceTo(t, g, StepPrecombatMain)
	card := seedHandCard(me, "Test Goblin", suspendOracle, "Creature — Goblin", "{2}{R}")
	setHandOrExilePT(g, me, card.InstanceID, 2, 2)
	me.ManaPool.AddMana(ManaToken{Color: "C"}, ManaToken{Color: "C"}, ManaToken{Color: "R"})

	if err := g.CastSpell(me.ID, card.InstanceID, CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("hard cast: %v", err)
	}
	settleStack(t, g)
	c, ok := battlefieldCardByID(g, card.InstanceID)
	if !ok {
		t.Fatal("the hard-cast creature is not on the battlefield")
	}
	if HasKeyword(&c, "haste") {
		t.Error("a hard-cast copy has haste")
	}
}

// Suspend is a built kind.
func TestSuspendIsABuiltKind(t *testing.T) {
	if !SpecialActionKindBuilt(SpecialActionSuspend) {
		t.Error("suspend is declared but has no performer")
	}
}

// The whole of suspend's state is carried: the counters are ordinary
// card state and the permission is ADR 0066's. An undo of the special
// action puts the card back in hand with the cost refunded.
func TestSuspendSurvivesAnUndo(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	withSuspendCard(t, 3, "{R}")
	advanceTo(t, g, StepPrecombatMain)
	card := seedHandCard(me, "Rift Bolt", suspendOracle, "Sorcery", "{2}{R}")
	me.ManaPool.AddMana(ManaToken{Color: "R"})

	before := g.Clone()
	if err := g.PerformSpecialAction(me.ID, card.InstanceID, SpecialActionSuspend, SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("suspend: %v", err)
	}
	g.WithWriteLock(func() { g.RestoreFrom(before) })

	me = g.Seats[0]
	if handCard(me, card.InstanceID) == nil {
		t.Error("undo did not return the suspended card to hand")
	}
	if exiledCardByIDLocked(g, card.InstanceID) != nil {
		t.Error("undo left the card in exile")
	}
	if len(me.ManaPool) != 1 {
		t.Errorf("mana pool after undo: got %d tokens, want 1", len(me.ManaPool))
	}
}

// The time counters and the free-cast permission both survive a
// snapshot round trip — the first is ordinary card state, the second
// is ADR 0066's carried store.
func TestSuspendedStateSurvivesASnapshotRoundTrip(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	withSuspendCard(t, 2, "{R}")
	advanceTo(t, g, StepPrecombatMain)
	me.ManaPool.AddMana(ManaToken{Color: "R"})
	id := suspendIt(t, g, me, "Rift Bolt", "Sorcery", "{2}{R}")

	_, restored := roundTrip(t, g)
	if n := timeCounters(restored, id); n != 2 {
		t.Errorf("time counters after the round trip: got %d, want 2", n)
	}
}
