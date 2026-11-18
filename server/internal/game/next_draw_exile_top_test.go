package game

import (
	"testing"

	"github.com/google/uuid"
)

// next_draw_exile_top_test.go — #5875: the one-shot "the next time they
// would draw a card this turn, instead they exile the top card of their
// library; they may play it this turn" (next_draw_exile_top.go). The card
// that makes it, Urabrask, Heretic Praetor, is tested end to end in
// cards/effects.

func nextDrawRecords(g *Game, player uuid.UUID) int {
	n := 0
	g.ReadSnapshot(func() {
		for _, e := range g.ScopedEffects {
			for _, m := range e.Mods {
				if m.Kind == ModNextDrawExileTopPlay && m.Player == player {
					n++
				}
			}
		}
	})
	return n
}

// The record survives a restore point, and the restored binary replaces
// the first draw of a "draw two" (the drawer may play the card, the
// record's controller may not) and draws the second.
func TestNextDrawExileTopSurvivesARestoreAndReplacesOneDraw(t *testing.T) {
	g := newActiveGame(t)
	you, them := g.Seats[0], g.Seats[1]
	var ok bool
	g.WithWriteLock(func() { ok = g.NextDrawExileTopPlayableForEffect(uuid.New(), you.ID, them.ID, "test") })
	if !ok || nextDrawRecords(g, them.ID) != 1 {
		t.Fatal("no record was written")
	}
	if g.NextDrawExileTopPlayableForEffect(uuid.New(), you.ID, uuid.Nil, "test") {
		t.Error("a record naming no player was written")
	}

	_, r := roundTrip(t, g)
	p := r.playerByIDLocked(them.ID)
	second, first := uuid.New(), uuid.New()
	p.Library.PushTop(Card{InstanceID: second, Name: "Second", TypeLine: "Creature — Test", Owner: p.ID, Controller: p.ID})
	p.Library.PushTop(Card{InstanceID: first, Name: "First", TypeLine: "Creature — Test", Owner: p.ID, Controller: p.ID})
	handBefore := p.Hand.Size()

	var err error
	r.WithWriteLock(func() { err = r.DrawNForEffect(p.ID, 2) })
	if err != nil {
		t.Fatalf("draw two: %v", err)
	}
	if !r.Exile.Contains(first) || p.Hand.Contains(first) {
		t.Fatal("the first draw was not replaced by an exile")
	}
	if !p.Hand.Contains(second) || p.Hand.Size() != handBefore+1 {
		t.Error("the second draw was not an ordinary draw")
	}
	var mine, yours *CastPermission
	r.ReadSnapshot(func() {
		c, _ := r.LookupCardForEffect(first)
		mine = r.CastPermissionForLocked(p.ID, c, ZoneExile)
		yours = r.CastPermissionForLocked(you.ID, c, ZoneExile)
	})
	if mine == nil {
		t.Error("the drawer may not play the exiled card")
	}
	if yours != nil {
		t.Error("the record's controller may play the drawer's card")
	}
	if nextDrawRecords(r, them.ID) != 0 {
		t.Error("the record was not spent")
	}
}

// A draw by another player does not touch the record.
func TestNextDrawExileTopIgnoresOtherPlayersDraws(t *testing.T) {
	g := newActiveGame(t)
	you, them := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() { g.NextDrawExileTopPlayableForEffect(uuid.New(), you.ID, them.ID, "test") })
	top := uuid.New()
	you.Library.PushTop(Card{InstanceID: top, Name: "Mine", TypeLine: "Creature — Test", Owner: you.ID, Controller: you.ID})
	g.WithWriteLock(func() { _ = g.DrawNForEffect(you.ID, 1) })
	if !you.Hand.Contains(top) {
		t.Error("another player's draw was replaced")
	}
	if nextDrawRecords(g, them.ID) != 1 {
		t.Error("another player's draw spent the record")
	}
}
