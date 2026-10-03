package game

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
)

// redirect_damage_test.go — ADR 0108 §9 (#1905): damage dealt to
// something else instead (CR 614.9), the scoped ModRedirectDamage and the
// primitive every redirection goes through.

// redirect registers a redirection, failing the test when none is
// written. The source, when set, is pinned as the object it is now.
func redirect(t *testing.T, g *Game, r DamageRedirection, source uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if source != uuid.Nil {
			ref, zone, ok := g.DamageSourceRefLocked(source)
			if !ok {
				t.Fatal("source in no zone")
			}
			r.Source, r.SourceZone = ref, zone
		}
		if r.Label == "" {
			r.Label = "test redirection"
		}
		if !g.RedirectDamageThisTurnForEffect(r) {
			t.Fatal("no redirection registered")
		}
	})
}

// Beacon of Destiny: "The next time a source of your choice would deal
// damage to you this turn, that damage is dealt to this creature instead."
// The damage lands on the creature, through the permanent path; the
// player loses nothing; the next instance is dealt normally (CR 615.8).
func TestRedirectNextTimeToACreature(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	beacon := pushKeywordCreature(t, g, me, 1, 5)
	src := pushColouredCreature(g, opp, "Pinger", []string{"R"})
	redirect(t, g, DamageRedirection{Controller: me.ID, ProtectPlayer: me.ID, Next: true, To: beacon}, src)
	start := lifeOf(g, me.ID)

	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(src, me.ID, 3) })
	if got := lifeOf(g, me.ID); got != start {
		t.Fatalf("life %d, want %d: the damage was dealt to the Beacon instead", got, start)
	}
	if got := markedOn(g, beacon); got != 3 {
		t.Fatalf("Beacon has %d damage marked, want 3", got)
	}
	g.WithWriteLock(func() {
		g.beginEventBatchLocked()
		_ = g.DealDamageToPlayerForEffect(src, me.ID, 2)
	})
	if got := lifeOf(g, me.ID); got != start-2 {
		t.Fatalf("life %d, want %d: the next instance is dealt normally", got, start-2)
	}
	if n := len(g.ScopedEffects); n != 0 {
		t.Errorf("%d records, want the spent redirection gone", n)
	}
}

// Jade Monolith: damage to a creature dealt to a player instead lands
// through the player path — life loss, not marked damage.
func TestRedirectFromACreatureToAPlayer(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := pushKeywordCreature(t, g, me, 2, 2)
	src := pushColouredCreature(g, opp, "Pinger", []string{"R"})
	redirect(t, g, DamageRedirection{Controller: me.ID, ProtectPermanent: bear, Next: true, To: me.ID}, src)
	start := lifeOf(g, me.ID)
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(src, bear, 2) })
	if got := markedOn(g, bear); got != 0 {
		t.Fatalf("bear has %d damage, want none", got)
	}
	if got := lifeOf(g, me.ID); got != start-2 {
		t.Fatalf("life %d, want %d", got, start-2)
	}
}

// Reflect Damage: "that damage is dealt to that source's controller
// instead", read off the source as the damage would be dealt.
func TestRedirectToTheSourcesController(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushColouredCreature(g, opp, "Dragon", []string{"R"})
	redirect(t, g, DamageRedirection{Controller: me.ID, Next: true, ToSourceController: true}, src)
	mine, theirs := lifeOf(g, me.ID), lifeOf(g, opp.ID)
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(src, me.ID, 4) })
	if lifeOf(g, me.ID) != mine || lifeOf(g, opp.ID) != theirs-4 {
		t.Fatalf("life me %d (want %d), opp %d (want %d)", lifeOf(g, me.ID), mine, lifeOf(g, opp.ID), theirs-4)
	}
}

