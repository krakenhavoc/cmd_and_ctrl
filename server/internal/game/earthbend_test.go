package game

import (
	"testing"

	"github.com/google/uuid"
)

// earthbend_test.go — #1178, the engine half of the AVATAR keyword
// action:
//
//	Earthbend N. (Target land you control becomes a 0/0 creature with
//	haste that's still a land. Put N +1/+1 counters on it. When it
//	dies or is exiled, return it to the battlefield tapped.)
//
// Four parts in one verb, so the tests are about the seams between
// them: the animation and the haste share a duration, the counters go
// through the CR 614 placement window, the delayed return is keyed on
// the OBJECT, and the whole thing rides the keyword-action window so
// its count is replaceable.
//
// The card half (Earthbending Lesson, Rockalanche, The Legend of
// Kyoshi's chapter II) is in cards/effects/earthbend_cards_test.go.

// earthbendTestLand parks an untapped Forest on the battlefield under
// `owner`, long enough ago that summoning sickness is not the reason
// anything below can or cannot attack.
func earthbendTestLand(t *testing.T, g *Game, owner *Player, name string) uuid.UUID {
	t.Helper()
	id := pushBattlefieldForTest(g, owner.ID, name, "Basic Land — Forest", "")
	c := findCard(g, id)
	if c == nil {
		t.Fatalf("setup: %s is not on the battlefield", name)
	}
	c.EnteredBattlefieldAt = timeNowUnixNano()
	c.SummonedThisTurn = false
	return id
}

// earthbend takes the keyword action under the write lock.
func earthbend(t *testing.T, g *Game, actor *Player, land uuid.UUID, n int) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.EarthbendForEffect(actor.ID, uuid.Nil, land, n); err != nil {
			t.Fatalf("EarthbendForEffect(%d): %v", n, err)
		}
	})
}

// earthbentView is the land's post-layer characteristics.
func earthbentView(t *testing.T, g *Game, land uuid.UUID) *Card {
	t.Helper()
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	c := findCard(g, land)
	if c == nil {
		t.Fatalf("the land %s is not on the battlefield", land)
	}
	return c
}

// --- part 1-3: the animation, the haste and the counters -------------

// TestEarthbendMakesALandCreatureWithHasteAndCounters is the whole
// reminder text's first two sentences at once: still a land, also a
// creature, base 0/0 with N +1/+1 counters on top (so a 4/4), and
// hasty.
func TestEarthbendMakesALandCreatureWithHasteAndCounters(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	land := earthbendTestLand(t, g, me, "Forest")

	earthbend(t, g, me, land, 4)

	c := earthbentView(t, g, land)
	if !c.IsLand() {
		t.Error(`"that's still a land": the Creature add must not remove the Land type (CR 613.1d)`)
	}
	if !c.IsCreature() {
		t.Error("the land did not become a creature (layer 4)")
	}
	if got := counterOn(g, land, CounterPlusOne); got != 4 {
		t.Errorf("+1/+1 counters = %d, want 4", got)
	}
	if p, tough := c.CurrentPower(), c.CurrentToughness(); p != 4 || tough != 4 {
		t.Errorf("power/toughness = %d/%d, want 4/4 — base 0/0 at layer 7b with four counters at 7d",
			p, tough)
	}
	if !HasKeyword(c, "haste") {
		t.Error("the animated land has no haste")
	}
}

// TestEarthbendZeroStillAnimatesAndReturnsIt is Rockalanche with no
// Forests, and it is the reason `KeywordAction.actsAtZeroCount`
// exists: "earthbend 0" is not a no-op. The land becomes a 0/0 with
// no counters, the toughness state-based action kills it (CR 704.5f —
// the layer-7b set is what makes its toughness KNOWN, #690), and the
// delayed return hands it back tapped.
func TestEarthbendZeroStillAnimatesAndReturnsIt(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	land := earthbendTestLand(t, g, me, "Forest")

	earthbend(t, g, me, land, 0)

	c := earthbentView(t, g, land)
	if !c.IsCreature() || c.CurrentToughness() != 0 {
		t.Fatalf("earthbend 0 did not make a 0/0 creature: creature=%v toughness=%d",
			c.IsCreature(), c.CurrentToughness())
	}
	runSBAsForTest(g)
	settleStack(t, g)

	back := findCard(g, land)
	if back == nil {
		t.Fatal("the 0/0 land died and never came back")
	}
	if !back.Tapped {
		t.Error("it returned untapped; the keyword says tapped")
	}
	if back.IsCreature() {
		t.Error("the returned land is still animated — CR 400.7 makes it a new object")
	}
}

