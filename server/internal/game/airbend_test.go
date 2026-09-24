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
func airbendGrant() CastPermission {
	return CastPermission{CastOnly: true, Duration: WhileInZoneDuration(), Cost: "{2}"}
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

// exilePlayOf reads the permission covering a card sitting in exile.
// Since ADR 0066 the permission lives on the PLAYER, so this asks the
// game rather than the card.
func exilePlayOf(g *Game, id uuid.UUID) *CastPermission {
	if perm := g.CastPermissionOnCardByIDForEffect(id); perm != nil {
		return perm
	}
	return &CastPermission{}
}

// A zone-bound grant ignores the turn entirely — that is the whole
// of piece 2, and since #945 it is Duration.WhileInZone saying so.
func TestUnboundedExilePermissionIgnoresTheTurn(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	perm := CastPermission{Player: me.ID, Duration: WhileInZoneDuration()}
	for _, turns := range []int{0, 1, 5, 40} {
		probe := g.Clone()
		probe.Seats[0].TurnsBegun += turns
		probe.Turn.Seq += turns
		probe.Turn.Round += turns
		var live bool
		probe.ReadSnapshot(func() { live = probe.CastPermissionActiveForEffect(&perm, me.ID) })
		if !live {
			t.Errorf("%d turns later: zone-bound grant reported inactive", turns)
		}
	}
	// It is still a grant to ONE player. Unbounded is a duration,
	// not a licence.
	var admitted bool
	g.ReadSnapshot(func() { admitted = g.CastPermissionActiveForEffect(&perm, g.Seats[1].ID) })
	if admitted {
		t.Errorf("zone-bound grant admitted somebody it doesn't name")
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
	thisTurn := impulseExile(t, g, opp, me, CastPermission{})

	forever := permanentFor(g, opp, "Airbent Bear", "Creature — Bear", "{1}{G}")
	g.WithWriteLock(func() {
		if err := g.ExileCardWithPermissionForEffect(forever, airbendGrant()); err != nil {
			t.Fatalf("ExileCardWithPermissionForEffect: %v", err)
		}
	})

	g.WithWriteLock(func() { g.sweepCastPermissionsLocked(true) })

	if reaped := exilePlayOf(g, thisTurn); reaped.Granted() {
		t.Errorf("the turn-bounded grant survived cleanup")
	}
	if got := exilePlayOf(g, forever); !got.Granted() || got.Duration.Kind != WhileInZone {
		t.Errorf("cleanup reaped an unbounded grant: %+v", got)
	}
	// Still there several turns later.
	g.WithWriteLock(func() {
		g.Turn.Seq += 5
		g.Turn.Round += 5
		for _, p := range g.Seats {
			p.TurnsBegun += 5
		}
		g.sweepCastPermissionsLocked(true)
	})
	if later := exilePlayOf(g, forever); !later.Granted() {
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
	if got.Duration.Kind != WhileInZone || !got.CastOnly || got.Cost != "{2}" {
		t.Errorf("grant shape = %+v, want cast-only, zone-bound, {2}", got)
	}
	// An explicit holder still wins — the primitive is not
	// airbend-only.
	other := permanentFor(g, opp, "Another Bear", "Creature — Bear", "{1}{G}")
	g.WithWriteLock(func() {
		_ = g.ExileCardWithPermissionForEffect(other, CastPermission{Player: me.ID})
	})
	if h := exilePlayOf(g, other).Player; h != me.ID {
		t.Errorf("explicit holder = %v, want %v", h, me.ID)
	}
	// …and a grant with no Duration defaults to "until end of turn",
	// stamped against this turn, matching the impulse primitive.
	got = exilePlayOf(g, other)
	if got.Duration.Kind != UntilEndOfTurn || got.Duration.Player != g.Seats[g.Turn.ActiveSeat].ID {
		t.Errorf("duration = %+v, want until end of this turn", got.Duration)
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
		if perm := g.CastPermissionOnCardByIDForEffect(c.InstanceID); c.InstanceID == expensive && perm.Granted() {
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

// "Becomes the target of a spell or ability" (CR 115.3) fires at
// announce, once per slot, for spells and for abilities alike — and
// for every VERB that announces one.
//
// Six sites finish choosing targets and all six call the one fan-out
// helper (emitBecameTargetLocked, events.go). The three rows below are
// the ones this package can announce with nothing but the sandbox
// verbs: the cast path, the manual activation (#968 — it stamped
// Targets onto the item and announced nothing) and the manual trigger
// announce. The other three — the catalog activation (activated.go), a
// trigger's CR 603.3d target pick (pending_choice.go) and a copy's
// re-target (spell_copy.go) — need a catalog card to announce and are
// not asserted here. What pins them is the helper itself: one call per
// site, so a site either announces or does not, which is exactly the
// shape of the bug this row was added for.
//
// Every row also DRAINS (#968 for the activation, #974 for the trigger
// announce): the announcer receives priority right after announcing,
// so CR 603.3b puts whatever the announcement triggered on the stack
// at that boundary rather than leaving it queued until something else
// happens to drain it — by which time the announced object has
// resolved and a ward trigger counters nothing. Nothing left in
// PendingTriggers is that drain, asserted from this package; the ward
// itself is a catalog card and is asserted in
// cards/effects/sandbox_announce_target_test.go.
func TestBecomesTargetFiresAtAnnounce(t *testing.T) {
	for _, tc := range []struct {
		name string
		// announce announces something at `targets` and returns the
		// Source the events must carry.
		announce func(t *testing.T, g *Game, me *Player, targets []TargetRef) uuid.UUID
	}{
		{"cast", func(t *testing.T, g *Game, me *Player, targets []TargetRef) uuid.UUID {
			bolt := NewCard("Two-Target Bolt", me.ID)
			bolt.TypeLine = "Instant"
			me.Hand.PushTop(bolt)
			if err := g.CastSpell(me.ID, bolt.InstanceID, CastSpellParams{Targets: targets}); err != nil {
				t.Fatalf("cast: %v", err)
			}
			return bolt.InstanceID
		}},
		{"sandbox activation", func(t *testing.T, g *Game, me *Player, targets []TargetRef) uuid.UUID {
			src := permanentFor(g, me, "Hand-Read Rock", "Artifact", "{2}")
			if err := g.ActivateAbility(me.ID, src, AbilityParams{
				Label: "{T}: point at three things", Targets: targets,
			}); err != nil {
				t.Fatalf("activate: %v", err)
			}
			return src
		}},
		{"sandbox trigger announce", func(t *testing.T, g *Game, me *Player, targets []TargetRef) uuid.UUID {
			src := permanentFor(g, me, "Hand-Read Enchantment", "Enchantment", "{2}")
			if err := g.AnnounceTrigger(me.ID, src, AbilityParams{
				Label: "when this triggers, point at three things", Targets: targets,
			}); err != nil {
				t.Fatalf("announce: %v", err)
			}
			return src
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			toMainPhase(t, g)
			one := permanentFor(g, opp, "Bear One", "Creature — Bear", "{1}{G}")
			two := permanentFor(g, opp, "Bear Two", "Creature — Bear", "{1}{G}")
			targets := []TargetRef{
				{Kind: TargetCard, ID: one},
				{Kind: TargetCard, ID: two},
				{Kind: TargetPlayer, ID: opp.ID},
			}

			before := len(g.Events)
			source := tc.announce(t, g, me, targets)

			var cards []uuid.UUID
			var players int
			for _, ev := range g.Events[before:] {
				if ev.Kind != EventBecomesTarget {
					continue
				}
				if ev.Actor != me.ID || ev.Source != source {
					t.Errorf("event attribution = actor %v source %v, want %v / %v",
						ev.Actor, ev.Source, me.ID, source)
				}
				if ev.StackItemID == uuid.Nil {
					t.Errorf("no StackItemID on the event; ward reads it to counter the right object")
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
			if len(g.PendingTriggers) != 0 {
				t.Errorf("%d triggers still queued after the announce — a became-target "+
					"trigger harvested here reaches the stack only after the announced "+
					"object has resolved", len(g.PendingTriggers))
			}
		})
	}
}
