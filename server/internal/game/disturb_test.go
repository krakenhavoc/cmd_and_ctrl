package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// disturb_test.go — ADR 0107 §4 (#1855), the engine half of disturb:
// an alternative cost that casts a `transform` card's BACK face out of
// its owner's graveyard (CR 702.146a), with the zone and the price read
// off the front face (CR 712.11d) and everything else off the back
// (CR 712.8c). The card files and the back face's exile replacement are
// pinned in effects/disturb_test.go against real cards.

const disturbFixtureOracle = "d15707b0-0000-4000-8000-000000000001"

// disturbOffer is effects.Disturb, spelled out so the game package can
// test the mechanic without importing the effects package.
func disturbOffer(cost string) AlternativeCost {
	return AlternativeCost{
		Key:       "disturb",
		Label:     "Disturb " + cost,
		ManaCost:  cost,
		FromZone:  ZoneGraveyard,
		CastsFace: 1,
	}
}

// disturbFixture is Baithook Angler // Hook-Haunt Drifter's shape: a
// {1}{U} 2/1 front face and a 1/2 back face with no mana cost.
func disturbFixture(owner uuid.UUID) Card {
	c := Card{
		InstanceID: uuid.New(),
		OracleID:   disturbFixtureOracle,
		Owner:      owner,
		Controller: owner,
		Layout:     LayoutTransform,
		Faces: []Face{
			{Name: "Fixture Angler", TypeLine: "Creature — Human Peasant", ManaCost: "{1}{U}", Colors: []string{"U"}, Power: 2, Toughness: 1},
			{Name: "Fixture Drifter", TypeLine: "Creature — Spirit", Colors: []string{"U"}, Power: 1, Toughness: 2},
		},
	}
	c.SetFace(0)
	return c
}

// seedDisturb wires the catalog for the fixture — the front face's
// entry opens the graveyard and offers disturb {1}{U}{U} — and puts one
// in the active seat's graveyard at a main phase.
func seedDisturb(t *testing.T, g *Game, me *Player) uuid.UUID {
	t.Helper()
	withCatalogCastableZones(t, castableZonesFor(disturbFixtureOracle, ZoneGraveyard))
	withCatalogAlternativeCosts(t, altCostFor(disturbFixtureOracle, disturbOffer("{1}{U}{U}")))
	advanceTo(t, g, StepPrecombatMain)
	c := disturbFixture(me.ID)
	me.Graveyard.PushTop(c)
	return c.InstanceID
}

func stackCardForTest(t *testing.T, g *Game, id uuid.UUID) Card {
	t.Helper()
	for _, c := range g.Stack.Cards {
		if c.InstanceID == id {
			return c
		}
	}
	t.Fatalf("card %s is not on the stack", id)
	return Card{}
}

// The headline: a disturb cast puts the BACK face on the stack (CR
// 712.11a) and the permanent enters back face up (CR 702.146b), paying
// the disturb cost rather than the front face's.
func TestDisturbCastsTheBackFaceFromTheGraveyard(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := seedDisturb(t, g, me)

	var quoted CastPrice
	var err error
	g.ReadSnapshot(func() {
		c, _ := g.cardInZoneLocked(me.Graveyard, id)
		quoted, err = g.priceCastLocked(me.ID, c, CastSpellParams{FromZone: "graveyard", AlternativeCost: "disturb"})
	})
	if err != nil {
		t.Fatalf("price the disturb cast: %v", err)
	}
	if quoted.Base.ManaValue() != 3 || quoted.Card.ActiveFace != 1 {
		t.Errorf("disturb priced at mana value %d on face %d, want 3 ({1}{U}{U}) on face 1", quoted.Base.ManaValue(), quoted.Card.ActiveFace)
	}

	me.ManaPool = nil
	for _, color := range []string{"U", "U", "U"} {
		g.WithWriteLock(func() {
			if err := g.AddManaForEffect(me.ID, uuid.Nil, "{"+color+"}"); err != nil {
				t.Fatalf("add mana: %v", err)
			}
		})
	}
	if err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: "graveyard", AlternativeCost: "disturb", Strict: true}); err != nil {
		t.Fatalf("disturb cast: %v", err)
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("disturb left %v floating, want the whole {1}{U}{U} spent", me.ManaPool)
	}
	spell := stackCardForTest(t, g, id)
	if spell.ActiveFace != 1 || spell.Name != "Fixture Drifter" {
		t.Fatalf("spell on the stack is face %d (%q), want the back face", spell.ActiveFace, spell.Name)
	}
	// CR 712.8c: the back face's mana value is the front face's.
	g.ReadSnapshot(func() {
		if mv, ok := g.ManaValueForEffect(spell); !ok || mv != 2 {
			t.Errorf("disturbed spell's mana value = (%d, %v), want (2, true) from the front face's {1}{U}", mv, ok)
		}
	})
	if item := g.StackMeta[id]; item == nil || item.AltCost != "disturb" || item.CastFromZone != ZoneGraveyard {
		t.Errorf("stack item %+v, want disturb from the graveyard", item)
	}

	passPriorityUntilResolvedForTest(t, g, id)
	var landed *Card
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			landed = &g.Battlefield.Cards[i]
		}
	}
	if landed == nil {
		t.Fatal("the disturbed spell never reached the battlefield")
	}
	if landed.ActiveFace != 1 || landed.Name != "Fixture Drifter" || landed.Power != 1 || landed.Toughness != 2 {
		t.Fatalf("permanent is face %d (%q %d/%d), want the 1/2 back face", landed.ActiveFace, landed.Name, landed.Power, landed.Toughness)
	}
	// CR 712.8e: and the permanent's.
	if mv := landed.ManaValue(); mv != 2 {
		t.Errorf("disturbed permanent's mana value = %d, want 2", mv)
	}
	AssertFaceInvariant(t, g)
}

