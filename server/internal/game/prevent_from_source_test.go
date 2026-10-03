package game

import (
	"testing"

	"github.com/google/uuid"
)

// prevent_from_source_test.go — ADR 0108 §7 (#1904): the shield against a
// source that is not one-use (ModPreventFromSource), and Dark Sphere's
// half shield (Mod.Half on ModPreventNextFromSource).

// sourceShield registers a preventFromSource shield protecting `player`
// against the object `source` is right now.
func sourceShield(t *testing.T, g *Game, player, source uuid.UUID, amount int, then BodyRef, qs ...PermanentQuery) {
	t.Helper()
	g.WithWriteLock(func() {
		s := DamageShield{Controller: player, ProtectPlayer: player, Amount: amount, Then: then, Queries: qs, Label: "test source shield"}
		if source != uuid.Nil {
			ref, zone, ok := g.DamageSourceRefLocked(source)
			if !ok {
				t.Fatal("source in no zone")
			}
			s.Source, s.SourceZone = ref, zone
		}
		if !g.PreventDamageFromSourceThisTurnForEffect(s) {
			t.Fatal("no shield registered")
		}
	})
}

// "Prevent all damage a source of your choice would deal to you this
// turn" (Auriok Replica): every instance from that source this turn, not
// only the next one, and nothing from any other source.
func TestSourceShieldPreventsEveryInstanceThisTurn(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dragon := pushColouredCreature(g, opp, "Dragon", []string{"R"})
	other := pushColouredCreature(g, opp, "Other", []string{"R"})
	sourceShield(t, g, me.ID, dragon, 0, BodyRef{})
	start := lifeOf(g, me.ID)

	g.WithWriteLock(func() {
		for _, n := range []int{3, 4} {
			if err := g.DealDamageToPlayerForEffect(dragon, me.ID, n); err != nil {
				t.Fatal(err)
			}
		}
		g.beginEventBatchLocked()
		if err := g.DealDamageToPlayerForEffect(dragon, me.ID, 5); err != nil {
			t.Fatal(err)
		}
		if err := g.DealDamageToPlayerForEffect(other, me.ID, 1); err != nil {
			t.Fatal(err)
		}
	})
	if got := lifeOf(g, me.ID); got != start-1 {
		t.Fatalf("life %d, want %d: every instance from the dragon prevented, the other source's 1 dealt", got, start-1)
	}
	if n := len(g.ScopedEffects); n != 1 {
		t.Errorf("%d records, want the shield still live", n)
	}
}

// CR 615.9 / 609.7b: the property is rechecked as the damage would be
// dealt; a source that no longer matches gets through.
func TestSourceShieldRechecksItsProperty(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushColouredCreature(g, opp, "Shifter", []string{"R"})
	sourceShield(t, g, me.ID, src, 0, BodyRef{}, PermanentQuery{Colors: []string{"R"}})
	start := lifeOf(g, me.ID)
	g.WithWriteLock(func() {
		findBattlefieldCard(g, src).Colors = []string{"U"}
		g.layerVersion.Add(1)
		g.RecomputeLayersIfStaleLocked()
		_ = g.DealDamageToPlayerForEffect(src, me.ID, 2)
	})
	if got := lifeOf(g, me.ID); got != start-2 {
		t.Fatalf("life %d, want %d: a blue source fails the red recheck", got, start-2)
	}
}

// CR 615.7: "Prevent the next 3 damage that would be dealt to any target
// this turn by a source of your choice" (Healing Grace) prevents 3 in
// all, across instances, and is gone once spent.
func TestChargedSourceShieldPreventsTheNextN(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushColouredCreature(g, opp, "Pinger", []string{"R"})
	other := pushColouredCreature(g, opp, "Other", []string{"R"})
	sourceShield(t, g, me.ID, src, 3, BodyRef{})
	start := lifeOf(g, me.ID)
	g.WithWriteLock(func() {
		_ = g.DealDamageToPlayerForEffect(other, me.ID, 2)
		_ = g.DealDamageToPlayerForEffect(src, me.ID, 2)
	})
	if got := lifeOf(g, me.ID); got != start-2 {
		t.Fatalf("life %d, want %d: the other source's 2 dealt, the pinger's 2 prevented", got, start-2)
	}
	if got := g.ScopedEffects[0].Mods[0].Amount; got != 1 {
		t.Fatalf("charge %d left, want 1", got)
	}
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(src, me.ID, 4) })
	if got := lifeOf(g, me.ID); got != start-5 {
		t.Fatalf("life %d, want %d: 1 more prevented, 3 dealt", got, start-5)
	}
	if n := len(g.ScopedEffects); n != 0 {
		t.Errorf("%d records left, want the spent shield gone", n)
	}
}

