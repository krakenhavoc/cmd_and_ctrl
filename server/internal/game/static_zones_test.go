package game

import (
	"testing"

	"github.com/google/uuid"
)

// static_zones_test.go — the engine half of #1221's statics row:
// StaticAbility.Zones, the layer pass's declared-zone gather, and the
// cache invalidation that keeps it honest. The catalog half — the
// four incarnations — is in cards/effects/incarnations_test.go.

// hasteAnthemFromGraveyard is Anger's clause with nothing else on it:
// "creatures you control have haste", functioning from a GRAVEYARD.
func hasteAnthemFromGraveyard() StaticAbility {
	return StaticAbility{
		Layer: Layer6Ability,
		Zones: []ZoneKind{ZoneGraveyard},
		AppliesTo: func(target *Card, _ *Game, source *Card) bool {
			return target.IsCreature() && target.Controller == source.Controller
		},
		Apply: func(c *Characteristic, _ *Card, _ *Game, _ *Card) {
			c.Abilities = append(c.Abilities, "haste")
		},
	}
}

// withGraveyardStaticCatalog installs a catalog hook and index entry
// for one oracle key, and restores both when the test ends. The index
// is boot-time state, which is exactly why a test that writes it has
// to put it back.
func withGraveyardStaticCatalog(t *testing.T, key string, abilities ...StaticAbility) {
	t.Helper()
	prevHook := CatalogStaticAbilities
	prevIndex := staticZones
	staticZones = newStaticZoneIndex()
	IndexStaticZones(key, abilities)
	CatalogStaticAbilities = func(oracleID string) []StaticAbility {
		if oracleID == key {
			return abilities
		}
		return nil
	}
	t.Cleanup(func() {
		CatalogStaticAbilities = prevHook
		staticZones = prevIndex
	})
}

// seedGraveyardStaticSource puts a card carrying `key` into p's
// graveyard, with `staleController` left on it.
//
// The stale controller is the point of the parameter and not a
// convenience: a card that was under a Mind Control when it died
// carries the STEALER in Card.Controller into the graveyard, and
// CR 108.4 says the ability is its OWNER's from that moment. Seeding
// a card whose Controller already equals its Owner would make the
// gather's `Controller = Owner` line unobservable, which is exactly
// how a back-out of it passes a test that looks like it checks it.
func seedGraveyardStaticSource(p *Player, key string, staleController uuid.UUID) uuid.UUID {
	c := NewCard("Anger", p.ID)
	c.TypeLine = "Creature — Incarnation"
	c.OracleID = key
	if staleController != uuid.Nil {
		c.Controller = staleController
	}
	p.Graveyard.PushTop(c)
	return c.InstanceID
}

const graveyardStaticKey = "00000000-0000-0000-0000-00000000ange"

// The whole of the seam: a static on a card in a graveyard reaches
// the layer pass and applies to the owner's creatures.
func TestGraveyardStaticAppliesToItsOwnersCreatures(t *testing.T) {
	withGraveyardStaticCatalog(t, graveyardStaticKey, hasteAnthemFromGraveyard())
	g := newActiveGame(t)
	me := g.Seats[0]
	bear := NewCard("Bear", me.ID)
	bear.TypeLine = "Creature — Bear"
	bear.Power, bear.Toughness = 2, 2
	g.Battlefield.PushTop(bear)

	g.WithWriteLock(func() { g.recomputeLayersLocked() })
	if got := findBattlefieldCard(g, bear.InstanceID); HasKeyword(got, "haste") {
		t.Fatal("haste before the source is in any graveyard")
	}

	seedGraveyardStaticSource(me, graveyardStaticKey, uuid.Nil)
	g.WithWriteLock(func() { g.recomputeLayersLocked() })
	if got := findBattlefieldCard(g, bear.InstanceID); !HasKeyword(got, "haste") {
		t.Fatal("the graveyard static did not reach the layer pass")
	}
}

// CR 108.4: a card in a graveyard has no controller, so "creatures
// YOU control" is its OWNER's — even when Card.Controller still says
// otherwise. The source here died under a Mind Control and carried
// the stealer's ID into the graveyard with it; the anthem has to
// follow the OWNER anyway, which is what the gather's
// `Controller = Owner` line is for.
func TestGraveyardStaticReadsTheOwnerAsYou(t *testing.T) {
	withGraveyardStaticCatalog(t, graveyardStaticKey, hasteAnthemFromGraveyard())
	g := newActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	mine := NewCard("My Bear", me.ID)
	mine.TypeLine = "Creature — Bear"
	g.Battlefield.PushTop(mine)
	theirs := NewCard("Their Bear", them.ID)
	theirs.TypeLine = "Creature — Bear"
	theirs.Controller = them.ID
	g.Battlefield.PushTop(theirs)

	// Owned by `them`, still carrying MY id as its last controller.
	seedGraveyardStaticSource(them, graveyardStaticKey, me.ID)
	g.WithWriteLock(func() { g.recomputeLayersLocked() })

	if HasKeyword(findBattlefieldCard(g, mine.InstanceID), "haste") {
		t.Error("the anthem followed the stale controller rather than the owner")
	}
	if !HasKeyword(findBattlefieldCard(g, theirs.InstanceID), "haste") {
		t.Error("the anthem did not reach the source owner's own creature")
	}
}