// TestEarthbendOverwritesAPrintedBody is the half of layer 7b that a
// plain Forest cannot show: a land that ALREADY has a printed power
// and toughness (Dryad Arbor is a 1/1 Land Creature) becomes 0/0 and
// then takes the counters, so an earthbend 2 is a 2/2 and not a 3/3.
//
// "Becomes a 0/0 creature" SETS the body (CR 613.4b); it does not add
// to whatever was there.
func TestEarthbendOverwritesAPrintedBody(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	arbor := pushBattlefieldForTest(g, me.ID, "Dryad Arbor", "Land Creature — Forest Dryad", "")
	c := findCard(g, arbor)
	c.Power, c.Toughness = 1, 1
	c.PrintedPTKnown = true
	c.EnteredBattlefieldAt = timeNowUnixNano()
	c.SummonedThisTurn = false

	earthbend(t, g, me, arbor, 2)

	view := earthbentView(t, g, arbor)
	if p, tough := view.CurrentPower(), view.CurrentToughness(); p != 2 || tough != 2 {
		t.Errorf("power/toughness = %d/%d, want 2/2 — the printed 1/1 is REPLACED by 0/0 at "+
			"layer 7b before the two counters apply at 7d", p, tough)
	}
}

// TestEarthbentLandCanAttackTheTurnItWasPlayed is what the haste is
// actually for, and the setup is the half that makes the claim real.
//
// CR 302.6 runs on continuous CONTROL since the turn began, not on
// when the permanent became a creature, so a land that has been
// around since last turn can attack the moment it is animated whether
// or not it has haste (`HasSummoningSickness`'s own "same for a
// manland animated the turn it entered" note). The land that NEEDS
// the haste is the one played THIS turn — earthbend it and swing, in
// one turn, off a land drop.
func TestEarthbentLandCanAttackTheTurnItWasPlayed(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0]
	land := earthbendTestLand(t, g, me, "Forest")
	findCard(g, land).SummonedThisTurn = true

	advanceIntoStep(t, g, StepDeclareAttackers)
	earthbend(t, g, me, land, 3)

	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if HasSummoningSickness(findCard(g, land)) {
		t.Fatal("the earthbent land is summoning-sick; the haste grant did not land")
	}
	if err := g.DeclareAttacker(land, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v — an earthbent land has haste and may attack", err)
	}
	if c := findCard(g, land); c == nil || c.AttackingTarget != g.Seats[1].ID {
		t.Error("the attack was not staged on the animated land")
	}
}

// TestAnEarthbentLandPlayedThisTurnIsSickWithoutTheHaste is the
// control for the test above: strip the haste and the same board
// refuses the attack. Without it "it can attack" would pass on a land
// that was never summoning-sick to begin with.
func TestAnEarthbentLandPlayedThisTurnIsSickWithoutTheHaste(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0]
	land := earthbendTestLand(t, g, me, "Forest")
	findCard(g, land).SummonedThisTurn = true

	advanceIntoStep(t, g, StepDeclareAttackers)
	// The animation without its third static: a 0/0 Land Creature with
	// three counters and no haste.
	g.WithWriteLock(func() {
		c, ok := g.battlefieldCardLocked(land)
		if !ok {
			t.Fatal("setup: the land is not on the battlefield")
		}
		g.animateLandWithoutHasteForTest(land, c.EnteredBattlefieldAt)
		if err := g.AddCounterByForEffect(me.ID, land, CounterPlusOne, 3); err != nil {
			t.Fatalf("AddCounterByForEffect: %v", err)
		}
		g.RecomputeLayersIfStaleLocked()
	})

	if !HasSummoningSickness(findCard(g, land)) {
		t.Fatal("setup: a land animated the turn it was played must be summoning-sick without haste")
	}
	if err := g.DeclareAttacker(land, g.Seats[1].ID); err == nil {
		t.Error("a hasteless animated land attacked the turn it was played (CR 302.6)")
	}
}

