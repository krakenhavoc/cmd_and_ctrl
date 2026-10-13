package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// cast_zones_conditional_test.go — #3105, the 2026-10-13 amendment to
// ADR 0066. A card may open a cast zone for itself only while a
// condition holds (Lightwheel Enhancements: "Max speed — You may cast
// this card from your graveyard"). The engine, the view and the
// enumerator all ask Game.cardOpensZoneLocked, so these tests pin the
// engine half and the shared predicate; the view and the enumerator
// have their own tests beside their code.

const conditionalZoneOracle = "test-conditional-zone"

// withConditionalCastZones stubs the conditional-zone hook for one
// test: the card with `oracle` opens its graveyard while its owner has
// max speed, which is the shape Lightwheel Enhancements uses.
func withConditionalCastZones(t *testing.T, oracle string) {
	t.Helper()
	prev := CatalogCastableZonesIf
	CatalogCastableZonesIf = func(id string) []ConditionalCastZone {
		if id != oracle {
			return nil
		}
		return []ConditionalCastZone{{
			Zone: ZoneGraveyard,
			Cond: func(g *Game, owner uuid.UUID, _ Card) bool { return g.HasMaxSpeed(owner) },
		}}
	}
	t.Cleanup(func() { CatalogCastableZonesIf = prev })
}

func conditionalCardInGraveyard(t *testing.T, g *Game, me *Player, oracle string) uuid.UUID {
	t.Helper()
	advanceTo(t, g, StepPrecombatMain)
	c := NewCard("Test Lightwheel", me.ID)
	c.TypeLine = "Sorcery"
	c.ManaCost = "{W}"
	c.OracleID = oracle
	me.Graveyard.PushTop(c)
	return c.InstanceID
}

// The headline: refused below max speed with the card left where it
// was, accepted at max speed, refused again when the speed falls. The
// card is never rebuilt and nothing is stored on it — the condition is
// read from the game each time.
func TestConditionalZoneOpensAndClosesWithTheCondition(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	withConditionalCastZones(t, conditionalZoneOracle)
	id := conditionalCardInGraveyard(t, g, me, conditionalZoneOracle)

	if err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: "graveyard"}); !errors.Is(err, ErrCastZoneNotAllowed) {
		t.Fatalf("below max speed: got %v, want ErrCastZoneNotAllowed", err)
	}
	if !me.Graveyard.Contains(id) {
		t.Fatalf("a refused cast moved the card out of the graveyard")
	}

	g.SetSpeedForTest(me.ID, MaxSpeed-1)
	if err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: "graveyard"}); !errors.Is(err, ErrCastZoneNotAllowed) {
		t.Fatalf("at speed %d: got %v, want ErrCastZoneNotAllowed", MaxSpeed-1, err)
	}

	g.SetSpeedForTest(me.ID, MaxSpeed)
	card := *findCardForTest(me.Graveyard, id)
	g.WithWriteLock(func() {
		if !g.CardCastableFromAnyFaceLocked(card, ZoneGraveyard) {
			t.Errorf("the shared predicate says closed at max speed")
		}
	})

	g.SetSpeedForTest(me.ID, 1)
	g.WithWriteLock(func() {
		if g.CardCastableFromAnyFaceLocked(card, ZoneGraveyard) {
			t.Errorf("the shared predicate says open after the speed fell")
		}
	})

	g.SetSpeedForTest(me.ID, MaxSpeed)
	if err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: "graveyard"}); err != nil {
		t.Fatalf("at max speed: %v", err)
	}
	if !g.Stack.Contains(id) {
		t.Fatalf("the cast did not reach the stack")
	}
	if item := g.StackMeta[id]; item == nil || item.CastFromZone != ZoneGraveyard {
		t.Errorf("stack item = %+v, want CastFromZone graveyard", item)
	}
}

// "Max speed" is the OWNER's. Another seat reaching max speed opens
// nothing on a card in this seat's graveyard.
func TestConditionalZoneReadsTheOwnersSpeedNotTheTables(t *testing.T) {
	g := newActiveGame(t)
	me, other := g.Seats[0], g.Seats[1]
	withConditionalCastZones(t, conditionalZoneOracle)
	id := conditionalCardInGraveyard(t, g, me, conditionalZoneOracle)

	g.SetSpeedForTest(other.ID, MaxSpeed)
	if err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: "graveyard"}); !errors.Is(err, ErrCastZoneNotAllowed) {
		t.Fatalf("an opponent's max speed opened the graveyard: %v", err)
	}
}