// CR 615.12 is not a redirection ban: damage that can't be prevented is
// still dealt to something else instead. "That damage can't be … dealt
// instead to another permanent or player" (Whippoorwill) is, and a
// "next time" redirection that could not apply is not used up
// (CR 609.7b).
func TestUnpreventableIsRedirectedAndUnredirectableIsNot(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	beacon := pushKeywordCreature(t, g, me, 1, 9)
	src := pushColouredCreature(g, opp, "Pinger", []string{"R"})
	redirect(t, g, DamageRedirection{Controller: me.ID, ProtectPlayer: me.ID, Next: true, To: beacon}, src)
	start := lifeOf(g, me.ID)
	g.WithWriteLock(func() {
		_ = g.DealMarkedDamageForEffect(src, nil, me.ID, 2, DamageMarks{CantBePrevented: true})
	})
	if lifeOf(g, me.ID) != start || markedOn(g, beacon) != 2 {
		t.Fatalf("unpreventable: life %d (want %d), beacon %d (want 2)", lifeOf(g, me.ID), start, markedOn(g, beacon))
	}

	g2 := newActiveGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	beacon2 := pushKeywordCreature(t, g2, me2, 1, 9)
	src2 := pushColouredCreature(g2, opp2, "Pinger", []string{"R"})
	redirect(t, g2, DamageRedirection{Controller: me2.ID, ProtectPlayer: me2.ID, Next: true, To: beacon2}, src2)
	start2 := lifeOf(g2, me2.ID)
	g2.WithWriteLock(func() {
		_ = g2.DealMarkedDamageForEffect(src2, nil, me2.ID, 2, DamageMarks{CantBeRedirected: true})
	})
	if lifeOf(g2, me2.ID) != start2-2 || markedOn(g2, beacon2) != 0 {
		t.Fatalf("unredirectable: life %d (want %d), beacon %d (want 0)", lifeOf(g2, me2.ID), start2-2, markedOn(g2, beacon2))
	}
	if g2.ScopedEffects[0].Mods[0].SpentInstance != 0 {
		t.Fatal("a redirection that could not apply was used up (CR 609.7b)")
	}
	g2.WithWriteLock(func() {
		g2.beginEventBatchLocked()
		_ = g2.DealDamageToPlayerForEffect(src2, me2.ID, 3)
	})
	if markedOn(g2, beacon2) != 3 {
		t.Fatalf("beacon %d, want the next ordinary instance redirected", markedOn(g2, beacon2))
	}
}

// Whippoorwill's grant pins the creature: damage dealt TO it can't be
// redirected away.
func TestPinnedCantBeRedirectedStopsARedirection(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := pushKeywordCreature(t, g, me, 2, 2)
	src := pushColouredCreature(g, opp, "Pinger", []string{"R"})
	redirect(t, g, DamageRedirection{Controller: me.ID, ProtectPermanent: bear, To: me.ID}, src)
	g.WithWriteLock(func() {
		if !g.DamageToCantBePreventedThisTurnForEffect(uuid.Nil, bear, true, "Whippoorwill") {
			t.Fatal("no grant")
		}
	})
	start := lifeOf(g, me.ID)
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(src, bear, 1) })
	if markedOn(g, bear) != 1 || lifeOf(g, me.ID) != start {
		t.Fatalf("bear %d (want 1), life %d (want %d)", markedOn(g, bear), lifeOf(g, me.ID), start)
	}
}

// Harm's Way: "The next 2 damage … is dealt to any target instead." A
// charge smaller than the event splits it: 2 to the destination, the rest
// where it was, and the charge is gone (CR 615.7).
func TestChargedRedirectionSplitsAnEvent(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushColouredCreature(g, opp, "Dragon", []string{"R"})
	redirect(t, g, DamageRedirection{Controller: me.ID, ProtectPlayer: me.ID, Amount: 2, To: opp.ID}, src)
	mine, theirs := lifeOf(g, me.ID), lifeOf(g, opp.ID)
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(src, me.ID, 5) })
	if lifeOf(g, me.ID) != mine-3 || lifeOf(g, opp.ID) != theirs-2 {
		t.Fatalf("life me %d (want %d), opp %d (want %d)", lifeOf(g, me.ID), mine-3, lifeOf(g, opp.ID), theirs-2)
	}
	if n := len(g.ScopedEffects); n != 0 {
		t.Errorf("%d records, want the spent charge gone", n)
	}
}