// animateLandWithoutHasteForTest is the layer-4 and layer-7b halves of
// an earthbend with the layer-6 haste left out — the control board for
// the test above, built from the same predicate the real thing uses so
// the only difference is the grant.
func (g *Game) animateLandWithoutHasteForTest(land uuid.UUID, stamp int64) {
	applies := func(t *Card, _ *Game, _ *Card) bool {
		return t.InstanceID == land && t.EnteredBattlefieldAt == stamp
	}
	d := g.PinnedTo(IndefiniteDuration(), land)
	ts := timeNowUnixNano()
	g.registerScopedStaticLocked(StaticAbility{
		Layer:     Layer4Type,
		AppliesTo: applies,
		Apply: func(ch *Characteristic, _ *Card, _ *Game, _ *Card) {
			if !typeListHas(ch.Types, "Creature") {
				ch.Types = append(ch.Types, "Creature")
			}
		},
	}, uuid.Nil, "probe — animate without haste", d, ts)
	g.registerScopedStaticLocked(StaticAbility{
		Layer:     Layer7PT,
		SubLayer:  SubLayer7B_Set,
		AppliesTo: applies,
		Apply: func(ch *Characteristic, _ *Card, _ *Game, _ *Card) {
			ch.Power, ch.Toughness = 0, 0
		},
	}, uuid.Nil, "probe — animate without haste (base P/T)", d, ts)
}

// TestEarthbendCountersGoThroughThePlacementWindow: the counters are
// PLACED on a permanent already on the battlefield, so the CR 614
// counter window opens and Hardened Scales sees them.
//
// The entry pipeline is deliberately not involved — the land is not
// entering — which is what makes this the placing path and not the
// "enters with N counters" one.
func TestEarthbendCountersGoThroughThePlacementWindow(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	land := earthbendTestLand(t, g, me, "Forest")

	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(hardenedScalesForTest())
	})
	earthbend(t, g, me, land, 4)

	if got := counterOn(g, land, CounterPlusOne); got != 5 {
		t.Errorf("+1/+1 counters = %d, want 5 — Hardened Scales adds one to the placement", got)
	}
	if c := earthbentView(t, g, land); c.CurrentPower() != 5 {
		t.Errorf("power = %d, want 5", c.CurrentPower())
	}
}

// hardenedScalesForTest is "if one or more +1/+1 counters would be
// put on a creature you control, that many plus one are put instead",
// as a test injection with no source card.
func hardenedScalesForTest() ReplacementEffect {
	return ReplacementEffect{
		Watches: []EventKind{EventCounterPlaced},
		AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
			return ev.Kind == RepEventCounter && ev.CounterName == CounterPlusOne && ev.CounterDelta > 0
		},
		Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
			ev.CounterDelta++
			return nil
		},
		Label: "Hardened Scales (probe)",
	}
}

// --- the duration ----------------------------------------------------

// TestTheHasteAndTheAnimationEndTogether: all three continuous effects
// share one duration pinned to the object, so there is no window in
// which the land is a creature without haste or hasty without being a
// creature. Once it has left and come back it is neither, and the
// registry has dropped all three rather than carrying dead entries.
func TestTheHasteAndTheAnimationEndTogether(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	land := earthbendTestLand(t, g, me, "Forest")

	earthbend(t, g, me, land, 2)
	if n := len(g.ScopedStatics); n != 3 {
		t.Fatalf("earthbend registered %d continuous effects, want 3 (type, base P/T, haste)", n)
	}
	first := g.ScopedStatics[0]
	for _, s := range g.ScopedStatics[1:] {
		if s.Timestamp != first.Timestamp {
			t.Errorf("the three halves of one printed sentence have different CR 613.7 timestamps")
		}
		if s.Duration != first.Duration {
			t.Errorf("the three halves have different durations: %+v vs %+v", s.Duration, first.Duration)
		}
	}
	if first.Duration.Kind != Indefinite {
		t.Errorf("duration kind = %v, want Indefinite — earthbend states none (CR 611.2a)", first.Duration.Kind)
	}
	if first.Duration.Pinned != land {
		t.Errorf("the duration is not pinned to the land; a dead entry would outlive the object (CR 400.7)")
	}

	// Kill it. The return is a new object, so nothing applies any more
	// — and the pin drops all three registry entries at the next sweep.
	killEarthbentLand(t, g, land)

	c := earthbentView(t, g, land)
	if c.IsCreature() {
		t.Error("the returned land is still a creature")
	}
	if HasKeyword(c, "haste") {
		t.Error("the returned land still has haste; the grant outlived the animation")
	}
	g.WithWriteLock(func() { g.ClearExpiredScopedStaticsLocked() })
	if n := len(g.ScopedStatics); n != 0 {
		t.Errorf("%d continuous effects survived the object they were pinned to", n)
	}
}