// The difference from a CastCondition: the condition gates only the
// card's OWN opening of the zone. A permission that opens the same
// zone (Past in Flames, Underworld Breach) still does, at any speed.
func TestAPermissionStillOpensAConditionalZoneThatIsClosed(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	advanceTo(t, g, StepPrecombatMain)
	withConditionalCastZones(t, conditionalZoneOracle)
	withCatalogCastPermissions(t, func(id string) []CastPermission {
		if id != "test-breach" {
			return nil
		}
		return []CastPermission{{
			Zone: ZoneGraveyard, Scope: ScopeStanding, Duration: WhileInZoneDuration(),
			Filter: PermissionFilter{NonLandOnly: true},
		}}
	})
	source := permanentFor(g, me, "Test Breach", "Enchantment", "{1}{R}")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == source {
				g.Battlefield.Cards[i].OracleID = "test-breach"
			}
		}
	})
	id := seedGraveyard(me, "Test Lightwheel", "Sorcery", "{W}")
	g.WithWriteLock(func() {
		for i := range me.Graveyard.Cards {
			if me.Graveyard.Cards[i].InstanceID == id {
				me.Graveyard.Cards[i].OracleID = conditionalZoneOracle
			}
		}
	})

	if g.HasMaxSpeed(me.ID) {
		t.Fatal("fixture has max speed; the point is that the condition is false")
	}
	if err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: "graveyard"}); err != nil {
		t.Fatalf("a permission-opened cast was refused because the card's own condition is false: %v", err)
	}
}

// An unconditional card is untouched by the hook: the hook answers nil
// for it and a plain declaration still opens the zone at any speed.
func TestUnconditionalZoneIgnoresTheConditionalHook(t *testing.T) {
	const plain = "test-plain-zone"
	g := newActiveGame(t)
	me := g.Seats[0]
	withConditionalCastZones(t, conditionalZoneOracle)
	withCatalogCastableZones(t, castableZonesFor(plain, ZoneGraveyard))
	id := conditionalCardInGraveyard(t, g, me, plain)

	if err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: "graveyard"}); err != nil {
		t.Fatalf("an unconditional graveyard cast at speed 0: %v", err)
	}
	if !g.Stack.Contains(id) {
		t.Errorf("the cast did not reach the stack")
	}
}

// A card that declares nothing stays a hand card whatever the hook
// does for other cards.
func TestUndeclaredCardStaysClosedWithTheHookWired(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	withConditionalCastZones(t, conditionalZoneOracle)
	id := conditionalCardInGraveyard(t, g, me, "test-undeclared")
	g.SetSpeedForTest(me.ID, MaxSpeed)

	if err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: "graveyard"}); !errors.Is(err, ErrCastZoneNotAllowed) {
		t.Fatalf("undeclared card at max speed: got %v, want ErrCastZoneNotAllowed", err)
	}
}

// Nothing is stored for the condition, so Clone (undo, the bot's
// look-ahead) answers from the clone's own state: speed raised on the
// clone opens the clone and only the clone.
func TestConditionalZoneFollowsTheCloneNotTheOriginal(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	withConditionalCastZones(t, conditionalZoneOracle)
	id := conditionalCardInGraveyard(t, g, me, conditionalZoneOracle)

	clone := g.Clone()
	clone.SetSpeedForTest(me.ID, MaxSpeed)

	card := *findCardForTest(me.Graveyard, id)
	var origOpen, cloneOpen bool
	g.WithWriteLock(func() { origOpen = g.CardCastableFromAnyFaceLocked(card, ZoneGraveyard) })
	clone.WithWriteLock(func() { cloneOpen = clone.CardCastableFromAnyFaceLocked(card, ZoneGraveyard) })
	if origOpen || !cloneOpen {
		t.Errorf("original open = %v, clone open = %v; want false, true", origOpen, cloneOpen)
	}
}
