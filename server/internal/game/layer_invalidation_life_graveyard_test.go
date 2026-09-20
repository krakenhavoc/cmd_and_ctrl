package game

import (
	"testing"

	"github.com/google/uuid"
)

// layer_invalidation_life_graveyard_test.go covers #1117: the two
// inputs a static could read that nothing invalidated on — a player's
// LIFE TOTAL and the contents of a GRAVEYARD.
//
// Every assertion here is made with NO other event in between, which
// is the whole point. Both bugs were invisible under any test that
// moved a permanent afterwards, because a battlefield move drops the
// cached resolution and hides a missing bump. The historical failure
// mode was a card that was right whenever something unrelated
// happened and one event behind the rest of the time.

const (
	lifeStaticOracle      = "life-total-static"
	graveyardCDAOracle    = "graveyard-cda"
	lifeThresholdForTests = 30
)

// lifeGatedPumpForTest is a stub in the Serra Ascendant shape: +5/+5
// while its controller is at lifeThresholdForTests or more, declaring
// the dependency so the listener knows to invalidate.
func lifeGatedPumpForTest(declare bool) StaticAbility {
	return StaticAbility{
		Layer:              Layer7PT,
		SubLayer:           SubLayer7C_Modify,
		DependsOnLifeTotal: declare,
		AppliesTo: func(target *Card, g *Game, source *Card) bool {
			if target.InstanceID != source.InstanceID {
				return false
			}
			p := g.PlayerByIDForEffect(source.Controller)
			return p != nil && p.Life >= lifeThresholdForTests
		},
		Apply: func(c *Characteristic, _ *Card, _ *Game, _ *Card) {
			c.Power += 5
			c.Toughness += 5
		},
	}
}

// graveyardCDAForTest is Tarmogoyf's shape: power equal to the number
// of card types among cards in all graveyards. It declares NOTHING —
// the graveyard bump is unconditional, and that is the property this
// stub exists to prove, since every Lhurgoyf in the catalog was
// written before the bump and none of them could declare anything.
func graveyardCDAForTest() StaticAbility {
	return StaticAbility{
		Layer:    Layer7PT,
		SubLayer: SubLayer7A_CDA,
		AppliesTo: func(target *Card, _ *Game, source *Card) bool {
			return target.InstanceID == source.InstanceID
		},
		Apply: func(c *Characteristic, _ *Card, g *Game, _ *Card) {
			n := DistinctCardTypesInAllGraveyards(g)
			c.Power = n
			c.Toughness = n + 1
		},
	}
}

// pushLifeGatedCreature seeds a 1/1 carrying the stub and returns it.
func pushLifeGatedCreature(t *testing.T, g *Game) uuid.UUID {
	t.Helper()
	owner := g.Seats[0]
	return pushTypedTestCard(g, Card{
		Name:       "Ascendant Stand-In",
		TypeLine:   "Creature — Human Monk",
		OracleID:   lifeStaticOracle,
		Power:      1,
		Toughness:  1,
		Owner:      owner.ID,
		Controller: owner.ID,
	})
}

// --- the life-total gate -------------------------------------------

// TestLayerVersionBumpsOnALifeChangeWhenALifeStaticIsLive is the fix,
// at the level of the version counter.
func TestLayerVersionBumpsOnALifeChangeWhenALifeStaticIsLive(t *testing.T) {
	withStaticAbilities(t, func(oracleID string) []StaticAbility {
		if oracleID == lifeStaticOracle {
			return []StaticAbility{lifeGatedPumpForTest(true)}
		}
		return nil
	})
	g := newActiveGame(t)
	pushLifeGatedCreature(t, g)
	me := g.Seats[0].ID

	before := readLayerVersion(g)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me, -11) })
	if got := readLayerVersion(g); got <= before {
		t.Errorf("layerVersion did not bump on a life change while a life-keyed static was in play: was %d, now %d", before, got)
	}

	// Damage is the other route to a life total, and it does NOT emit
	// EventChangeLife — it emits EventDealDamage, before the write on
	// the non-combat route. Missing this one would leave every burn
	// spell and every combat damage step invisible to the gate.
	before = readLayerVersion(g)
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(uuid.Nil, me, 3) })
	if got := readLayerVersion(g); got <= before {
		t.Errorf("layerVersion did not bump on damage to a player while a life-keyed static was in play: was %d, now %d", before, got)
	}
}