// killEarthbentLand destroys the animated land and settles whatever
// the death put on the stack, leaving the return resolved.
func killEarthbentLand(t *testing.T, g *Game, land uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(land); err != nil {
			t.Fatalf("DestroyPermanentForEffect: %v", err)
		}
	})
	settleStack(t, g)
}

// --- part 4: the delayed return --------------------------------------

// TestDyingReturnsTheLandTappedAsANewObject is the fourth sentence.
// The land really DIES — the trigger is not a replacement, so a
// graveyard is where it goes and dies-watchers see it — and then comes
// back tapped, stripped of the animation and of its counters
// (CR 400.7).
func TestDyingReturnsTheLandTappedAsANewObject(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	land := earthbendTestLand(t, g, me, "Forest")
	before := findCard(g, land).EnteredBattlefieldAt

	earthbend(t, g, me, land, 3)
	killEarthbentLand(t, g, land)

	if me.Graveyard.Contains(land) {
		t.Error("the land is still in the graveyard; the delayed return did not fire")
	}
	c := findCard(g, land)
	if c == nil {
		t.Fatal("the land did not come back")
	}
	if !c.Tapped {
		t.Error("it came back untapped")
	}
	if c.EnteredBattlefieldAt == before {
		t.Error("CR 400.7: the returned permanent must carry a fresh battlefield-entry stamp")
	}
	if n := counterOn(g, land, CounterPlusOne); n != 0 {
		t.Errorf("the new object kept %d +1/+1 counters from the old one", n)
	}
	if !hasEventFor(g, EventLTB, land) {
		t.Error("no leaves-the-battlefield event: the land has to really die, not be replaced")
	}
	// Fires once and ceases to exist (CR 603.7b).
	if n := len(g.DelayedTriggers); n != 0 {
		t.Errorf("%d delayed triggers survived the return", n)
	}
}

// TestExileAlsoReturnsTheLand: "dies OR is exiled" is two
// destinations on one trigger, and exile is the half a removal spell
// actually reaches for against a recursive threat.
func TestExileAlsoReturnsTheLand(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	land := earthbendTestLand(t, g, me, "Forest")

	earthbend(t, g, me, land, 3)
	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(land); err != nil {
			t.Fatalf("ExileCardForEffect: %v", err)
		}
	})
	settleStack(t, g)

	if g.Exile.Contains(land) {
		t.Fatal("the exiled land was not returned")
	}
	// The exile return re-mints the instance ID (CR 400.7), so the old
	// one names nothing and the land has to be found by name.
	c := battlefieldCardNamed(g, "Forest")
	if c == nil {
		t.Fatal("no Forest on the battlefield after the return")
	}
	if !c.Tapped {
		t.Error("it came back untapped")
	}
	if c.IsCreature() {
		t.Error("the returned land is still animated")
	}
}

// battlefieldCardNamed is the by-name lookup an exile return forces:
// resetAsNewObjectLocked mints a fresh instance ID on the way in.
func battlefieldCardNamed(g *Game, name string) *Card {
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].Name == name {
			return &g.Battlefield.Cards[i]
		}
	}
	return nil
}

// TestAReplacementThatExilesInsteadOfDyingStillReturnsIt is the CR 614
// half: a Rest in Peace-style replacement sends the dying land to
// exile instead of to a graveyard, and the trigger still fires
// because it reads where the card ACTUALLY went (EventLTB.NewZone,
// after the move's own window settled) rather than assuming the
// graveyard.
func TestAReplacementThatExilesInsteadOfDyingStillReturnsIt(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	land := earthbendTestLand(t, g, me, "Forest")

	earthbend(t, g, me, land, 3)
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(graveyardToExileForTest(land))
	})
	killEarthbentLand(t, g, land)

	if me.Graveyard.Contains(land) {
		t.Fatal("setup: the replacement did not redirect the death to exile")
	}
	if g.Exile.Contains(land) {
		t.Error("the land was left in exile; the trigger names both destinations")
	}
	if battlefieldCardNamed(g, "Forest") == nil {
		t.Error(`"when it dies OR IS EXILED" — a replaced death is still one of the two`)
	}
}

// graveyardToExileForTest is Rest in Peace's clause, narrowed to one
// card: "if it would be put into a graveyard from the battlefield,
// exile it instead".
func graveyardToExileForTest(card uuid.UUID) ReplacementEffect {
	return ReplacementEffect{
		Watches: []EventKind{EventZoneMove},
		AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
			return ev.Kind == RepEventMove && ev.CardID == card &&
				ev.OldZone == ZoneBattlefield && ev.NewZone == ZoneGraveyard
		},
		Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
			ev.NewZone = ZoneExile
			return nil
		},
		Label: "Rest in Peace (probe)",
	}
}

