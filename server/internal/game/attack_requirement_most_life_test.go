package game

import "testing"

// attack_requirement_most_life_test.go — #2744: which attacks obey
// "attacks an opponent with the most life among your opponents".

func TestMostLifeOpponentRequirementReadsLiveOpponentsOnly(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, a, b, c := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	r := AttackRequirement{MostLifeOpponentOf: me.ID}
	me.Life, a.Life, b.Life, c.Life = 99, 30, 25, 50
	c.Eliminated = true

	g.WithWriteLock(func() {
		if !r.obeyedBy(g, a.ID) {
			t.Error("the living opponent with the most life does not obey it")
		}
		if r.obeyedBy(g, b.ID) {
			t.Error("an opponent with less life obeys it")
		}
		if r.obeyedBy(g, c.ID) {
			t.Error("an eliminated seat with more life obeys it")
		}
		if r.obeyedBy(g, me.ID) {
			t.Error("you are not your own opponent, whatever your life")
		}
		b.Life = 30
		if !r.obeyedBy(g, b.ID) {
			t.Error("an opponent tied for the most life does not obey it")
		}
	})
}