// CR 614.5: the split-off part keeps what was applied to the event. A
// doubler applied before the redirection does not double the part again.
func TestRedirectedPartIsNotReplacedAgain(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushColouredCreature(g, opp, "Dragon", []string{"R"})
	g.RegisterReplacementForTest(ReplacementEffect{
		Watches:    []EventKind{EventDealDamage},
		Preemptive: true,
		AppliesTo:  func(ev *ReplacementEvent, _ *Game, _ *Card) bool { return ev.DamageAmount > 0 },
		Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
			ev.DamageAmount *= 2
			return nil
		},
		Label: "test doubler",
	})
	redirect(t, g, DamageRedirection{Controller: me.ID, ProtectPlayer: me.ID, Amount: 2, To: opp.ID}, src)
	mine, theirs := lifeOf(g, me.ID), lifeOf(g, opp.ID)
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(src, me.ID, 3) })
	if lifeOf(g, me.ID) != mine-4 || lifeOf(g, opp.ID) != theirs-2 {
		t.Fatalf("life me %d (want %d), opp %d (want %d): 3 doubled to 6, 2 of it redirected and not doubled again",
			lifeOf(g, me.ID), mine-4, lifeOf(g, opp.ID), theirs-2)
	}
}

// CR 616.2: the redirected damage meets the new recipient's shield.
func TestRedirectedDamageMeetsTheNewRecipientsShield(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	beacon := pushKeywordCreature(t, g, me, 1, 5)
	src := pushColouredCreature(g, opp, "Pinger", []string{"R"})
	redirect(t, g, DamageRedirection{Controller: me.ID, ProtectPlayer: me.ID, Next: true, To: beacon}, src)
	g.WithWriteLock(func() {
		if !g.PreventNextDamageThisTurnForEffect(uuid.Nil, beacon, 2, false, "Mending Hands") {
			t.Fatal("no shield")
		}
	})
	start := lifeOf(g, me.ID)
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(src, me.ID, 3) })
	if lifeOf(g, me.ID) != start || markedOn(g, beacon) != 1 {
		t.Fatalf("life %d (want %d), beacon %d (want 1: 2 of the redirected 3 prevented)", lifeOf(g, me.ID), start, markedOn(g, beacon))
	}
}

// CR 614.9: a destination that has left the battlefield makes the
// redirection do nothing, and it is not used up (CR 609.7b).
func TestRedirectionToAGoneCreatureDoesNothing(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	beacon := pushKeywordCreature(t, g, me, 1, 5)
	src := pushColouredCreature(g, opp, "Pinger", []string{"R"})
	redirect(t, g, DamageRedirection{Controller: me.ID, ProtectPlayer: me.ID, Next: true, To: beacon}, src)
	if err := g.MoveCardByID(ZoneRef{Kind: ZoneBattlefield}, ZoneRef{Kind: ZoneGraveyard, Owner: me.ID}, beacon); err != nil {
		t.Fatal(err)
	}
	start := lifeOf(g, me.ID)
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(src, me.ID, 3) })
	if got := lifeOf(g, me.ID); got != start-3 {
		t.Fatalf("life %d, want %d: nothing to redirect to", got, start-3)
	}
}

// CR 120.4b: the redirected damage keeps its source's deathtouch.
func TestRedirectedDamageKeepsDeathtouch(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	big := pushKeywordCreature(t, g, me, 1, 9)
	src := pushKeywordCreature(t, g, opp, 1, 1, "deathtouch")
	redirect(t, g, DamageRedirection{Controller: me.ID, ProtectPlayer: me.ID, Next: true, To: big}, src)
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(src, me.ID, 1) })
	var lethal bool
	g.WithWriteLock(func() { lethal = findBattlefieldCard(g, big).MarkedLethalByDeathtouch })
	if !lethal {
		t.Fatal("the redirected damage lost its source's deathtouch")
	}
}

// Eye for an Eye's shape: no destination; the event is dealt as it was
// and the follow-up runs with the amount.
func TestRedirectionWithNoDestinationOwesItsFollowUp(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushColouredCreature(g, opp, "Dragon", []string{"R"})
	var amounts []int
	key := fmt.Sprintf("%seye-for-an-eye-%d", testEffectKeyPrefix, testEffectKeySeq.Add(1))
	body := DelayedBody(key, func(g *Game, item *StackItem, p EffectParams) error {
		amounts = append(amounts, item.Trigger.Event.Amount)
		return nil
	})
	redirect(t, g, DamageRedirection{Controller: me.ID, ProtectPlayer: me.ID, Next: true, Then: body}, src)
	start := lifeOf(g, me.ID)
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(src, me.ID, 4) })
	if got := lifeOf(g, me.ID); got != start-4 {
		t.Fatalf("life %d, want %d: the source still deals the damage to you", got, start-4)
	}
	if len(amounts) != 1 || amounts[0] != 4 {
		t.Fatalf("follow-ups %v, want one with 4", amounts)
	}
}

