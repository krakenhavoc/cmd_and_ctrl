package game

import (
	"testing"

	"github.com/google/uuid"
)

// airbend_test.go — S22: the two pieces airbend needed on top of
// #257's alternative costs. A permission with no expiry, and an
// exile-a-targeted-permanent primitive that grants to the card's
// OWNER.

// airbendGrant is the permission every airbend card hands out.
func airbendGrant() ExilePlayPermission {
	return ExilePlayPermission{CastOnly: true, WhileExiled: true, CostOverride: "{2}"}
}

// permanentFor puts a card on the battlefield under owner's control
// and returns its instance ID.
func permanentFor(g *Game, owner *Player, name, typeLine, manaCost string) uuid.UUID {
	c := NewCard(name, owner.ID)
	c.TypeLine = typeLine
	c.ManaCost = manaCost
	c.Controller = owner.ID
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

// exilePlayOf reads the grant off a card sitting in exile.
func exilePlayOf(g *Game, id uuid.UUID) ExilePlayPermission {
	for _, c := range g.Exile.Cards {
		if c.InstanceID == id {
			return c.ExilePlay
		}
	}
	return ExilePlayPermission{}
}

// An unbounded grant ignores the turn number entirely — that is the
// whole of piece 2.
func TestUnboundedExilePermissionIgnoresTheTurn(t *testing.T) {
	me := uuid.New()
	perm := ExilePlayPermission{Player: me, WhileExiled: true}
	for _, turn := range []int{1, 2, 50, 9999} {
		if !perm.Active(me, turn) {
			t.Errorf("turn %d: unbounded grant reported inactive", turn)
		}
	}
	// It is still a grant to ONE player. Unbounded is a duration,
	// not a licence.
	if perm.Active(uuid.New(), 1) {
		t.Errorf("unbounded grant admitted somebody it doesn't name")
	}
	// UntilTurn is not consulted, even when it is in the past.
	stale := ExilePlayPermission{Player: me, WhileExiled: true, UntilTurn: 1}
	if !stale.Active(me, 40) {
		t.Errorf("UntilTurn overrode WhileExiled")
	}
}

// The cleanup sweep reaps turn-bounded grants and must leave
// unbounded ones alone — a reaped airbend grant would strand the
// card in exile with the client still offering the cast.
func TestCleanupSpareUnboundedExilePermissions(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	opp.Library.Cards = nil
	seedLibraryTop(opp, "This Turn", "Instant")
	thisTurn := impulseExile(t, g, opp, me, ExilePlayPermission{})

	forever := permanentFor(g, opp, "Airbent Bear", "Creature — Bear", "{1}{G}")
	g.WithWriteLock(func() {
		if err := g.ExileCardWithPermissionForEffect(forever, airbendGrant()); err != nil {
			t.Fatalf("ExileCardWithPermissionForEffect: %v", err)
		}
	})

	g.WithWriteLock(func() { g.clearExpiredExilePlayLocked() })

	if exilePlayOf(g, thisTurn).Granted() {
		t.Errorf("the turn-bounded grant survived cleanup")
	}
	if got := exilePlayOf(g, forever); !got.Granted() || !got.WhileExiled {
		t.Errorf("cleanup reaped an unbounded grant: %+v", got)
	}
	// Still there several turns later.
	g.WithWriteLock(func() {
		g.Turn.Number += 5
		g.clearExpiredExilePlayLocked()
	})
	if !exilePlayOf(g, forever).Granted() {
		t.Errorf("an unbounded grant lapsed on a later cleanup")
	}
}

// The airbend primitive exiles a targeted battlefield permanent and
// grants to its OWNER — the direction that makes airbend removal
// with a refund rather than theft.
func TestExileCardWithPermissionGrantsToTheOwner(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := permanentFor(g, opp, "Their Bear", "Creature — Bear", "{1}{G}")

	var ltb bool
	g.WithWriteLock(func() {
		before := len(g.Events)
		if err := g.ExileCardWithPermissionForEffect(victim, airbendGrant()); err != nil {
			t.Fatalf("ExileCardWithPermissionForEffect: %v", err)
		}
		for _, ev := range g.Events[before:] {
			if ev.Kind == EventLTB && ev.CardID == victim && ev.NewZone == ZoneExile {
				ltb = true
			}
		}
	})

	if g.Battlefield.Contains(victim) || !g.Exile.Contains(victim) {
		t.Fatalf("the permanent did not move to exile")
	}
	if !ltb {
		t.Errorf("no leaves-the-battlefield event: airbend has to fire LTB triggers")
	}
	got := exilePlayOf(g, victim)
	if got.Player != opp.ID {
		t.Errorf("grant holder = %v, want the OWNER %v (not the exiler %v)", got.Player, opp.ID, me.ID)
	}
	if !got.WhileExiled || !got.CastOnly || got.CostOverride != "{2}" {
		t.Errorf("grant shape = %+v, want cast-only, unbounded, {2}", got)
	}
	// An explicit holder still wins — the primitive is not
	// airbend-only.
	other := permanentFor(g, opp, "Another Bear", "Creature — Bear", "{1}{G}")
	g.WithWriteLock(func() {
		_ = g.ExileCardWithPermissionForEffect(other, ExilePlayPermission{Player: me.ID})
	})
	if h := exilePlayOf(g, other).Player; h != me.ID {
		t.Errorf("explicit holder = %v, want %v", h, me.ID)
	}
	// …and a turn-bounded grant with no UntilTurn defaults to this
	// turn, matching the impulse primitive.
	if got := exilePlayOf(g, other); got.UntilTurn != g.Turn.Number {
		t.Errorf("UntilTurn = %d, want this turn %d", got.UntilTurn, g.Turn.Number)
	}
}

// The arithmetic test that matters, borrowed from ADR 0025: a pool
// of exactly {2} pays for an airbent card and ends EMPTY. Anything
// left over would mean the printed cost was charged instead.
func TestAirbendGrantChargesTheOverrideNotThePrintedCost(t *testing.T) {
	g := newActiveGame(t)
	// The owner is the active player: a creature spell is
	// sorcery-speed, and airbending your own board to rebuy it is
	// the mode the deck is built around anyway.
	me := g.Seats[0]
	toMainPhase(t, g)
	expensive := permanentFor(g, me, "Overpriced Bear", "Creature — Bear", "{5}{G}{G}")
	g.WithWriteLock(func() {
		if err := g.ExileCardWithPermissionForEffect(expensive, airbendGrant()); err != nil {
			t.Fatalf("airbend: %v", err)
		}
	})

	me.ManaPool.AddMana(ManaToken{Color: "C"}, ManaToken{Color: "C"})
	if err := g.CastSpell(me.ID, expensive, CastSpellParams{Strict: true, FromZone: "exile"}); err != nil {
		t.Fatalf("cast an airbent {5}{G}{G} for {2}: %v", err)
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("pool after the cast = %+v, want empty (a leftover means the printed cost ran)", me.ManaPool)
	}
	if !g.Stack.Contains(expensive) {
		t.Fatalf("the spell never reached the stack")
	}
	// The grant is spent as the card leaves exile, so a later
	// re-exile cannot inherit it.
	for _, c := range g.Stack.Cards {
		if c.InstanceID == expensive && c.ExilePlay.Granted() {
			t.Errorf("an unbounded grant survived the cast")
		}
	}
}

// One mana is not two: the override is a real cost, not a bypass.
func TestAirbendGrantStillNeedsTheTwo(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)
	bear := permanentFor(g, me, "Cheap Bear", "Creature — Bear", "{G}")
	g.WithWriteLock(func() {
		_ = g.ExileCardWithPermissionForEffect(bear, airbendGrant())
	})
	me.ManaPool.AddMana(ManaToken{Color: "G"})

	err := g.CastSpell(me.ID, bear, CastSpellParams{Strict: true, FromZone: "exile"})
	var im *InsufficientManaError
	if !errorsAs(err, &im) {
		t.Fatalf("one mana against a {2} override: got %v, want *InsufficientManaError", err)
	}
	if !g.Exile.Contains(bear) {
		t.Errorf("a rejected cast moved the card out of exile")
	}
}