// CR 113.6 in the other direction: a static that declares the
// graveyard does NOT also apply from the battlefield. Anger on the
// board is a 2/2 with its own printed haste and no anthem.
func TestGraveyardStaticDoesNotApplyFromTheBattlefield(t *testing.T) {
	withGraveyardStaticCatalog(t, graveyardStaticKey, hasteAnthemFromGraveyard())
	g := newActiveGame(t)
	me := g.Seats[0]
	bear := NewCard("Bear", me.ID)
	bear.TypeLine = "Creature — Bear"
	g.Battlefield.PushTop(bear)
	src := NewCard("Anger", me.ID)
	src.TypeLine = "Creature — Incarnation"
	src.OracleID = graveyardStaticKey
	g.Battlefield.PushTop(src)

	g.WithWriteLock(func() { g.recomputeLayersLocked() })
	if HasKeyword(findBattlefieldCard(g, bear.InstanceID), "haste") {
		t.Fatal("a graveyard static applied from the battlefield")
	}
}

// The invalidation half. A card ARRIVING in a graveyard from
// somewhere that is not the battlefield — a mill, a discard — has to
// drop the cached layer resolution, or the anthem is one read behind.
// #1117 widened the zone-keyed bump for exactly this; the test is
// here because #1221 is what made it observable.
func TestGraveyardArrivalInvalidatesTheLayerCache(t *testing.T) {
	withGraveyardStaticCatalog(t, graveyardStaticKey, hasteAnthemFromGraveyard())
	g := newActiveGame(t)
	me := g.Seats[0]
	bear := NewCard("Bear", me.ID)
	bear.TypeLine = "Creature — Bear"
	g.Battlefield.PushTop(bear)
	// A read that CACHES the "no haste" answer before the source
	// arrives. Without the bump this is the answer the next read
	// would keep.
	g.WithWriteLock(func() { g.recomputeLayersLocked() })
	if HasKeyword(findBattlefieldCard(g, bear.InstanceID), "haste") {
		t.Fatal("haste before anything is in a graveyard")
	}

	src := NewCard("Anger", me.ID)
	src.TypeLine = "Creature — Incarnation"
	src.OracleID = graveyardStaticKey
	me.Library.PushTop(src)
	g.WithWriteLock(func() {
		if err := g.MillNForEffect(me.ID, 1); err != nil {
			t.Fatalf("MillForEffect: %v", err)
		}
		g.RecomputeLayersIfStaleLocked()
	})
	if !me.Graveyard.Contains(src.InstanceID) {
		t.Fatalf("the mill did not put the source in the graveyard")
	}
	if !HasKeyword(findBattlefieldCard(g, bear.InstanceID), "haste") {
		t.Error("the cached resolution survived the graveyard arrival")
	}
}

// And the departure: exiling the source takes the anthem away at the
// next read.
func TestGraveyardDepartureInvalidatesTheLayerCache(t *testing.T) {
	withGraveyardStaticCatalog(t, graveyardStaticKey, hasteAnthemFromGraveyard())
	g := newActiveGame(t)
	me := g.Seats[0]
	bear := NewCard("Bear", me.ID)
	bear.TypeLine = "Creature — Bear"
	g.Battlefield.PushTop(bear)
	src := seedGraveyardStaticSource(me, graveyardStaticKey, uuid.Nil)
	g.WithWriteLock(func() { g.recomputeLayersLocked() })
	if !HasKeyword(findBattlefieldCard(g, bear.InstanceID), "haste") {
		t.Fatal("the anthem never applied")
	}

	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(src); err != nil {
			t.Fatalf("ExileCardForEffect: %v", err)
		}
		g.RecomputeLayersIfStaleLocked()
	})
	if HasKeyword(findBattlefieldCard(g, bear.InstanceID), "haste") {
		t.Error("the anthem survived the source leaving the graveyard")
	}
}

// The index is the reason a table with no such card pays nothing, so
// its contract is worth pinning: a card that declares no zone is not
// in it, and an unsupported zone is named rather than silently walked.
func TestStaticZoneIndexAndSupportedZones(t *testing.T) {
	ix := newStaticZoneIndex()
	ix.index("plain", []StaticAbility{{Layer: Layer6Ability}})
	if ix.declares("plain") || len(ix.zones()) != 0 {
		t.Errorf("a battlefield-only static was indexed: %+v", ix)
	}
	ix.index("grave", []StaticAbility{hasteAnthemFromGraveyard()})
	if !ix.declares("grave") {
		t.Error("a graveyard static was not indexed")
	}
	if zs := ix.zones(); len(zs) != 1 || zs[0] != ZoneGraveyard {
		t.Errorf("zones = %v, want [graveyard]", zs)
	}
	if why := StaticZoneUnsupported(ZoneGraveyard); why != "" {
		t.Errorf("graveyard refused: %s", why)
	}
	for _, zone := range []ZoneKind{ZoneBattlefield, ZoneHand, ZoneExile, ZoneLibrary} {
		if StaticZoneUnsupported(zone) == "" {
			t.Errorf("%s is accepted but nothing gathers it", zone)
		}
	}
}