// TestASecondEarthbendAddsCountersAndDoesNotStackASecondReturn: the
// delayed trigger's identity is the OBJECT it watches, so earthbending
// the same land twice puts two batches of counters on it and leaves
// ONE return owed. In paper both delayed abilities would trigger and
// the second would find a new object and do nothing; one queue entry
// is the same game state with less bookkeeping (see
// scheduleEarthbendReturnLocked's declared divergence).
func TestASecondEarthbendAddsCountersAndDoesNotStackASecondReturn(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	land := earthbendTestLand(t, g, me, "Forest")

	earthbend(t, g, me, land, 2)
	earthbend(t, g, me, land, 3)

	if got := counterOn(g, land, CounterPlusOne); got != 5 {
		t.Errorf("+1/+1 counters = %d, want 5 — a second earthbend adds its own", got)
	}
	if c := earthbentView(t, g, land); c.CurrentPower() != 5 {
		t.Errorf("power = %d, want 5", c.CurrentPower())
	}
	if n := len(g.DelayedTriggers); n != 1 {
		t.Errorf("%d delayed returns queued, want 1 — the trigger is keyed on the object", n)
	}

	// And the land comes back exactly once.
	killEarthbentLand(t, g, land)
	if n := countBattlefieldNamed(g, "Forest"); n != 1 {
		t.Errorf("%d Forests on the battlefield after one death, want 1", n)
	}
}

// countBattlefieldNamed counts battlefield permanents with that name.
func countBattlefieldNamed(g *Game, name string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == name {
			n++
		}
	}
	return n
}

// TestEarthbendingTheSameLandAfterItReturnsSchedulesAFreshReturn is
// the other side of the object key: the returned land is a NEW object
// with a new battlefield-entry stamp, so the second earthbend is not
// deduplicated against the first.
func TestEarthbendingTheSameLandAfterItReturnsSchedulesAFreshReturn(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	land := earthbendTestLand(t, g, me, "Forest")

	earthbend(t, g, me, land, 1)
	killEarthbentLand(t, g, land)
	if n := len(g.DelayedTriggers); n != 0 {
		t.Fatalf("%d delayed triggers survived the return", n)
	}

	again := battlefieldCardNamed(g, "Forest")
	if again == nil {
		t.Fatal("the land did not come back")
	}
	earthbend(t, g, me, again.InstanceID, 1)
	if n := len(g.DelayedTriggers); n != 1 {
		t.Errorf("%d delayed returns after earthbending the returned land, want 1", n)
	}
}

// TestABouncedEarthbentLandIsNotReturned: the trigger names two
// destinations and a hand is neither. The land stays in hand, and the
// pin sweeps the now-unsatisfiable trigger at the turn's end rather
// than carrying it for the rest of the game.
func TestABouncedEarthbentLandIsNotReturned(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	land := earthbendTestLand(t, g, me, "Forest")

	earthbend(t, g, me, land, 2)
	g.WithWriteLock(func() {
		if err := g.BounceToHandForEffect(land); err != nil {
			t.Fatalf("BounceToHandForEffect: %v", err)
		}
	})
	settleStack(t, g)

	if !me.Hand.Contains(land) {
		t.Fatal("setup: the bounced land is not in hand")
	}
	if g.Battlefield.Contains(land) {
		t.Error("a bounce is neither dying nor being exiled; nothing should have returned it")
	}
	g.WithWriteLock(func() { g.clearExpiredDelayedTriggersLocked(true) })
	if n := len(g.DelayedTriggers); n != 0 {
		t.Errorf("%d unsatisfiable triggers survived the sweep; the pin is the garbage collector", n)
	}
}