// TestLayerVersionIgnoresALifeChangeForAStaticThatDoesNotDeclareIt is
// the other side of the gate: a board with no life-keyed static pays
// nothing for the many life changes a game of Commander makes.
// TestLayerVersionDoesNotBumpOnIrrelevantEvent, which fires a bare
// EventChangeLife at an empty board, is the same guard from the other
// end and still passes unchanged.
func TestLayerVersionIgnoresALifeChangeForAStaticThatDoesNotDeclareIt(t *testing.T) {
	withStaticAbilities(t, func(oracleID string) []StaticAbility {
		if oracleID == lifeStaticOracle {
			return []StaticAbility{lifeGatedPumpForTest(false)}
		}
		return nil
	})
	g := newActiveGame(t)
	pushLifeGatedCreature(t, g)
	me := g.Seats[0].ID

	before := readLayerVersion(g)
	g.WithWriteLock(func() {
		_ = g.ChangePlayerLifeForEffect(uuid.Nil, me, -11)
		_ = g.DealDamageToPlayerForEffect(uuid.Nil, me, 3)
	})
	if got := readLayerVersion(g); got != before {
		t.Errorf("layerVersion bumped on a life change with no life-keyed static in play: was %d, now %d", before, got)
	}
}

// TestALifeKeyedStaticIsCorrectImmediatelyAfterALifeChange is the
// behaviour the counter exists for: the recompute happens on the READ
// after the change, with nothing else in between.
func TestALifeKeyedStaticIsCorrectImmediatelyAfterALifeChange(t *testing.T) {
	withStaticAbilities(t, func(oracleID string) []StaticAbility {
		if oracleID == lifeStaticOracle {
			return []StaticAbility{lifeGatedPumpForTest(true)}
		}
		return nil
	})
	g := newActiveGame(t)
	id := pushLifeGatedCreature(t, g)
	me := g.Seats[0].ID

	if got := layeredBattlefieldCard(t, g, id).Effective().Power; got != 6 {
		t.Fatalf("power at %d life = %d, want 6", g.Seats[0].Life, got)
	}
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me, -11) })
	if got := layeredBattlefieldCard(t, g, id).Effective().Power; got != 1 {
		t.Errorf("power right after dropping to %d life = %d, want 1", g.Seats[0].Life, got)
	}
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me, +1) })
	if got := layeredBattlefieldCard(t, g, id).Effective().Power; got != 6 {
		t.Errorf("power right after regaining to %d life = %d, want 6", g.Seats[0].Life, got)
	}
}

// TestALifeKeyedStaticIsCorrectImmediatelyAfterDamage is the ordering
// half, and the reason the bump rides the WRITE rather than an event.
//
// The non-combat damage route emits EventDealDamage BEFORE it changes
// the life total (applyResolvedDamageToPlayerLocked keeps both orders
// on purpose). A listener arm on that event would therefore invalidate
// while the old total was still in place, and any read taken inside
// the emit — the trigger harvester asks for effective characteristics
// — would recache the stale answer with nothing left to invalidate it.
// The card would then shrink exactly one Lightning Bolt late, forever.
func TestALifeKeyedStaticIsCorrectImmediatelyAfterDamage(t *testing.T) {
	withStaticAbilities(t, func(oracleID string) []StaticAbility {
		if oracleID == lifeStaticOracle {
			return []StaticAbility{lifeGatedPumpForTest(true)}
		}
		return nil
	})
	g := newActiveGame(t)
	id := pushLifeGatedCreature(t, g)
	me := g.Seats[0].ID

	if got := layeredBattlefieldCard(t, g, id).Effective().Power; got != 6 {
		t.Fatalf("power at %d life = %d, want 6", g.Seats[0].Life, got)
	}
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(uuid.Nil, me, 11) })
	if g.Seats[0].Life != StartingLife-11 {
		t.Fatalf("life after 11 damage = %d, want %d", g.Seats[0].Life, StartingLife-11)
	}
	if got := layeredBattlefieldCard(t, g, id).Effective().Power; got != 1 {
		t.Errorf("power right after taking 11 damage to %d life = %d, want 1", g.Seats[0].Life, got)
	}
}

// --- the graveyard bump --------------------------------------------

// TestLayerVersionBumpsOnAGraveyardCrossing covers the three shapes
// that were missing: a card arriving from a library (a mill), from a
// hand (a discard) and from the stack (a spell resolving) — plus a
// card LEAVING a graveyard, which shrinks a Lhurgoyf exactly as an
// arrival grows it.
func TestLayerVersionBumpsOnAGraveyardCrossing(t *testing.T) {
	for _, tc := range []struct {
		name             string
		kind             EventKind
		oldZone, newZone ZoneKind
	}{
		{"mill", EventZoneMove, ZoneLibrary, ZoneGraveyard},
		{"discard", EventDiscardCard, ZoneHand, ZoneGraveyard},
		{"spell resolving", EventZoneMove, ZoneStack, ZoneGraveyard},
		{"graveyard exiled", EventZoneMove, ZoneGraveyard, ZoneExile},
		{"reanimated to hand", EventZoneMove, ZoneGraveyard, ZoneHand},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			before := readLayerVersion(g)
			g.WithWriteLock(func() {
				g.EmitEvent(Event{Kind: tc.kind, CardID: uuid.New(), OldZone: tc.oldZone, NewZone: tc.newZone})
			})
			if got := readLayerVersion(g); got <= before {
				t.Errorf("layerVersion did not bump on %s (%s → %s): was %d, now %d", tc.name, tc.oldZone, tc.newZone, before, got)
			}
		})
	}
}

