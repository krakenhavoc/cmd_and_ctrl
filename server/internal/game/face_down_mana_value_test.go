package game

import (
	"testing"

	"github.com/google/uuid"
)

// face_down_mana_value_test.go — #4682. CR 708.2a: a face-down permanent
// has no mana cost, so its mana value is 0 (CR 202.3a). The card it
// hides still has one, and until #4682 every read that took a mana
// value (ManaValueForEffect, Card.ManaValue, the "mana value N or less"
// filters, the X bound the view ships) looked through the face-down
// state at the hidden card: a sweep spared or caught the wrong
// creatures, and the wire told the table what the 2/2 really was.
//
// Only a CR 708.2 object is zero. A card face down in EXILE
// (Necropotence, foretell, hideaway) is not one and keeps its printed
// cost (TestFaceDownExilePricesAtItsPrintedCost).

// hiddenCost is the cost of the card every fixture here hides: mana
// value 6, far enough from 0 that no accident reads as the right answer.
const hiddenCost = "{4}{B}{B}"

// battlefieldCardOf is the card as the battlefield holds it right now.
func battlefieldCardOf(t *testing.T, g *Game, id uuid.UUID) Card {
	t.Helper()
	c, zone, ok := cardAnywhere(g, id)
	if !ok || zone != ZoneBattlefield {
		t.Fatalf("card %s is in %q (found=%v), want the battlefield", id, zone, ok)
	}
	return c
}

// requireManaValueReads fails unless every card-side read of mana value
// agrees on want. The reads are the ones callers use: the game's
// effect read, and the three Card methods behind it.
func requireManaValueReads(t *testing.T, g *Game, c Card, want int, what string) {
	t.Helper()
	if mv, ok := g.ManaValueForEffect(c); mv != want || !ok {
		t.Errorf("%s: ManaValueForEffect = %d, %v; want %d, true", what, mv, ok, want)
	}
	if mv := c.ManaValue(); mv != want {
		t.Errorf("%s: Card.ManaValue = %d; want %d", what, mv, want)
	}
	if mv, ok := c.ParsedManaValue(); mv != want || !ok {
		t.Errorf("%s: Card.ParsedManaValue = %d, %v; want %d, true", what, mv, ok, want)
	}
	// {X} counts as the chosen value only on the stack, and only for a
	// card that has a cost to put it in: a face-down spell has none.
	if mv := c.ManaValueWithX(3); want == 0 && mv != 0 {
		t.Errorf("%s: Card.ManaValueWithX(3) = %d; want 0", what, mv)
	}
}

// TestFaceDownPermanentHasManaValueZero is the rule for each way the
// engine makes a face-down permanent: it reads zero through every
// mana-value read, and reads the hidden card's cost again once it is
// turned face up (CR 708.8: the real card is back, and with it its
// characteristics).
func TestFaceDownPermanentHasManaValueZero(t *testing.T) {
	// face puts a hiddenCost creature on the battlefield face down.
	type maker func(t *testing.T) (*Game, uuid.UUID)
	makers := []struct {
		name string
		make maker
	}{
		{"manifest (CR 701.40a)", func(t *testing.T) (*Game, uuid.UUID) {
			g, id := faceDownGame(t, "Griselbrand", "Legendary Creature — Demon", "oracle-grissy")
			g.WithWriteLock(func() {
				if _, err := g.ManifestForEffect(g.Seats[0].ID); err != nil {
					t.Fatalf("manifest: %v", err)
				}
			})
			return g, id
		}},
		{"turned face down by an effect (CR 708.2a)", func(t *testing.T) (*Game, uuid.UUID) {
			g := newActiveGame(t)
			id := turnableCreature(t, g, g.Seats[0], "Sheoldred")
			if got := g.TurnFaceDownForEffect(uuid.Nil, id); len(got) != 1 {
				t.Fatalf("turned %v face down, want 1", got)
			}
			return g, id
		}},
		{"morph cast face down (CR 702.37b)", func(t *testing.T) (*Game, uuid.UUID) {
			g, me, id := morphGame(t, morphOffer(FaceDownMorphed, morphCost, false))
			setHiddenCost(g, me, id)
			castFaceDownAndResolve(t, g, me, id)
			return g, id
		}},
		{"disguise cast face down (CR 702.168a)", func(t *testing.T) (*Game, uuid.UUID) {
			g, me, id := morphGame(t, morphOffer(FaceDownDisguised, morphCost, false))
			setHiddenCost(g, me, id)
			castFaceDownAndResolve(t, g, me, id)
			return g, id
		}},
		// Nothing in the catalog cloaks yet (#95), so the state is put on
		// the permanent directly, the way the cloak keyword will leave it.
		{"cloak (CR 701.58a)", func(t *testing.T) (*Game, uuid.UUID) {
			g := newActiveGame(t)
			id := turnableCreature(t, g, g.Seats[0], "Sheoldred")
			g.applyFaceDownLandingLocked(g.Battlefield, id, FaceDownCloaked, nil)
			return g, id
		}},
	}
	for _, m := range makers {
		t.Run(m.name, func(t *testing.T) {
			g, id := m.make(t)
			c := battlefieldCardOf(t, g, id)
			if !c.FaceDownIsPermanent() {
				t.Fatalf("the fixture is not a face-down permanent: FaceDown=%v kind=%q", c.FaceDown, c.FaceDownKind)
			}
			if c.ManaCost != hiddenCost {
				t.Fatalf("the hidden card's cost = %q, want %q (the fixture drifted)", c.ManaCost, hiddenCost)
			}

			requireManaValueReads(t, g, c, 0, "face down")

			// The hidden card is untouched: turning it face up shows it.
			c.ClearFaceDown()
			requireManaValueReads(t, g, c, 6, "face up")
		})
	}
}