// TestAReturnWhoseLandHasMovedOnIsSilent is the CR 608.2b posture on
// the trigger's own effect: the land dies, and somebody reanimates it
// out of the graveyard before the return resolves. The object the
// trigger names is not where it was left, so the return does NOTHING
// — silently. An instruction that legally does nothing is a
// resolution, not a card that threw, and an EventEffectError here
// would fail a catalog soak over a perfectly legal board.
func TestAReturnWhoseLandHasMovedOnIsSilent(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	land := earthbendTestLand(t, g, me, "Forest")

	earthbend(t, g, me, land, 3)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(land); err != nil {
			t.Fatalf("DestroyPermanentForEffect: %v", err)
		}
		// The trigger is on the stack now, and the land is in the
		// graveyard. Take it out from under the trigger.
		if err := g.ReturnFromGraveyardUnderControlForEffect(land, ZoneBattlefield, me.ID); err != nil {
			t.Fatalf("ReturnFromGraveyardUnderControlForEffect: %v", err)
		}
	})
	base := len(g.Events)
	settleStack(t, g)

	if countBattlefieldNamed(g, "Forest") != 1 {
		t.Errorf("%d Forests on the battlefield, want 1 — the return must not double the land",
			countBattlefieldNamed(g, "Forest"))
	}
	for _, ev := range g.Events[base:] {
		if ev.Kind == EventEffectError {
			t.Errorf("the return errored on a land that had moved on: %q", ev.ErrorMsg)
		}
	}
	assertTableIsFree(t, g)
}

// --- the keyword-action window ---------------------------------------

// TestTheKeywordActionWindowCanRewriteEarthbendsCount is what makes
// earthbend a keyword action rather than a bespoke primitive: the
// count rides a CR 614 event, so "if you would earthbend, earthbend
// twice that much instead" is a replacement and not an engine change.
//
// Nothing printed does this yet; the seam is the point (ADR 0013 §5s).
func TestTheKeywordActionWindowCanRewriteEarthbendsCount(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	land := earthbendTestLand(t, g, me, "Forest")

	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(
			keywordCountReplacement(KeywordActionEarthbend, timesTwoForTest, "earthbend twice that much"))
	})
	earthbend(t, g, me, land, 3)

	if got := counterOn(g, land, CounterPlusOne); got != 6 {
		t.Errorf("+1/+1 counters = %d, want 6 — the window doubled the printed 3", got)
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("one replacement queued %d prompts, want none", len(g.PendingChoices))
	}
}

// TestAProliferateDoublerDoesNotSeeAnEarthbend: the action is the
// discriminator, so Tekuthal is not woken by a land animation. One
// event kind, four actions, no crosstalk.
func TestAProliferateDoublerDoesNotSeeAnEarthbend(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	land := earthbendTestLand(t, g, me, "Forest")

	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(
			keywordCountReplacement(KeywordActionProliferate, timesTwoForTest, "proliferate twice"))
	})
	earthbend(t, g, me, land, 3)

	if got := counterOn(g, land, CounterPlusOne); got != 3 {
		t.Errorf("+1/+1 counters = %d, want 3 — a proliferate doubler must not narrow onto earthbend", got)
	}
}

// TestACancelledEarthbendDoesNothing is CR 614.10's null replacement.
// Unlike a count of zero, a CANCEL means the action is not taken at
// all: no animation, no haste, no counters, no delayed return.
func TestACancelledEarthbendDoesNothing(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	land := earthbendTestLand(t, g, me, "Forest")

	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(keywordCancelReplacement(KeywordActionEarthbend))
	})
	earthbend(t, g, me, land, 4)

	if c := earthbentView(t, g, land); c.IsCreature() {
		t.Error("a cancelled earthbend animated the land anyway")
	}
	if got := counterOn(g, land, CounterPlusOne); got != 0 {
		t.Errorf("+1/+1 counters = %d, want 0", got)
	}
	if n := len(g.DelayedTriggers); n != 0 {
		t.Errorf("a cancelled earthbend queued %d delayed returns", n)
	}
	if n := len(g.ScopedStatics); n != 0 {
		t.Errorf("a cancelled earthbend registered %d continuous effects", n)
	}
}

// TestEarthbendingALandThatHasLeftIsANoOp: CR 608.2b has already had
// its say about the target by the time the action is taken, so a land
// that left between the re-check and the verb is silently skipped
// rather than erroring — and nothing is registered against an object
// that is not there.
func TestEarthbendingALandThatHasLeftIsANoOp(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	land := earthbendTestLand(t, g, me, "Forest")
	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(land); err != nil {
			t.Fatalf("ExileCardForEffect: %v", err)
		}
	})

	earthbend(t, g, me, land, 4)

	if n := len(g.ScopedStatics); n != 0 {
		t.Errorf("%d continuous effects registered against a land that is not on the battlefield", n)
	}
	if n := len(g.DelayedTriggers); n != 0 {
		t.Errorf("%d delayed returns queued for a land that is not on the battlefield", n)
	}
}
