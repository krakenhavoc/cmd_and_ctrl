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
	withZoneTriggerCard(t, suspendOracle, []TriggeredAbility{SuspendUpkeepTrigger(), SuspendLastCounterTrigger()})
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

// --- the two triggers (#990) ------------------------------------------

// suspendItemsWaiting counts the stack items — queued or already on
// the stack — carrying `label`. Both halves are counted because a
// harvested trigger spends a moment in PendingTriggers before the
// next priority check drains it onto the stack, and which of the two
// a caller catches it in is not the thing under test.
func suspendItemsWaiting(g *Game, label string) int {
	n := 0
	for _, it := range g.PendingTriggers {
		if it != nil && it.Label == label {
			n++
		}
	}
	for _, it := range g.StackMeta {
		if it != nil && it.Label == label {
			n++
		}
	}
	return n
}

// passUntil passes priority until `done` or the budget runs out.
func passUntil(t *testing.T, g *Game, done func() bool) {
	t.Helper()
	for i := 0; i < 32; i++ {
		if done() {
			return
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority iter %d: %v", i, err)
		}
	}
	t.Fatal("condition never held after 32 priority passes")
}

// CR 702.62b prints TWO triggered abilities and #990 gave the engine
// two. The natural countdown's last tick reaches the cast offer the
// long way round — through the counter event — so the offer is its
// own object on the stack, which is what makes it separately
// counterable, and it is made exactly once.
func TestTheCountdownAndTheCastAreTwoTriggers(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	withSuspendCard(t, 1, "{R}")
	advanceTo(t, g, StepPrecombatMain)
	me.ManaPool.AddMana(ManaToken{Color: "R"})
	id := suspendIt(t, g, me, "Rift Bolt", "Sorcery", "{2}{R}")

	upkeepFor(t, g, 0)
	// Drain the upkeep trigger only.
	passUntil(t, g, func() bool { return timeCounters(g, id) == 0 })

	if pendingChoiceOfKind(g, PendingChoiceMayCast) != nil {
		t.Error("the upkeep trigger made the cast offer itself; the second trigger is what offers it")
	}
	if n := suspendItemsWaiting(g, SuspendFreeCastLabel); n != 1 {
		t.Errorf("last-counter triggers waiting: got %d, want 1 — a separate ability on the stack (CR 702.62b)", n)
	}

	settleStack(t, g)
	if pendingChoiceOfKind(g, PendingChoiceMayCast) == nil {
		t.Fatal("the last-counter trigger made no cast offer")
	}
	offers := 0
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoiceMayCast {
			offers++
		}
	}
	if offers != 1 {
		t.Errorf("cast offers queued: got %d, want 1 — the two triggers must not both offer", offers)
	}
}

// The second trigger fires on ANY removal of the last time counter,
// not only the upkeep one: Jhoira's Timebug, Clockspinning and
// Vampire Hexmage all end a countdown early, and #990 is what lets
// them. No upkeep is involved here at all.
func TestAnExternalRemovalFiresTheLastCounterTrigger(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	withSuspendCard(t, 1, "{R}")
	advanceTo(t, g, StepPrecombatMain)
	me.ManaPool.AddMana(ManaToken{Color: "R"})
	id := suspendIt(t, g, me, "Rift Bolt", "Sorcery", "{2}{R}")

	// Clockspinning, in one line: remove a time counter, in a main
	// phase, from something else's effect.
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(id, CounterTime, -1); err != nil {
			t.Fatalf("AddCounterForEffect: %v", err)
		}
	})
	settleStack(t, g)

	offer := pendingChoiceOfKind(g, PendingChoiceMayCast)
	if offer == nil {
		t.Fatal("removing the last time counter outside an upkeep offered no cast (CR 702.62b)")
	}
	if offer.Chooser != me.ID {
		t.Errorf("offer chooser = %s, want the owner %s", offer.Chooser, me.ID)
	}
	if err := g.ResolveMayCast(offer.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveMayCast: %v", err)
	}
	if perm := grantOn(g, me.ID, id, ZoneExile); perm == nil {
		t.Error("the externally-ended countdown granted no free cast")
	}
}