// The claim settles the face. A caller may name the back face or leave
// the face unset (the wire's default, and what the client's graveyard
// menu sends), and nothing else.
func TestDisturbClaimSettlesTheFace(t *testing.T) {
	for _, tc := range []struct {
		face int
		ok   bool
	}{{0, true}, {1, true}, {2, false}} {
		g := newActiveGame(t)
		me := g.Seats[0]
		id := seedDisturb(t, g, me)
		err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: "graveyard", AlternativeCost: "disturb", Face: tc.face})
		if tc.ok && err != nil {
			t.Errorf("face %d: disturb cast refused: %v", tc.face, err)
		}
		if !tc.ok && !errors.Is(err, ErrInvalidFace) {
			t.Errorf("face %d: err = %v, want ErrInvalidFace", tc.face, err)
		}
		if tc.ok && err == nil {
			if got := stackCardForTest(t, g, id); got.ActiveFace != 1 {
				t.Errorf("face %d: spell is face %d, want 1", tc.face, got.ActiveFace)
			}
		}
	}
}

// A zone the card prices must be paid for (validateCastPathLocked rule
// 3): a disturb card in the graveyard is not castable for the front
// face's printed cost, and disturb is not claimable out of a hand.
func TestDisturbIsTheGraveyardsOnlyPrice(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := seedDisturb(t, g, me)
	if err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: "graveyard"}); !errors.Is(err, ErrCastCostRequired) {
		t.Errorf("printed-cost cast from the graveyard: err = %v, want ErrCastCostRequired", err)
	}

	hand := disturbFixture(me.ID)
	me.Hand.PushTop(hand)
	if err := g.CastSpell(me.ID, hand.InstanceID, CastSpellParams{AlternativeCost: "disturb"}); err == nil {
		t.Error("disturb was claimable from a hand")
	}
	// The hand cast is the front face for its printed cost, as ever.
	if err := g.CastSpell(me.ID, hand.InstanceID, CastSpellParams{}); err != nil {
		t.Fatalf("printed-cost cast from hand: %v", err)
	}
	if got := stackCardForTest(t, g, hand.InstanceID); got.ActiveFace != 0 {
		t.Errorf("a hand cast of a disturb card is face %d, want the front", got.ActiveFace)
	}
}

// CR 712.11a casts only a double-faced card transformed. An offer that
// names a face on a card that is not a `transform` card, or that lacks
// the face, is refused rather than clamped to the front face.
func TestDisturbRefusedOnACardWithNoBackFaceToCast(t *testing.T) {
	for name, mutate := range map[string]func(*Card){
		"modal DFC":    func(c *Card) { c.Layout = LayoutModalDFC },
		"single-faced": func(c *Card) { c.Faces = nil },
	} {
		g := newActiveGame(t)
		me := g.Seats[0]
		withCatalogCastableZones(t, castableZonesFor(disturbFixtureOracle, ZoneGraveyard))
		withCatalogAlternativeCosts(t, altCostFor(disturbFixtureOracle, disturbOffer("{1}{U}{U}")))
		advanceTo(t, g, StepPrecombatMain)
		c := disturbFixture(me.ID)
		mutate(&c)
		me.Graveyard.PushTop(c)
		if err := g.CastSpell(me.ID, c.InstanceID, CastSpellParams{FromZone: "graveyard", AlternativeCost: "disturb"}); !errors.Is(err, ErrInvalidFace) {
			t.Errorf("%s: err = %v, want ErrInvalidFace", name, err)
		}
	}
}

// The offer list the view and the bot read (#673): the disturb offer,
// and not the printed cost.
func TestDisturbIsOfferedFromTheGraveyard(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := seedDisturb(t, g, me)
	g.ReadSnapshot(func() {
		c, _ := g.cardInZoneLocked(me.Graveyard, id)
		offers := g.CastOffersForLocked(me.ID, c, ZoneGraveyard, nil)
		if len(offers) != 1 || offers[0] == nil || offers[0].Key != "disturb" {
			t.Fatalf("offers = %v, want disturb alone", offers)
		}
		if got := offers[0].CastFaceOf(c); got.ActiveFace != 1 || got.Name != "Fixture Drifter" {
			t.Errorf("CastFaceOf = face %d (%q), want the back face", got.ActiveFace, got.Name)
		}
	})
}
