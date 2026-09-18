package game

import (
	"testing"

	"github.com/google/uuid"
)

// cast_zones_test.go — S29. The foundation PR moves "which zone may
// a cast come out of" from a hard-coded whitelist to a per-card
// declaration, and the tests worth having are the ones that would
// pass against a whitelist that simply added "graveyard" to the
// switch:
//
//  1. The graveyard is closed by DEFAULT. A sorcery with no
//     declaration must not become castable from the graveyard just
//     because the zone now resolves.
//  2. A declared zone with a price must CHARGE it. The failure mode
//     is silent and in the player's favour: casting Faithless
//     Looting out of the graveyard for its printed {R}.
//  3. The binding cuts both ways. Flashback claimed from hand is the
//     same bug seen from the other side, and it is the one a client
//     bug produces most easily.
//
// Everything here stubs the two catalog hooks rather than
// registering real cards — the mechanics and their cards ride later
// PRs, and the gate has to be correct before any of them exist.

func withCatalogCastableZones(t *testing.T, fn func(oracleID string) []ZoneKind) {
	t.Helper()
	prev := CatalogCastableZones
	CatalogCastableZones = fn
	t.Cleanup(func() { CatalogCastableZones = prev })
}

// castableZonesFor wires a single card's declaration and returns the
// hook body, mirroring altCostFor in alternative_cost_test.go.
func castableZonesFor(oracle string, zones ...ZoneKind) func(string) []ZoneKind {
	return func(id string) []ZoneKind {
		if id == oracle {
			return zones
		}
		return nil
	}
}

// looterInGraveyard seeds a Faithless-Looting-shaped sorcery — {R}
// printed, flashback {2}{R} — in the active seat's graveyard at a
// main phase.
func looterInGraveyard(t *testing.T, g *Game, me *Player, oracle string) uuid.UUID {
	t.Helper()
	advanceTo(t, g, StepPrecombatMain)
	c := NewCard("Test Looting", me.ID)
	c.TypeLine = "Sorcery"
	c.ManaCost = "{R}"
	c.OracleID = oracle
	me.Graveyard.PushTop(c)
	return c.InstanceID
}

// flashbackCost is the S29 shape: an alternative cost bound to the
// graveyard. Nothing in the engine knows the word — the binding is
// the whole mechanic as far as the cast path is concerned.
func flashbackCost(cost string) AlternativeCost {
	return AlternativeCost{
		Key:      "flashback",
		Label:    "Flashback " + cost,
		ManaCost: cost,
		FromZone: ZoneGraveyard,
	}
}

// The default. A card that declares nothing is a hand card, and the
// graveyard resolving to a real zone must not change that.
func TestGraveyardCastRefusedWithoutDeclaration(t *testing.T) {
	const oracle = "test-looting"
	g := newActiveGame(t)
	me := g.Seats[0]
	id := looterInGraveyard(t, g, me, oracle)

	err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: "graveyard"})
	if err != ErrCastZoneNotAllowed {
		t.Fatalf("undeclared graveyard cast: got %v, want ErrCastZoneNotAllowed", err)
	}
	if !me.Graveyard.Contains(id) {
		t.Errorf("rejected cast moved the card out of the graveyard")
	}
}

// The happy path, and the one fact every mechanic in S29 rests on:
// the card leaves the graveyard, reaches the stack, and the stack
// item remembers where it came from so "cast from a graveyard"
// triggers and Wash Away's clause can ask.
func TestGraveyardCastWithBoundCostReachesTheStack(t *testing.T) {
	const oracle = "test-looting"
	g := newActiveGame(t)
	me := g.Seats[0]
	withCatalogCastableZones(t, castableZonesFor(oracle, ZoneGraveyard))
	withCatalogAlternativeCosts(t, altCostFor(oracle, flashbackCost("{2}{R}")))
	id := looterInGraveyard(t, g, me, oracle)

	if err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback"}); err != nil {
		t.Fatalf("flashback cast: %v", err)
	}
	if me.Graveyard.Contains(id) {
		t.Errorf("cast left the card in the graveyard")
	}
	if !g.Stack.Contains(id) {
		t.Fatalf("cast did not reach the stack")
	}
	item := g.StackMeta[id]
	if item == nil {
		t.Fatalf("no stack item for the flashback cast")
	}
	if item.CastFromZone != ZoneGraveyard {
		t.Errorf("CastFromZone: got %q, want %q", item.CastFromZone, ZoneGraveyard)
	}
	if item.AltCost != "flashback" {
		t.Errorf("AltCost: got %q, want %q", item.AltCost, "flashback")
	}
}

