package game

import (
	"testing"

	"github.com/google/uuid"
)

// cast_tally_from_hand_test.go — #5235. CastTally.NotFromHand is "if you
// haven't cast a spell from your hand this turn" (Outlaws of Thunder
// Junction): the zone a spell was cast from, read before CR 601.2a moved
// it to the stack. The catalog half is cards/effects/cast_from_hand.go.

func castFromHandCount(g *Game, p *Player) int {
	var n int
	g.ReadSnapshot(func() { n = g.CastTallyFor(p.ID).FromHand() })
	return n
}

// castFromZoneForTest casts the card, already in `zone`, from there and
// resolves it.
func castFromZoneForTest(t *testing.T, g *Game, p *Player, id uuid.UUID, zone string) {
	t.Helper()
	if err := g.CastSpell(p.ID, id, CastSpellParams{FromZone: zone}); err != nil {
		t.Fatalf("cast from %s: %v", zone, err)
	}
	passPriorityUntilResolvedForTest(t, g, id)
}

// grantedIn puts a {0} sorcery in `zone` and gives its owner permission
// to cast it from there.
func grantedIn(t *testing.T, g *Game, p *Player, z *Zone, name string) uuid.UUID {
	t.Helper()
	c := typedCard(p.ID, name, "Sorcery")
	g.WithWriteLock(func() {
		z.PushTop(c)
		g.markCardKnownInZoneLocked(z, c.InstanceID)
		if !g.GrantCastPermissionOverCardForEffect(c.InstanceID, CastPermission{Player: p.ID, Zone: z.Kind}) {
			t.Fatalf("setup: no permission over %s in %s", name, z.Kind)
		}
	})
	return c.InstanceID
}

// A spell cast from the hand counts, and one from each other zone does
// not: a graveyard, exile, the command zone and the top of the library.
func TestCastFromHandCountsOnlyTheHand(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)

	castFromZoneForTest(t, g, me, grantedIn(t, g, me, me.Graveyard, "Fixture Flashback"), "graveyard")
	castFromZoneForTest(t, g, me, grantedIn(t, g, me, g.Exile, "Fixture Plotted"), "exile")
	castFromZoneForTest(t, g, me, grantedIn(t, g, me, me.Library, "Fixture Top"), "library")
	commander := typedCard(me.ID, "Fixture Commander", "Legendary Creature — Bear")
	g.WithWriteLock(func() { me.Command.PushTop(commander) })
	castFromZoneForTest(t, g, me, commander.InstanceID, "command")

	var tally CastTally
	g.ReadSnapshot(func() { tally = g.CastTallyFor(me.ID) })
	if tally.Total != 4 || tally.NotFromHand != 4 || tally.FromHand() != 0 {
		t.Fatalf("four casts from outside the hand: %+v (FromHand %d), want 4 not from the hand", tally, tally.FromHand())
	}

	castCardForTest(t, g, me, typedCard(me.ID, "Fixture Bolt", "Sorcery"), CastSpellParams{})
	if got := castFromHandCount(g, me); got != 1 {
		t.Fatalf("after a cast from the hand: FromHand %d, want 1", got)
	}
}

// A copy of a spell isn't cast (CR 707.10), so copying a spell cast from
// a hand adds nothing to the copier's tally; and an opponent's cast is in
// their tally, not yours.
func TestCastFromHandIgnoresCopiesAndOpponents(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	toMainPhase(t, g)

	spell := typedCard(opp.ID, "Fixture Bolt", "Instant")
	g.WithWriteLock(func() {
		opp.Hand.PushTop(spell)
		g.markCardKnownInZoneLocked(opp.Hand, spell.InstanceID)
	})
	if err := g.CastSpell(opp.ID, spell.InstanceID, CastSpellParams{}); err != nil {
		t.Fatalf("opponent casts: %v", err)
	}
	g.WithWriteLock(func() {
		if err := g.CopySpellForEffect(spell.InstanceID, me.ID, false, nil); err != nil {
			t.Fatalf("copy: %v", err)
		}
	})
	if top := g.Stack.Cards[len(g.Stack.Cards)-1]; top.InstanceID == spell.InstanceID || !g.StackMeta[top.InstanceID].IsCopy {
		t.Fatal("no copy on top of the stack")
	}
	if got := castFromHandCount(g, me); got != 0 {
		t.Fatalf("a copy or an opponent's spell counted for me: %d", got)
	}
	if got := castFromHandCount(g, opp); got != 1 {
		t.Fatalf("the opponent's own count: %d, want 1", got)
	}
}

func TestCastFromHandResetsNextTurn(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)
	castCardForTest(t, g, me, typedCard(me.ID, "Fixture Bolt", "Sorcery"), CastSpellParams{})
	if castFromHandCount(g, me) != 1 {
		t.Fatal("setup: the cast was not counted")
	}
	seat := g.Turn.ActiveSeat
	for i := 0; i < 40 && g.Turn.ActiveSeat == seat; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if g.Turn.ActiveSeat == seat {
		t.Fatal("setup: the turn never passed")
	}
	if got := castFromHandCount(g, me); got != 0 {
		t.Errorf("the count survived into the next turn: %d", got)
	}
}

// A restore taken mid-turn keeps both halves, through a JSON snapshot and
// a Clone (undo).
func TestCastFromHandSurvivesASnapshotRoundTripAndAClone(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)
	castFromZoneForTest(t, g, me, grantedIn(t, g, me, me.Graveyard, "Fixture Flashback"), "graveyard")

	_, restored := roundTrip(t, g)
	for name, back := range map[string]*Game{"restore": restored, "clone": g.Clone()} {
		var tally CastTally
		back.ReadSnapshot(func() { tally = back.CastTallyFor(me.ID) })
		if tally.Total != 1 || tally.FromHand() != 0 {
			t.Errorf("after the %s: %+v, want one cast, not from the hand", name, tally)
		}
	}
	castCardForTest(t, g, me, typedCard(me.ID, "Fixture Bolt", "Sorcery"), CastSpellParams{})
	_, back := roundTrip(t, g)
	if got := castFromHandCount(back, me); got != 1 {
		t.Errorf("after a cast from the hand and a restore: %d, want 1", got)
	}
}

// A restore point written before the field existed has no notFromHand:
// every spell of that turn reads as cast from the hand, so the condition
// is off (weaker, never stronger).
func TestCastFromHandReadsAnOlderTallyAsFromTheHand(t *testing.T) {
	if got := (CastTally{Total: 2}).FromHand(); got != 2 {
		t.Errorf("an older tally of two casts: FromHand %d, want 2", got)
	}
	if got := (CastTally{}).FromHand(); got != 0 {
		t.Errorf("no casts: FromHand %d, want 0", got)
	}
}