// EventCast carries the zone the spell was cast from — the event
// half of #257's StackItem.CastFromZone, and the only way "whenever
// you cast a spell from exile" can be written.
func TestCastEventCarriesTheSourceZone(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)

	// Sorcery-speed first: a creature spell needs the stack empty,
	// and the instant below does not.
	bear := permanentFor(g, me, "Exiled Bear", "Creature — Bear", "{1}{G}")
	g.WithWriteLock(func() { _ = g.ExileCardWithPermissionForEffect(bear, airbendGrant()) })
	if err := g.CastSpell(me.ID, bear, CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("cast from exile: %v", err)
	}

	fromHand := NewCard("Hand Bolt", me.ID)
	fromHand.TypeLine = "Instant"
	me.Hand.PushTop(fromHand)
	if err := g.CastSpell(me.ID, fromHand.InstanceID, CastSpellParams{}); err != nil {
		t.Fatalf("cast from hand: %v", err)
	}

	zones := map[uuid.UUID]ZoneKind{}
	for _, ev := range g.Events {
		if ev.Kind == EventCast {
			if ev.NewZone != ZoneStack {
				t.Errorf("cast event NewZone = %q, want the stack", ev.NewZone)
			}
			zones[ev.CardID] = ev.OldZone
		}
	}
	if zones[fromHand.InstanceID] != ZoneHand {
		t.Errorf("hand cast OldZone = %q, want %q", zones[fromHand.InstanceID], ZoneHand)
	}
	if zones[bear] != ZoneExile {
		t.Errorf("exile cast OldZone = %q, want %q", zones[bear], ZoneExile)
	}
}