// The silent-and-in-your-favour failure. A declared graveyard with a
// bound price must not also offer the printed one.
func TestGraveyardCastRequiresItsBoundCost(t *testing.T) {
	const oracle = "test-looting"
	g := newActiveGame(t)
	me := g.Seats[0]
	withCatalogCastableZones(t, castableZonesFor(oracle, ZoneGraveyard))
	withCatalogAlternativeCosts(t, altCostFor(oracle, flashbackCost("{2}{R}")))
	id := looterInGraveyard(t, g, me, oracle)

	err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: "graveyard"})
	if err != ErrCastCostRequired {
		t.Fatalf("graveyard cast with no cost claimed: got %v, want ErrCastCostRequired", err)
	}
	if !me.Graveyard.Contains(id) {
		t.Errorf("rejected cast moved the card out of the graveyard")
	}
}

// Gravecrawler: the permission is printed without a price, so the
// printed mana cost is what a graveyard cast owes. The rule above
// must not overreach into "every graveyard cast needs a claim".
func TestGraveyardCastWithoutABoundCostPaysPrinted(t *testing.T) {
	const oracle = "test-crawler"
	g := newActiveGame(t)
	me := g.Seats[0]
	withCatalogCastableZones(t, castableZonesFor(oracle, ZoneGraveyard))
	id := looterInGraveyard(t, g, me, oracle)
	me.ManaPool.AddMana(ManaToken{Color: "R"})

	if err := g.CastSpell(me.ID, id, CastSpellParams{Strict: true, FromZone: "graveyard"}); err != nil {
		t.Fatalf("printed-cost graveyard cast: %v", err)
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("printed cost not charged: pool %+v", me.ManaPool)
	}
}

// The binding seen from the other side. Claiming flashback on a card
// in hand would charge {2}{R} for a {R} sorcery — harmless here, and
// a real discount on every card whose flashback is cheaper than its
// printed cost (Deep Analysis, Past in Flames off a Ritual).
func TestZoneBoundCostCannotBeClaimedFromHand(t *testing.T) {
	const oracle = "test-looting"
	g := newActiveGame(t)
	me := g.Seats[0]
	withCatalogCastableZones(t, castableZonesFor(oracle, ZoneGraveyard))
	withCatalogAlternativeCosts(t, altCostFor(oracle, flashbackCost("{2}{R}")))
	advanceTo(t, g, StepPrecombatMain)
	c := NewCard("Test Looting", me.ID)
	c.TypeLine = "Sorcery"
	c.ManaCost = "{R}"
	c.OracleID = oracle
	me.Hand.PushTop(c)

	err := g.CastSpell(me.ID, c.InstanceID, CastSpellParams{AlternativeCost: "flashback"})
	if err != ErrCastZoneNotAllowed {
		t.Fatalf("flashback from hand: got %v, want ErrCastZoneNotAllowed", err)
	}
	if !me.Hand.Contains(c.InstanceID) {
		t.Errorf("rejected cast moved the card out of hand")
	}
}