// CR 615.5, 615.13: the follow-up runs once per instance, with what was
// prevented in it.
func TestSourceShieldFollowUpRunsOncePerInstance(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushColouredCreature(g, opp, "Dragon", []string{"R"})
	var calls []followUpCall
	sourceShield(t, g, me.ID, src, 0, testFollowUp(&calls))
	g.WithWriteLock(func() {
		_ = g.DealDamageToPlayerForEffect(src, me.ID, 3)
		_ = g.DealDamageToPlayerForEffect(src, me.ID, 2)
	})
	if len(calls) != 2 || calls[0].amount != 3 || calls[1].amount != 2 {
		t.Fatalf("follow-ups %+v, want 3 then 2", calls)
	}
	if calls[0].colors == nil || calls[0].colors[0] != "R" {
		t.Errorf("follow-up colours %v, want the source's red", calls[0].colors)
	}
}

// CR 615.12: unpreventable damage is dealt, the follow-up still runs with
// zero, and the charge is not reduced.
func TestChargedSourceShieldUnderUnpreventableDamage(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushColouredCreature(g, opp, "Dragon", []string{"R"})
	var calls []followUpCall
	sourceShield(t, g, me.ID, src, 3, testFollowUp(&calls))
	start := lifeOf(g, me.ID)
	g.WithWriteLock(func() {
		g.DamageCantBePreventedThisTurnForEffect(uuid.Nil, "Skullcrack")
		_ = g.DealDamageToPlayerForEffect(src, me.ID, 2)
	})
	if got := lifeOf(g, me.ID); got != start-2 {
		t.Fatalf("life %d, want %d", got, start-2)
	}
	if len(calls) != 1 || calls[0].amount != 0 {
		t.Fatalf("follow-ups %+v, want one with 0", calls)
	}
	for _, e := range g.ScopedEffects {
		for _, m := range e.Mods {
			if m.Kind == ModPreventFromSource && m.Amount != 3 {
				t.Errorf("charge %d, want 3: unpreventable damage does not reduce a shield", m.Amount)
			}
		}
	}
}

// "Prevent all combat damage that would be dealt by target creature"
// leaves the creature's other damage alone.
func TestCombatOnlySourceShield(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushColouredCreature(g, opp, "Attacker", []string{"G"})
	g.WithWriteLock(func() {
		ref, zone, _ := g.DamageSourceRefLocked(src)
		g.PreventDamageFromSourceThisTurnForEffect(DamageShield{Controller: me.ID, Source: ref, SourceZone: zone, CombatOnly: true, Label: "Maze"})
	})
	start := lifeOf(g, me.ID)
	g.WithWriteLock(func() {
		g.markCombatDamageToPlayerLocked(me.ID, src, 4, "")
		_ = g.DealDamageToPlayerForEffect(src, me.ID, 1)
	})
	if got := lifeOf(g, me.ID); got != start-1 {
		t.Fatalf("life %d, want %d: combat damage prevented, the ping dealt", got, start-1)
	}
}

// §7 decision 3: a charged shield with no source stays preventDamage.
func TestChargedSourceShieldNeedsASource(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() {
		if g.PreventDamageFromSourceThisTurnForEffect(DamageShield{Controller: me.ID, ProtectPlayer: me.ID, Amount: 3}) {
			t.Error("registered a charged preventFromSource with no source")
		}
	})
	for _, m := range []Mod{
		{Kind: ModPreventFromSource, Amount: 3},
		{Kind: ModPreventFromSource, Half: true, Objects: []ObjectRef{{ID: uuid.New()}}},
		{Kind: ModPreventFromSource, SourceZone: ZoneBattlefield},
		{Kind: ModPreventDamage, Amount: 1, Half: true},
	} {
		if nextFromSourceModProblem(m) == "" {
			t.Errorf("mod %+v passed the check", m)
		}
	}
}

// Dark Sphere: "prevent half that damage, rounded down" (CR 107.1a),
// spent by the instance. A single point is not halved at all, so the
// shield prevents nothing and is not used up (CR 609.7b).
func TestHalfShieldPreventsHalfRoundedDown(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushColouredCreature(g, opp, "Dragon", []string{"R"})
	var calls []followUpCall
	g.WithWriteLock(func() {
		ref, zone, _ := g.DamageSourceRefLocked(src)
		if !g.PreventNextDamageFromSourceForEffect(NextDamageShield{
			Controller: me.ID, Source: ref, SourceZone: zone, ProtectPlayer: me.ID, Half: true,
			Then: testFollowUp(&calls), Label: "Dark Sphere",
		}) {
			t.Fatal("no shield")
		}
	})
	start := lifeOf(g, me.ID)
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(src, me.ID, 1) })
	if got := lifeOf(g, me.ID); got != start-1 {
		t.Fatalf("life %d, want %d: half of 1 is none", got, start-1)
	}
	if g.ScopedEffects[0].Mods[0].SpentInstance != 0 {
		t.Fatal("a shield that prevented nothing was used up")
	}
	g.WithWriteLock(func() {
		_ = g.DealDamageToPlayerForEffect(src, me.ID, 5)
		_ = g.DealDamageToPlayerForEffect(src, me.ID, 4)
	})
	if got := lifeOf(g, me.ID); got != start-1-3-4 {
		t.Fatalf("life %d, want %d: 2 of the 5 prevented, the next instance dealt in full", got, start-8)
	}
	if len(calls) != 1 || calls[0].amount != 2 {
		t.Errorf("follow-ups %+v, want one with 2", calls)
	}
}
