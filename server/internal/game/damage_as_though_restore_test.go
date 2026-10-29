package game

import (
	"slices"
	"testing"

	"github.com/google/uuid"
)

// damage_as_though_restore_test.go — #4389: the life totals a damage
// instance began with (Game.damageInstanceLives) are written to the
// snapshot, so Phyrexian Unlife reads the same total for a paused
// instance's later events after a restore as before it (CR 120.3b, 609.4).

// The records come back equal through JSON, and the restored instance
// counter starts past every instance they name.
func TestDamageInstanceLivesRoundTripThroughASnapshot(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	setLife(g, me, 3)
	setLife(g, opp, -1)
	var first, second DamageInstance
	g.WithWriteLock(func() { first = g.nextDamageInstanceLocked() })
	setLife(g, me, 0)
	g.WithWriteLock(func() { second = g.nextDamageInstanceLocked() })
	setLife(g, me, -5)
	setLife(g, opp, 4)

	r := strictRestore(t, g)
	if len(r.damageInstanceLives) != len(g.damageInstanceLives) {
		t.Fatalf("restored %d records, want %d", len(r.damageInstanceLives), len(g.damageInstanceLives))
	}
	for i, rec := range g.damageInstanceLives {
		got := r.damageInstanceLives[i]
		if got.inst != rec.inst || !slices.Equal(got.lives, rec.lives) {
			t.Fatalf("record %d restored %+v, want %+v", i, got, rec)
		}
	}
	r.WithWriteLock(func() {
		for _, c := range []struct {
			inst   DamageInstance
			player uuid.UUID
			want   bool
		}{
			{first, me.ID, false}, {first, opp.ID, true},
			{second, me.ID, true}, {second, opp.ID, true},
		} {
			if got := r.atOrBelowZeroLifeAsInstanceBeganLocked(c.inst, c.player); got != c.want {
				t.Errorf("instance %d, player %v: at or below zero %v, want %v", c.inst, c.player, got, c.want)
			}
		}
		if r.damageInstanceSeq < uint64(second) {
			t.Errorf("instance counter restored as %d, want at least %d", r.damageInstanceSeq, second)
		}
		if next := r.nextDamageInstanceLocked(); next <= second {
			t.Errorf("the restored game handed out instance %d again", next)
		}
	})
}

// A keyed Each walk paused on a CR 616 ordering prompt keeps its
// instance through a restore. Its first leg took Unlife's controller to
// 0 life, but the instance began at 2, so its last leg is still ordinary
// damage after the restore: life lost, no poison.
func TestUnlifeReadsTheInstancesLifeAfterARestore(t *testing.T) {
	withDamageAsThough(t, unlifeInfec)
	g := newActiveGameWithSeats(t, 3)
	src := pushKeywordCreature(t, g, g.Seats[0], 3, 3)
	me, p2 := g.Seats[1], g.Seats[2]
	pushAsThoughEnchantment(g, me)
	setLife(g, me, 2)
	// Two replacements for seat 2's damage: its controller orders them
	// (CR 616.1), which pauses the walk at its second leg.
	c := pushKeywordCreature(t, g, p2, 1, 20)
	g.WithWriteLock(func() {
		if !g.RedirectDamageThisTurnForEffect(DamageRedirection{Controller: p2.ID, ProtectPlayer: p2.ID, To: c, Label: "Saving Grace"}) {
			t.Fatal("no redirection")
		}
		if !g.PreventNextDamageThisTurnForEffect(uuid.Nil, p2.ID, 1, false, "Mending Hands") {
			t.Fatal("no shield")
		}
	})
	resetProbes()
	g.WithWriteLock(func() {
		if err := g.DealDamageEachKeyedForEffect(src, []uuid.UUID{me.ID, p2.ID, me.ID}, 2, probeTail()); err != nil {
			t.Fatal(err)
		}
	})
	if lifeOf(g, me.ID) != 0 || poisonOf(g, me.ID) != 0 {
		t.Fatalf("first leg: life %d poison %d, want 0 and 0", lifeOf(g, me.ID), poisonOf(g, me.ID))
	}
	if o := openReplacementOrder(g); o == nil || o.Chooser != p2.ID {
		t.Fatalf("ordering prompt %+v, want one for seat 2", o)
	}
	r := strictRestore(t, g)
	o := openReplacementOrder(r)
	if o == nil {
		t.Fatal("the restored game lost the ordering prompt")
	}
	if err := r.ResolveReplacementOrder(o.ID, o.Chooser, slices.Clone(o.ReplacementEffectIDs)); err != nil {
		t.Fatal(err)
	}
	if o := openReplacementOrder(r); o != nil {
		t.Fatalf("a second ordering prompt opened: %+v", o)
	}
	if lifeOf(r, me.ID) != -2 || poisonOf(r, me.ID) != 0 {
		t.Fatalf("last leg: life %d poison %d, want -2 and 0: the instance began at 2 life", lifeOf(r, me.ID), poisonOf(r, me.ID))
	}
}