// setHiddenCost makes the hand card the morph fixture a hiddenCost
// creature, so the face-down permanent hides a mana value of 6 whatever
// the morph fixture prints.
func setHiddenCost(g *Game, p *Player, id uuid.UUID) {
	g.WithWriteLock(func() {
		for i := range p.Hand.Cards {
			if p.Hand.Cards[i].InstanceID == id {
				p.Hand.Cards[i].ManaCost = hiddenCost
			}
		}
	})
}

// castFaceDownAndResolve casts the morph fixture face down and resolves
// it, so the permanent that results is the engine's own.
func castFaceDownAndResolve(t *testing.T, g *Game, p *Player, id uuid.UUID) {
	t.Helper()
	if err := castFaceDown(t, g, p, id, "morph"); err != nil {
		t.Fatalf("cast face down: %v", err)
	}
	// On the stack it is already zero: a face-down spell has no cost.
	c, zone, ok := cardAnywhere(g, id)
	if !ok || zone != ZoneStack {
		t.Fatalf("card is in %q, want the stack", zone)
	}
	requireManaValueReads(t, g, c, 0, "face-down spell on the stack")
	resolveTop(t, g)
}

// TestFaceDownExilePricesAtItsPrintedCost keeps the other half of the
// partition: CR 708.2 is about permanents (and morph spells), so a card
// face down in exile is not zero. Its owner may be entitled to cast it
// (foretell, Gonti, Outrageous Robbery), and what it costs is the card's
// own.
func TestFaceDownExilePricesAtItsPrintedCost(t *testing.T) {
	for _, kind := range []FaceDownKind{FaceDownExiled, FaceDownForetold, FaceDownHidden, FaceDownPermitted} {
		t.Run(string(kind), func(t *testing.T) {
			g, id := faceDownGame(t, "Griselbrand", "Legendary Creature — Demon", "oracle-grissy")
			me := g.Seats[0]
			g.WithWriteLock(func() {
				if _, err := g.routeCardToZoneLocked(zoneRoute{
					CardID: id, Dst: ZoneExile, Actor: me.ID, FaceDown: kind,
				}); err != nil {
					t.Fatalf("route: %v", err)
				}
			})
			c, zone, ok := cardAnywhere(g, id)
			if !ok || zone != ZoneExile || !c.FaceDown {
				t.Fatalf("card is in %q face down = %v (found=%v), want face-down exile", zone, c.FaceDown, ok)
			}
			requireManaValueReads(t, g, c, 6, "face down in exile")
		})
	}
}

// TestManaValueBoundAdmitsAFaceDownPermanentAsZero is the X bound on a
// target clause ("mana value X or less", Quillmane Baku): the number
// the engine judges a candidate by is the number BoundValuesForEffect
// ships to the client to narrow the same clause, so both are 0 for a
// face-down permanent. Before #4682 the second told the table the
// hidden card's cost.
func TestManaValueBoundAdmitsAFaceDownPermanentAsZero(t *testing.T) {
	g := newActiveGame(t)
	opp := g.Seats[1]
	hidden := turnableCreature(t, g, opp, "Sheoldred")
	faceUp := turnableCreature(t, g, opp, "Griselbrand")
	if got := g.TurnFaceDownForEffect(uuid.Nil, hidden); len(got) != 1 {
		t.Fatalf("turned %v face down, want 1", got)
	}

	spec := (&TargetSpec{Zones: []ZoneKind{ZoneBattlefield}}).WithManaValueAtMostX()
	var vals map[uuid.UUID]int
	g.WithWriteLock(func() {
		vals = g.BoundValuesForEffect(spec, []uuid.UUID{hidden, faceUp})
	})
	if v, ok := vals[hidden]; !ok || v != 0 {
		t.Errorf("the shipped mana value of a face-down permanent = %d, %v; want 0, true", v, ok)
	}
	if vals[faceUp] != 6 {
		t.Errorf("the shipped mana value of a face-up {4}{B}{B} = %d; want 6", vals[faceUp])
	}

	// Bound to X=0, only the face-down permanent qualifies.
	steps := AnnouncedClauses(spec, nil, nil)
	bindStepsX(steps, 0)
	if !steps[0].Clause.xBoundAdmitsIn(g, battlefieldCardOf(t, g, hidden)) {
		t.Error("X=0: a face-down permanent (mana value 0) is not admitted")
	}
	if steps[0].Clause.xBoundAdmitsIn(g, battlefieldCardOf(t, g, faceUp)) {
		t.Error("X=0: a face-up mana value 6 creature is admitted")
	}
}