// Owner decision 1, the Harm's Way ruling: a charged redirection meeting
// two attackers' damage for more than its charge is divided by the
// protected player before any of it is dealt.
func TestChargedRedirectionIsDivided(t *testing.T) {
	g := newActiveGame(t)
	atk, def := g.Seats[0], g.Seats[1]
	a := pushKeywordCreature(t, g, atk, 3, 3)
	b := pushKeywordCreature(t, g, atk, 3, 3)
	g.WithWriteLock(func() {
		if !g.RedirectDamageThisTurnForEffect(DamageRedirection{Controller: def.ID, ProtectPlayer: def.ID, Amount: 2, To: atk.ID, Label: "Harm's Way"}) {
			t.Fatal("no redirection")
		}
	})
	defStart, atkStart := lifeOf(g, def.ID), lifeOf(g, atk.ID)
	advanceIntoStep(t, g, StepDeclareAttackers)
	if err := g.DeclareAttacker(a, def.ID); err != nil {
		t.Fatal(err)
	}
	if err := g.DeclareAttacker(b, def.ID); err != nil {
		t.Fatal(err)
	}
	advanceIntoStep(t, g, StepCombatDamage)
	c := openDivideShield(g)
	if c == nil {
		t.Fatal("no divide_shield prompt for 6 damage against a 2-point redirection")
	}
	if c.Chooser != def.ID || c.DivideShield.Charge != 2 {
		t.Fatalf("prompt chooser %v charge %d; want the defender dividing 2", c.Chooser, c.DivideShield.Charge)
	}
	if err := g.ResolveDivideShield(c.ID, def.ID, shareFor(c, map[uuid.UUID]int{a: 1, b: 1})); err != nil {
		t.Fatal(err)
	}
	if lifeOf(g, def.ID) != defStart-4 || lifeOf(g, atk.ID) != atkStart-2 {
		t.Fatalf("defender %d (want %d), attacker %d (want %d)", lifeOf(g, def.ID), defStart-4, lifeOf(g, atk.ID), atkStart-2)
	}
}

// The registration refuses a shape restore would refuse, and every other
// kind refuses toSourceController (a newer binary's field).
func TestRedirectDamageModProblems(t *testing.T) {
	src := ObjectRef{ID: uuid.New(), Epoch: 1}
	to := []ObjectRef{{ID: uuid.New()}}
	for _, c := range []struct {
		m    Mod
		want bool
	}{
		{Mod{Kind: ModRedirectDamage, To: to}, false},
		{Mod{Kind: ModRedirectDamage, ToSourceController: true, Objects: []ObjectRef{src}, SourceZone: ZoneBattlefield, Next: true}, false},
		{Mod{Kind: ModRedirectDamage}, true},
		{Mod{Kind: ModRedirectDamage, To: to, ToSourceController: true}, true},
		{Mod{Kind: ModRedirectDamage, To: to, Next: true, Amount: 2}, true},
		{Mod{Kind: ModRedirectDamage, To: to, SpentInstance: 3}, true},
		{Mod{Kind: ModRedirectDamage, To: to, Half: true}, true},
		{Mod{Kind: ModPreventFromSource, ToSourceController: true}, true},
	} {
		if got := redirectDamageModProblem(c.m) != ""; got != c.want {
			t.Errorf("%+v: problem %v, want %v (%q)", c.m, got, c.want, redirectDamageModProblem(c.m))
		}
	}
	for _, m := range []Mod{
		{Kind: ModRedirectDamage, To: to, Next: true, Objects: []ObjectRef{src}, SourceZone: ZoneStack, Queries: []PermanentQuery{{Types: []string{"Instant"}}}},
	} {
		for _, check := range []func(Mod) string{nextFromSourceModProblem, multiplyDamageModProblem, followUpModProblem, blockRequirementModProblem} {
			if p := check(m); p != "" {
				t.Errorf("a sound redirection refused: %s", p)
			}
		}
	}
}