// "Becomes the target of a spell or ability" (CR 115.7) fires at
// announce, once per slot, for spells and for abilities alike.
func TestBecomesTargetFiresAtAnnounce(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	toMainPhase(t, g)
	one := permanentFor(g, opp, "Bear One", "Creature — Bear", "{1}{G}")
	two := permanentFor(g, opp, "Bear Two", "Creature — Bear", "{1}{G}")

	bolt := NewCard("Two-Target Bolt", me.ID)
	bolt.TypeLine = "Instant"
	me.Hand.PushTop(bolt)
	before := len(g.Events)
	if err := g.CastSpell(me.ID, bolt.InstanceID, CastSpellParams{
		Targets: []TargetRef{
			{Kind: TargetCard, ID: one},
			{Kind: TargetCard, ID: two},
			{Kind: TargetPlayer, ID: opp.ID},
		},
	}); err != nil {
		t.Fatalf("cast: %v", err)
	}

	var cards []uuid.UUID
	var players int
	for _, ev := range g.Events[before:] {
		if ev.Kind != EventBecomesTarget {
			continue
		}
		if ev.Actor != me.ID || ev.Source != bolt.InstanceID {
			t.Errorf("event attribution = actor %v source %v, want %v / %v",
				ev.Actor, ev.Source, me.ID, bolt.InstanceID)
		}
		if ev.CardID != uuid.Nil {
			cards = append(cards, ev.CardID)
			if ev.Target != ev.CardID {
				t.Errorf("card target: Target %v != CardID %v", ev.Target, ev.CardID)
			}
			continue
		}
		players++
		if ev.Target != opp.ID {
			t.Errorf("player target = %v, want %v", ev.Target, opp.ID)
		}
	}
	if len(cards) != 2 || cards[0] != one || cards[1] != two {
		t.Errorf("card target events = %v, want [%v %v] in slot order", cards, one, two)
	}
	if players != 1 {
		t.Errorf("player target events = %d, want 1", players)
	}
}
