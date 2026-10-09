package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// x_ceiling_test.go — #2581: a printed "X can't be greater than
// <count>" on a spell (CR 107.3a, 601.2b), at the announce gate.
//
// The catalog cannot be imported here, so the ceiling is installed
// through the CatalogXCeiling hook for a test oracle, with Winter's
// Chill's count: the number of snow lands the caster controls.

const testSnowXCeilingOracle = "test-x-ceiling-snow-lands"

func withSnowLandXCeiling(t *testing.T) {
	t.Helper()
	prev := CatalogXCeiling
	CatalogXCeiling = func(key string) *XCeiling {
		if key == testSnowXCeilingOracle {
			return &XCeiling{
				Label: "the number of snow lands you control",
				Count: func(g *Game, caster uuid.UUID) int {
					return g.CountControlledMatchingForEffect(caster, PermanentQuery{
						Types: []string{"land"}, Supertypes: []string{"snow"},
					})
				},
			}
		}
		if prev != nil {
			return prev(key)
		}
		return nil
	}
	t.Cleanup(func() { CatalogXCeiling = prev })
}

// snowCeilingSpell puts a {X}{U} instant with the ceiling in `p`'s hand.
func snowCeilingSpell(p *Player) uuid.UUID {
	c := NewCard("Test Chill", p.ID)
	c.TypeLine = "Instant"
	c.ManaCost = "{X}{U}"
	c.Layout = "normal"
	c.OracleID = testSnowXCeilingOracle
	p.Hand.PushTop(c)
	return c.InstanceID
}

func snowLandsFor(g *Game, p *Player, n int) []uuid.UUID {
	out := make([]uuid.UUID, n)
	for i := range out {
		out[i] = pushBattlefieldForTest(g, p.ID, "Snow-Covered Island", "Basic Snow Land — Island", "")
	}
	return out
}

func castWithX(g *Game, p *Player, id uuid.UUID, x int) error {
	return g.CastSpell(p.ID, id, CastSpellParams{XValue: x})
}

// With no snow land the ceiling is 0, so X = 0 is the only
// announcement.
func TestXCeilingOfZeroAllowsOnlyXZero(t *testing.T) {
	withSnowLandXCeiling(t)
	g := newActiveGame(t)
	me := g.Seats[0]
	advanceTo(t, g, StepPrecombatMain)
	// A land that is not snow does not count, nor does an opponent's
	// snow land.
	pushBattlefieldForTest(g, me.ID, "Island", "Basic Land — Island", "")
	snowLandsFor(g, g.Seats[1], 2)
	id := snowCeilingSpell(me)

	if err := castWithX(g, me, id, 1); !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("X=1 over a ceiling of 0: err = %v, want ErrInvalidParam", err)
	}
	if g.Stack.Contains(id) {
		t.Fatalf("a refused cast reached the stack")
	}
	if err := castWithX(g, me, id, 0); err != nil {
		t.Fatalf("X=0 under a ceiling of 0: %v", err)
	}
	if got := g.StackMeta[id].XValue; got != 0 {
		t.Errorf("XValue = %d, want 0", got)
	}
}

// X equal to the count is legal; one more is refused and nothing moves.
func TestXCeilingEqualToTheCountIsAcceptedAndOneMoreIsRefused(t *testing.T) {
	withSnowLandXCeiling(t)
	g := newActiveGame(t)
	me := g.Seats[0]
	advanceTo(t, g, StepPrecombatMain)
	snowLandsFor(g, me, 3)
	id := snowCeilingSpell(me)

	if got, ok := g.SpellXCeiling(me.ID, testSnowXCeilingOracle); !ok || got != 3 {
		t.Fatalf("SpellXCeiling = %d, %v; want 3, true", got, ok)
	}
	if err := castWithX(g, me, id, 4); !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("X=4 over a ceiling of 3: err = %v, want ErrInvalidParam", err)
	}
	if !me.Hand.Contains(id) {
		t.Fatalf("the refused cast left the hand")
	}
	if err := castWithX(g, me, id, 3); err != nil {
		t.Fatalf("X=3 at a ceiling of 3: %v", err)
	}
	if got := g.StackMeta[id].XValue; got != 3 {
		t.Errorf("XValue = %d, want 3", got)
	}
}

// CR 601.2b: the count is read as X is announced and never again. A
// snow land that leaves in response changes nothing about the X on the
// stack, and a snapshot restore keeps it too.
func TestXCeilingIsReadOnlyAtAnnounce(t *testing.T) {
	withSnowLandXCeiling(t)
	g := newActiveGame(t)
	me := g.Seats[0]
	advanceTo(t, g, StepPrecombatMain)
	snow := snowLandsFor(g, me, 2)
	id := snowCeilingSpell(me)
	if err := castWithX(g, me, id, 2); err != nil {
		t.Fatalf("X=2 at a ceiling of 2: %v", err)
	}

	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(snow[0]); err != nil {
			t.Fatalf("destroy a snow land: %v", err)
		}
	})
	if got, _ := g.SpellXCeiling(me.ID, testSnowXCeilingOracle); got != 1 {
		t.Fatalf("ceiling after a snow land left = %d, want 1", got)
	}
	if got := g.StackMeta[id].XValue; got != 2 {
		t.Errorf("XValue after the count fell = %d, want the announced 2", got)
	}

	_, restored := roundTrip(t, g)
	if got := restored.StackMeta[id].XValue; got != 2 {
		t.Errorf("XValue after a restore = %d, want 2", got)
	}
	// The ceiling is catalog data, not state: the restored game reads
	// the same count off its own board.
	if got, ok := restored.SpellXCeiling(me.ID, testSnowXCeilingOracle); !ok || got != 1 {
		t.Errorf("restored ceiling = %d, %v; want 1, true", got, ok)
	}
}

// CR 107.3b: a cast that pays neither the mana cost nor an alternative
// cost with X still has exactly one legal X, 0 — the ceiling does not
// make a larger one legal.
func TestXCeilingDoesNotOpenAFreeCastsX(t *testing.T) {
	withSnowLandXCeiling(t)
	g := newActiveGame(t)
	me := g.Seats[0]
	snowLandsFor(g, me, 2)
	c := NewCard("Test Chill", me.ID)
	c.TypeLine = "Instant"
	c.ManaCost = "{X}{U}"
	c.Layout = "normal"
	c.OracleID = testSnowXCeilingOracle
	id := exileWithGrant(t, g, c, freeCastGrant(me.ID))

	err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: "exile", XValue: 2})
	if !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("free cast at X=2: err = %v, want ErrInvalidParam", err)
	}
	if err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: "exile", XValue: 0}); err != nil {
		t.Fatalf("free cast at X=0: %v", err)
	}
}

// A card with no printed ceiling has none.
func TestNoXCeilingWithoutADeclaration(t *testing.T) {
	withSnowLandXCeiling(t)
	g := newActiveGame(t)
	if _, ok := g.SpellXCeiling(g.Seats[0].ID, "test-stroke"); ok {
		t.Errorf("an undeclared card reported a ceiling")
	}
}