// TestLayerVersionDoesNotBumpOnAMoveTouchingNoGraveyard keeps the
// widening honest: it is a graveyard crossing that bumps, not any
// zone move at all. Library → hand is still the no-op
// TestLayerVersionDoesNotBumpOnZoneMoveOutsideBattlefield asserts.
func TestLayerVersionDoesNotBumpOnAMoveTouchingNoGraveyard(t *testing.T) {
	g := newActiveGame(t)
	before := readLayerVersion(g)
	g.WithWriteLock(func() {
		g.EmitEvent(Event{Kind: EventZoneMove, CardID: uuid.New(), OldZone: ZoneLibrary, NewZone: ZoneExile})
		g.EmitEvent(Event{Kind: EventZoneMove, CardID: uuid.New(), OldZone: ZoneHand, NewZone: ZoneLibrary})
	})
	if got := readLayerVersion(g); got != before {
		t.Errorf("layerVersion bumped on a move that touched no graveyard: was %d, now %d", before, got)
	}
}

// TestAGraveyardCDAIsCorrectImmediatelyAfterAMill is the Lhurgoyf
// assertion, driven through the real mill path rather than a
// hand-fired event — so it also proves the mill emits a zone move the
// listener can see.
func TestAGraveyardCDAIsCorrectImmediatelyAfterAMill(t *testing.T) {
	withStaticAbilities(t, func(oracleID string) []StaticAbility {
		if oracleID == graveyardCDAOracle {
			return []StaticAbility{graveyardCDAForTest()}
		}
		return nil
	})
	g := newActiveGame(t)
	owner := g.Seats[0]
	id := pushTypedTestCard(g, Card{
		Name:       "Goyf Stand-In",
		TypeLine:   "Creature — Lhurgoyf",
		OracleID:   graveyardCDAOracle,
		Owner:      owner.ID,
		Controller: owner.ID,
	})

	// The stock test deck is untyped filler, so seed the top of the
	// library with cards that actually have card types to count.
	owner.Library.PushTop(Card{InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear", Owner: owner.ID})
	owner.Library.PushTop(Card{InstanceID: uuid.New(), Name: "Bolt", TypeLine: "Instant", Owner: owner.ID})

	before := layeredBattlefieldCard(t, g, id).Effective().Power
	g.WithWriteLock(func() {
		if err := g.MillNForEffect(owner.ID, 2); err != nil {
			t.Fatalf("MillNForEffect: %v", err)
		}
	})
	got := layeredBattlefieldCard(t, g, id).Effective().Power
	var want int
	g.ReadSnapshot(func() { want = DistinctCardTypesInAllGraveyards(g) })
	if want <= before {
		t.Fatalf("the mill put no new card type into a graveyard (%d → %d); the fixture cannot prove anything", before, want)
	}
	if got != want {
		t.Errorf("power right after a mill = %d, want %d (stale value was %d)", got, want, before)
	}
}

// --- what the conditional design costs -----------------------------

// BenchmarkLifeChangeWithNoLifeTotalCDA is the life sibling of
// BenchmarkHandMoveWithNoHandSizeCDA: what a life change costs the
// listener at a full four-player board holding nothing that reads a
// life total — the case every table is in almost all of the time, and
// the one the conditional exists to keep cheap.
func BenchmarkLifeChangeWithNoLifeTotalCDA(b *testing.B) {
	g := benchGameWithBoard(b, 40)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.WithWriteLock(func() { g.invalidateLayersForLifeChangeLocked() })
	}
}

// BenchmarkGraveyardArrivalOnABoard is the measurement behind the
// decision NOT to gate the graveyard bump on a declared flag: what one
// mill costs at a 40-permanent board, walk and all.
func BenchmarkGraveyardArrivalOnABoard(b *testing.B) {
	g := benchGameWithBoard(b, 40)
	ev := Event{Kind: EventZoneMove, CardID: uuid.New(), OldZone: ZoneLibrary, NewZone: ZoneGraveyard}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.WithWriteLock(func() { g.EmitEvent(ev) })
	}
}
