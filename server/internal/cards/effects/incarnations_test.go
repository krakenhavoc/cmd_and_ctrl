package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// incarnations_test.go — the catalog half of #1221's statics row. The
// engine half (StaticAbility.Zones, the declared-zone gather, the
// layer-cache invalidation) is in game/static_zones_test.go.

const (
	angerOracle  = "eaabd151-2160-4bff-82c0-3fa88659be98"
	wonderOracle = "232284f7-c623-4895-9ab9-8b1a39926830"
	brawnOracle  = "00876e98-d062-4a12-85e6-86a2b20cf867"
	valorOracle  = "b2ac84e3-cc3c-49c6-918b-a407ef1ee06c"
)

// seedIncarnation drops one incarnation into a seat's graveyard.
func seedIncarnation(p *game.Player, name, oracle string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: "Creature — Incarnation", OracleID: oracle,
		Power: 2, Toughness: 2,
		Owner: p.ID, Controller: p.ID,
	})
	return id
}

// seedBasic puts a basic land on the battlefield under p's control.
func incarnationLand(g *game.Game, p *game.Player, name string) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: "Basic Land — " + name,
		Owner: p.ID, Controller: p.ID,
	})
	return id
}

func incarnationBear(g *game.Game, p *game.Player) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id, Name: "Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2,
		Owner: p.ID, Controller: p.ID,
	})
	return id
}

func bearHasKeyword(t *testing.T, g *game.Game, id uuid.UUID, kw string) bool {
	t.Helper()
	// The seeds below push straight onto a zone, which emits no
	// event, so nothing has bumped the layer version. Force the
	// recompute rather than relying on a staleness the test itself
	// never created — the INVALIDATION is pinned in
	// game/static_zones_test.go, where a real mill does the moving.
	g.BumpLayerVersionForTest()
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		t.Fatalf("creature %s vanished", id)
	}
	return game.HasKeyword(&c, kw)
}

// All four incarnations declare their anthem from a GRAVEYARD and
// from nowhere else — the boot-time invariant, asserted here so a
// future card file that hand-rolls one fails in this package.
func TestIncarnationsDeclareAGraveyardStatic(t *testing.T) {
	for _, tc := range []struct{ name, oracle string }{
		{"Anger", angerOracle},
		{"Wonder", wonderOracle},
		{"Brawn", brawnOracle},
		{"Valor", valorOracle},
	} {
		statics := game.StaticAbilitiesForCard(game.Card{OracleID: tc.oracle})
		if len(statics) != 1 {
			t.Errorf("%s: %d statics, want the one anthem", tc.name, len(statics))
			continue
		}
		s := statics[0]
		if !game.StaticFunctionsFromZone(s, game.ZoneGraveyard) {
			t.Errorf("%s: the anthem does not function from a graveyard", tc.name)
		}
		if game.StaticFunctionsFromZone(s, game.ZoneBattlefield) {
			t.Errorf("%s: the anthem also applies from the battlefield", tc.name)
		}
		if s.Layer != game.Layer6Ability {
			t.Errorf("%s: layer %v, want the ability layer", tc.name, s.Layer)
		}
	}
}

// The card, end to end: Anger in your graveyard plus a Mountain you
// control gives your creatures haste — and either half missing gives
// them nothing.
func TestAngerGrantsHasteFromTheGraveyardWithAMountain(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := incarnationBear(g, me)

	// Neither half.
	if bearHasKeyword(t, g, bear, "haste") {
		t.Fatal("haste with no Anger and no Mountain")
	}
	// The Mountain alone.
	incarnationLand(g, me, "Mountain")
	if bearHasKeyword(t, g, bear, "haste") {
		t.Fatal("haste from a Mountain alone")
	}
	// Both.
	seedIncarnation(me, "Anger", angerOracle)
	if !bearHasKeyword(t, g, bear, "haste") {
		t.Fatal("Anger in the graveyard with a Mountain gave no haste")
	}
}

// The other half of the clause: Anger in the graveyard and no
// Mountain is nothing at all.
func TestAngerWithoutAMountainGrantsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := incarnationBear(g, me)
	seedIncarnation(me, "Anger", angerOracle)
	incarnationLand(g, me, "Forest")
	if bearHasKeyword(t, g, bear, "haste") {
		t.Fatal("haste from Anger with no Mountain")
	}
}

// CR 108.4 at the card level: an Anger in an OPPONENT's graveyard is
// theirs, not yours. "You" off the battlefield is the owner.
func TestAngerInAnotherSeatsGraveyardHelpsThatSeat(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	mine := incarnationBear(g, me)
	theirs := incarnationBear(g, them)
	incarnationLand(g, me, "Mountain")
	incarnationLand(g, them, "Mountain")
	seedIncarnation(them, "Anger", angerOracle)

	if bearHasKeyword(t, g, mine, "haste") {
		t.Error("another seat's Anger gave my creature haste")
	}
	if !bearHasKeyword(t, g, theirs, "haste") {
		t.Error("their own Anger gave their creature nothing")
	}
}

// CR 113.6 on the card: Anger ON THE BATTLEFIELD is a 2/2 with its
// own printed haste and no anthem. The printed keyword rides the
// imported card, so the team getting haste here would be the catalog
// entry saying something the card does not.
func TestAngerOnTheBattlefieldGrantsNoAnthem(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := incarnationBear(g, me)
	incarnationLand(g, me, "Mountain")
	g.Battlefield.PushTop(game.Card{
		InstanceID: uuid.New(), Name: "Anger", TypeLine: "Creature — Incarnation",
		OracleID: angerOracle, Power: 2, Toughness: 2,
		Owner: me.ID, Controller: me.ID,
	})
	if bearHasKeyword(t, g, bear, "haste") {
		t.Fatal("a battlefield Anger granted the graveyard anthem")
	}
}

// The other three, each with its own land and its own keyword, so a
// helper that got the pairing wrong for one of them is caught.
func TestTheOtherIncarnationsGrantTheirOwnKeyword(t *testing.T) {
	for _, tc := range []struct{ name, oracle, land, keyword string }{
		{"Wonder", wonderOracle, "Island", "flying"},
		{"Brawn", brawnOracle, "Forest", "trample"},
		{"Valor", valorOracle, "Plains", "first strike"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			bear := incarnationBear(g, me)
			seedIncarnation(me, tc.name, tc.oracle)
			// The wrong land first: the clause names one.
			incarnationLand(g, me, "Swamp")
			if bearHasKeyword(t, g, bear, tc.keyword) {
				t.Fatalf("%s granted %s off a Swamp", tc.name, tc.keyword)
			}
			incarnationLand(g, me, tc.land)
			if !bearHasKeyword(t, g, bear, tc.keyword) {
				t.Fatalf("%s with a %s granted no %s", tc.name, tc.land, tc.keyword)
			}
		})
	}
}

// A dual land is a Mountain (CR 305.6): the clause names a basic land
// TYPE, not the basic supertype, so a Sacred Foundry turns Anger on.
func TestAngerReadsABasicLandTypeNotTheSupertype(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := incarnationBear(g, me)
	seedIncarnation(me, "Anger", angerOracle)
	g.Battlefield.PushTop(game.Card{
		InstanceID: uuid.New(), Name: "Sacred Foundry",
		TypeLine: "Land — Mountain Plains",
		Owner:    me.ID, Controller: me.ID,
	})
	if !bearHasKeyword(t, g, bear, "haste") {
		t.Fatal("a nonbasic Mountain did not satisfy the clause")
	}
}