// INSTEAD, not alongside — the same property alternative_cost_test
// pins for overload, re-pinned across the zone boundary because the
// graveyard cast takes a different route into printedCostLocked.
func TestGraveyardCastChargesTheBoundCostNotThePrinted(t *testing.T) {
	const oracle = "test-looting"

	t.Run("three mana pays the flashback", func(t *testing.T) {
		g := newActiveGame(t)
		me := g.Seats[0]
		withCatalogCastableZones(t, castableZonesFor(oracle, ZoneGraveyard))
		withCatalogAlternativeCosts(t, altCostFor(oracle, flashbackCost("{2}{R}")))
		id := looterInGraveyard(t, g, me, oracle)
		me.ManaPool.AddMana(ManaToken{Color: "R"})
		me.ManaPool.AddMana(ManaToken{Color: "R"})
		me.ManaPool.AddMana(ManaToken{Color: "R"})

		if err := g.CastSpell(me.ID, id, CastSpellParams{Strict: true, FromZone: "graveyard", AlternativeCost: "flashback"}); err != nil {
			t.Fatalf("flashback on three red: %v", err)
		}
		if len(me.ManaPool) != 0 {
			t.Errorf("flashback cost not fully charged: pool %+v", me.ManaPool)
		}
	})

	t.Run("one mana does not", func(t *testing.T) {
		g := newActiveGame(t)
		me := g.Seats[0]
		withCatalogCastableZones(t, castableZonesFor(oracle, ZoneGraveyard))
		withCatalogAlternativeCosts(t, altCostFor(oracle, flashbackCost("{2}{R}")))
		id := looterInGraveyard(t, g, me, oracle)
		me.ManaPool.AddMana(ManaToken{Color: "R"})

		err := g.CastSpell(me.ID, id, CastSpellParams{Strict: true, FromZone: "graveyard", AlternativeCost: "flashback"})
		var im *InsufficientManaError
		if !errorsAs(err, &im) {
			t.Fatalf("flashback on one red: got %v, want *InsufficientManaError", err)
		}
		if !me.Graveyard.Contains(id) {
			t.Errorf("rejected cast moved the card out of the graveyard")
		}
	})
}

// "Your graveyard", not "a graveyard". The zone lookup is
// seat-scoped, so an opponent naming the card simply doesn't find it
// — no extra predicate needed, which is worth pinning because a
// later refactor to a shared-zone lookup would silently open it.
func TestGraveyardCastIsScopedToTheCastersOwnGraveyard(t *testing.T) {
	const oracle = "test-looting"
	g := newActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	withCatalogCastableZones(t, castableZonesFor(oracle, ZoneGraveyard))
	withCatalogAlternativeCosts(t, altCostFor(oracle, flashbackCost("{2}{R}")))
	id := looterInGraveyard(t, g, me, oracle)

	err := g.CastSpell(them.ID, id, CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback"})
	if err != ErrCardNotFound {
		t.Fatalf("opponent flashback: got %v, want ErrCardNotFound", err)
	}
}

// A from_zone the engine doesn't know is an error rather than a
// silent hand cast. The old castSourceZoneLocked defaulted unknown
// values to hand, which meant a client typo cast the wrong card out
// of the wrong pile with no trace.
func TestUnknownFromZoneIsRejected(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := looterInGraveyard(t, g, me, "test-looting")

	if err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: "nonsense"}); err != ErrZoneNotFound {
		t.Fatalf("from_zone nonsense: got %v, want ErrZoneNotFound", err)
	}
	// "library" became a real zone in S42 (CR 401.5), so it resolves
	// to a pile now and fails on the card not being in it. What must
	// NOT have changed is that it is still no free pass: the card is
	// in the graveyard, the library lookup does not find it, and a
	// card that IS in the library still needs a permission
	// (TestLibraryTopNeedsAPermission).
	if err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: "library"}); err != ErrCardNotFound {
		t.Fatalf("from_zone library on a graveyard card: got %v, want ErrCardNotFound", err)
	}
}

// The exile regression. S29 lets a card's own text open exile, and
// the risk is that the softening leaks: a card that declares nothing
// must still need a live grant.
func TestExileCastStillNeedsAGrantWithoutADeclaration(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	advanceTo(t, g, StepPrecombatMain)
	c := NewCard("Test Bear", me.ID)
	c.TypeLine = "Creature — Bear"
	c.ManaCost = "{1}{G}"
	c.OracleID = "test-bear"
	g.Exile.PushTop(c)

	if err := g.CastSpell(me.ID, c.InstanceID, CastSpellParams{FromZone: "exile"}); err != ErrNoPlayPermission {
		t.Fatalf("ungranted exile cast: got %v, want ErrNoPlayPermission", err)
	}
}