// A removal that is NOT the last one offers nothing: the trigger reads
// the post-change count, so 3 → 2 is a countdown and 1 → 0 is a cast.
func TestAnEarlierRemovalOffersNothing(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	withSuspendCard(t, 3, "{R}")
	advanceTo(t, g, StepPrecombatMain)
	me.ManaPool.AddMana(ManaToken{Color: "R"})
	id := suspendIt(t, g, me, "Rift Bolt", "Sorcery", "{2}{R}")

	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(id, CounterTime, -1); err != nil {
			t.Fatalf("AddCounterForEffect: %v", err)
		}
	})
	settleStack(t, g)
	if n := timeCounters(g, id); n != 2 {
		t.Fatalf("time counters: got %d, want 2", n)
	}
	if pendingChoiceOfKind(g, PendingChoiceMayCast) != nil {
		t.Error("a removal that left counters behind offered the cast")
	}
}

// --- the haste (#990) --------------------------------------------------

// castTheSuspendedCreature runs the whole keyword for a creature and
// returns it on the battlefield: suspend it, tick the counter off,
// take the offer, cast it, resolve it.
func castTheSuspendedCreature(t *testing.T, g *Game, me *Player) uuid.UUID {
	t.Helper()
	advanceTo(t, g, StepPrecombatMain)
	me.ManaPool.AddMana(ManaToken{Color: "R"})
	id := suspendIt(t, g, me, "Test Goblin", "Creature — Goblin", "{2}{R}")
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
	return id
}

// hasteOn reads the keyword off the live battlefield object.
func hasteOn(t *testing.T, g *Game, id uuid.UUID) bool {
	t.Helper()
	c, ok := battlefieldCardByID(g, id)
	if !ok {
		t.Fatalf("card %s is not on the battlefield", id)
	}
	return HasKeyword(&c, "haste")
}

// CR 702.62e: "it gains haste until you lose control of it" — not
// until end of turn, which is what #659 shipped and #990 retired. The
// cleanup step sweeps every UntilEndOfTurn effect there is, and this
// one is still standing on the far side of it.
func TestSuspendHasteOutlivesTheTurn(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	withSuspendCard(t, 1, "{R}")
	id := castTheSuspendedCreature(t, g, me)

	if !hasteOn(t, g, id) {
		t.Fatal("the suspended creature has no haste on the turn it arrived")
	}
	advancePastScopedCleanup(t, g)
	if !hasteOn(t, g, id) {
		t.Error("the haste was swept at the cleanup step; CR 702.62e ends it on a control change, not on a turn")
	}
}

// And it ends the moment control changes — permanently, because a
// ForAsLongAs duration that has gone false is dropped rather than
// re-evaluated back to life. Getting the creature back does not get
// the haste back, which is what "until that player loses control of
// it" says.
func TestSuspendHasteEndsWhenControlIsLost(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	withSuspendCard(t, 1, "{R}")
	id := castTheSuspendedCreature(t, g, me)
	if !hasteOn(t, g, id) {
		t.Fatal("the suspended creature has no haste to lose")
	}

	g.WithWriteLock(func() {
		if !g.GainControlForEffect(uuid.New(), id, opp.ID, IndefiniteDuration(), "test — steal it") {
			t.Fatal("GainControlForEffect refused the creature")
		}
	})
	if controllerOfCard(t, g, id) != opp.ID {
		t.Fatal("the theft did not take")
	}
	if hasteOn(t, g, id) {
		t.Error("the creature kept its suspend haste under a new controller (CR 702.62e)")
	}

	// Give it back. The grant is gone for good.
	g.WithWriteLock(func() {
		g.GainControlForEffect(uuid.New(), id, me.ID, IndefiniteDuration(), "test — give it back")
	})
	if controllerOfCard(t, g, id) != me.ID {
		t.Fatal("the creature did not come back")
	}
	if hasteOn(t, g, id) {
		t.Error("the suspend haste came back with the creature; CR 702.62e ended it the first time control left")
	}
	for _, s := range g.ScopedStatics {
		if s.Label == "Suspend — haste (CR 702.62e)" {
			t.Error("the ended haste grant is still in the registry")
		}
	}
}
