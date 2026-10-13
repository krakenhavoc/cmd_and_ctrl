package effects

import "testing"

// #3088: the card-type tally reads the entering permanent through the
// same dry layer pass as the subtype tally (#2940). A Plains entering
// under Living Plane is a creature only through that static, and has not
// been through the real layer pass when the tally listener runs.
func TestEntryTallyCountsACardTypeGrantedByLivingPlane(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Living Plane", "World Enchantment", livingPlaneOracle, 0, 0)
	b13PlayAs(t, g, 0, "Plains", "Basic Land — Plains", "")
	g.WithWriteLock(func() {
		for _, ty := range []string{"land", "creature"} {
			if got := g.EnteredWithCardTypeThisTurn(me.ID, ty); got != 1 {
				t.Errorf("%s entered under Living Plane: tally %d, want 1", ty, got)
			}
		}
	})
}