// …and the other half: a card whose own text declares exile
// castable needs no instance grant. The declaration is card-level,
// so it opens exile for every copy of the card at any time. That is
// NOT suspend or foretell: a suspended card may be cast only while
// its last-time-counter trigger resolves (CR 702.62a), and foretold
// status belongs to the exiled instance (CR 702.143). Those need a
// per-instance CastPermission, the way cascade's free cast
// works. No catalog card declares ZoneExile today; this test pins
// the engine path so it doesn't rot unnoticed.
func TestExileCastAllowedByCardLevelDeclaration(t *testing.T) {
	const oracle = "test-exile-castable"
	g := newActiveGame(t)
	me := g.Seats[0]
	withCatalogCastableZones(t, castableZonesFor(oracle, ZoneExile))
	advanceTo(t, g, StepPrecombatMain)
	c := NewCard("Test Exile Castable", me.ID)
	c.TypeLine = "Sorcery"
	c.ManaCost = "{R}"
	c.OracleID = oracle
	g.Exile.PushTop(c)
	me.ManaPool.AddMana(ManaToken{Color: "R"})

	if err := g.CastSpell(me.ID, c.InstanceID, CastSpellParams{Strict: true, FromZone: "exile"}); err != nil {
		t.Fatalf("declared exile cast: %v", err)
	}
	if !g.Stack.Contains(c.InstanceID) {
		t.Errorf("declared exile cast did not reach the stack")
	}
}

// CastableZonesFor never drops hand. Declaring the graveyard ADDS a
// path; a card that lost its ordinary cast because it gained
// flashback would be unplayable from hand, which is the bug a naive
// "return what the catalog said" implementation ships.
func TestCastableZonesAlwaysIncludeHand(t *testing.T) {
	const oracle = "test-looting"
	withCatalogCastableZones(t, castableZonesFor(oracle, ZoneGraveyard))

	if !CardCastableFromZone(oracle, ZoneHand) {
		t.Errorf("a graveyard declaration dropped the hand path")
	}
	if !CardCastableFromZone(oracle, ZoneGraveyard) {
		t.Errorf("the graveyard declaration did not take")
	}
	if CardCastableFromZone(oracle, ZoneBattlefield) {
		t.Errorf("an undeclared zone came back castable")
	}
	if !CardCastableFromZone("some-other-card", ZoneHand) {
		t.Errorf("an undeclared card lost its hand path")
	}
	if CardCastableFromZone("some-other-card", ZoneGraveyard) {
		t.Errorf("an undeclared card gained a graveyard path")
	}
}

// The view layer's filter. A card's offers partition by zone, and
// the client's buttons come straight from this — an unfiltered list
// would put a flashback button on a card in hand.
func TestAlternativeCostsOfferedFromZonePartitionsByZone(t *testing.T) {
	const oracle = "test-both"
	withCatalogAlternativeCosts(t, altCostFor(oracle,
		AlternativeCost{Key: "overload", Label: "Overload {4}{R}", ManaCost: "{4}{R}", ClearsTargets: true},
		flashbackCost("{2}{R}"),
	))

	fromHand := AlternativeCostsOfferedFromZone(oracle, ZoneHand)
	if len(fromHand) != 1 || fromHand[0].Key != "overload" {
		t.Errorf("from hand: got %+v, want overload only", fromHand)
	}
	fromCommand := AlternativeCostsOfferedFromZone(oracle, ZoneCommand)
	if len(fromCommand) != 1 || fromCommand[0].Key != "overload" {
		t.Errorf("from command: got %+v, want overload only", fromCommand)
	}
	fromYard := AlternativeCostsOfferedFromZone(oracle, ZoneGraveyard)
	if len(fromYard) != 1 || fromYard[0].Key != "flashback" {
		t.Errorf("from graveyard: got %+v, want flashback only", fromYard)
	}
	if got := AlternativeCostsOfferedFromZone(oracle, ZoneBattlefield); len(got) != 0 {
		t.Errorf("from battlefield: got %+v, want none", got)
	}
}
