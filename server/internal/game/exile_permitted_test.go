package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// exile_permitted_test.go — #1573, ADR 0066's 2026-09-24 amendment.
// Two gaps in the "exile it and you may cast it" family:
//
//  1. a face-down exile whose one viewer is the permission holder
//     (FaceDownPermitted — Gonti, Night Minister; Outrageous Robbery);
//  2. "mana of any TYPE can be spent", which pays {C} as well as the
//     five colours (CastPermission.AnyType — Hostage Taker).

// seedTop puts a card with the given cost on top of p's library.
func seedTop(p *Player, name, typeLine, cost string) uuid.UUID {
	c := NewCard(name, p.ID)
	c.TypeLine = typeLine
	c.ManaCost = cost
	p.Library.PushTop(c)
	return c.InstanceID
}

// exileFaceDownPermitted runs the #1573 primitive under the lock.
func exileFaceDownPermitted(t *testing.T, g *Game, victim, thief *Player, n int, perm CastPermission) {
	t.Helper()
	var err error
	g.WithWriteLock(func() {
		err = g.ExileTopFaceDownWithPermissionForEffect(victim.ID, thief.ID, n, perm)
	})
	if err != nil {
		t.Fatalf("ExileTopFaceDownWithPermissionForEffect: %v", err)
	}
}

// pushUntappedLands puts n lands of the given type line on p's side,
// not summoning sick.
func pushUntappedLands(t *testing.T, g *Game, p *Player, n int, typeLine string) {
	t.Helper()
	for i := 0; i < n; i++ {
		atPush(t, g, p.ID, typeLine, typeLine, "")
	}
}

func whileExiled() CastPermission {
	return CastPermission{AnyType: true, Duration: WhileInZoneDuration()}
}

// Only the holder is a knower: the owner and the third seat get a card
// back, and the kind is `permitted`.
func TestFaceDownPermittedExileOnlyTheHolderMayLook(t *testing.T) {
	g := newActiveGameWithSeats(t, 3)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	opp.Library.Cards = nil
	bottom := seedTop(opp, "Second Card", "Instant", "{U}")
	top := seedTop(opp, "Stolen Seer", "Creature — Eldrazi", "{2}{C}")

	exileFaceDownPermitted(t, g, opp, me, 2, whileExiled())

	for _, id := range []uuid.UUID{top, bottom} {
		c := exiledCard(t, g, id)
		if !c.FaceDown || c.FaceDownKind != FaceDownPermitted {
			t.Errorf("%s: face_down %v kind %q, want face down %q", c.Name, c.FaceDown, c.FaceDownKind, FaceDownPermitted)
		}
		if !c.IsKnownTo(me.ID) {
			t.Errorf("%s: the permission holder may look at it", c.Name)
		}
		if c.IsKnownTo(opp.ID) || c.IsKnownTo(third.ID) {
			t.Errorf("%s: knowers %v — the owner and the third seat may not look", c.Name, c.KnownBy)
		}
		g.ReadSnapshot(func() {
			if g.CastPermissionForLocked(me.ID, *c, ZoneExile) == nil {
				t.Errorf("%s: the holder has no permission over it", c.Name)
			}
			if g.CastPermissionForLocked(opp.ID, *c, ZoneExile) != nil {
				t.Errorf("%s: the owner holds a permission over it", c.Name)
			}
		})
	}
	if opp.Library.Size() != 0 {
		t.Errorf("library has %d cards left, want both exiled", opp.Library.Size())
	}
}

// The holder casts it, and it turns face up on the way to the stack.
func TestFaceDownPermittedCardIsCastableByItsHolder(t *testing.T) {
	g := newActiveGameWithSeats(t, 3)
	me, opp := g.Seats[0], g.Seats[1]
	toMainPhase(t, g)
	opp.Library.Cards = nil
	loot := seedTop(opp, "Stolen Seer", "Creature — Eldrazi", "{2}{C}")
	exileFaceDownPermitted(t, g, opp, me, 1, whileExiled())
	pushUntappedLands(t, g, me, 3, "Basic Land — Swamp")

	if err := g.CastSpell(opp.ID, loot, CastSpellParams{FromZone: "exile"}); !errors.Is(err, ErrNoPlayPermission) {
		t.Errorf("the owner's cast: %v, want ErrNoPlayPermission", err)
	}
	if err := g.CastSpell(me.ID, loot, CastSpellParams{FromZone: "exile", AutoTap: true}); err != nil {
		t.Fatalf("the holder's cast: %v", err)
	}
	var onStack *Card
	for i := range g.Stack.Cards {
		if g.Stack.Cards[i].InstanceID == loot {
			onStack = &g.Stack.Cards[i]
		}
	}
	if onStack == nil {
		t.Fatal("the card is not on the stack")
	}
	if onStack.FaceDown {
		t.Error("the spell is still face down on the stack")
	}
	for _, p := range g.Seats {
		if !onStack.IsKnownTo(p.ID) {
			t.Errorf("seat %s cannot read the spell on the stack", p.Name)
		}
	}
	if item := g.StackMeta[loot]; item == nil || item.Controller != me.ID {
		t.Error("the holder should control the spell")
	}
}

// The grant stamps the look on a `permitted` card and on nothing else:
// Necropotence's `exiled` card stays unreadable to a permission that
// happens to name it.
func TestAGrantDoesNotOpenAPlainFaceDownExile(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	me.Library.Cards = nil
	id := seedTop(me, "Necro Card", "Instant", "{R}")
	g.WithWriteLock(func() {
		if _, err := g.ExileTopFaceDownForEffect(me.ID, 1); err != nil {
			t.Fatal(err)
		}
		g.GrantCastPermissionOverCardForEffect(id, CastPermission{Player: me.ID, Zone: ZoneExile})
	})
	if c := exiledCard(t, g, id); c.IsKnownTo(me.ID) {
		t.Errorf("a grant opened a plain face-down exile: knowers %v", c.KnownBy)
	}
}

// Undo takes the exile back whole, and restoring a later frame hands
// back the face-down card, its one viewer and the live permission.
func TestFaceDownPermittedSurvivesUndo(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	toMainPhase(t, g)
	opp.Library.Cards = nil
	loot := seedTop(opp, "Stolen Seer", "Creature — Eldrazi", "{2}{C}")

	beforeExile := g.Clone()
	exileFaceDownPermitted(t, g, opp, me, 1, whileExiled())
	afterExile := g.Clone()

	pushUntappedLands(t, g, me, 3, "Basic Land — Swamp")
	if err := g.CastSpell(me.ID, loot, CastSpellParams{FromZone: "exile", AutoTap: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}

	g.WithWriteLock(func() { g.RestoreFrom(afterExile) })
	me, opp = g.Seats[0], g.Seats[1]
	c := exiledCard(t, g, loot)
	if c.FaceDownKind != FaceDownPermitted || !c.IsKnownTo(me.ID) || c.IsKnownTo(opp.ID) {
		t.Errorf("after undoing the cast: kind %q knowers %v, want permitted and the holder alone", c.FaceDownKind, c.KnownBy)
	}
	g.ReadSnapshot(func() {
		if perm := g.CastPermissionForLocked(me.ID, *c, ZoneExile); perm == nil || !perm.AnyType {
			t.Errorf("after undoing the cast the permission is %+v, want the live any-type grant", perm)
		}
	})

	g.WithWriteLock(func() { g.RestoreFrom(beforeExile) })
	me, opp = g.Seats[0], g.Seats[1]
	if !opp.Library.Contains(loot) {
		t.Error("undoing the exile did not put the card back on the library")
	}
	if len(me.CastPermissions) != 0 {
		t.Errorf("undoing the exile left %d permissions behind", len(me.CastPermissions))
	}
}

// Kind, viewer and the any-type grant are carried state: a snapshot
// that dropped any of them would hand back a card nobody can read or
// one castable only with real colorless mana.
func TestFaceDownPermittedSurvivesASnapshotRoundTrip(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	opp.Library.Cards = nil
	loot := seedTop(opp, "Stolen Seer", "Creature — Eldrazi", "{2}{C}")
	exileFaceDownPermitted(t, g, opp, me, 1, whileExiled())

	_, restored := roundTrip(t, g)
	rme, ropp := restored.Seats[0], restored.Seats[1]
	c := exiledCard(t, restored, loot)
	if !c.FaceDown || c.FaceDownKind != FaceDownPermitted {
		t.Errorf("restored kind %q face_down %v, want permitted", c.FaceDownKind, c.FaceDown)
	}
	if !c.IsKnownTo(rme.ID) || c.IsKnownTo(ropp.ID) {
		t.Errorf("restored knowers %v, want the holder alone", c.KnownBy)
	}
	restored.ReadSnapshot(func() {
		if perm := restored.CastPermissionForLocked(rme.ID, *c, ZoneExile); perm == nil || !perm.AnyType {
			t.Errorf("restored permission %+v, want the any-type grant", perm)
		}
	})
}

func TestAsAnyTypeCostFoldsColorlessToo(t *testing.T) {
	cost, err := ParseCost("{1}{C}{W/U}{B}")
	if err != nil {
		t.Fatal(err)
	}
	folded := asAnyTypeCost(cost)
	if folded.Generic != 4 || len(folded.Required) != 0 {
		t.Errorf("{1}{C}{W/U}{B} → generic %d, required %d; want 4 and 0", folded.Generic, len(folded.Required))
	}
	if !(ManaPool{{Color: "R"}, {Color: "R"}, {Color: "R"}, {Color: "R"}}).CanPay(folded, 0) {
		t.Error("four red should pay it as any-type")
	}
	// spendAsThoughAny reads the wider clause first.
	if got := spendAsThoughAny(&CastPermission{AnyColor: true, AnyType: true}, cost); len(got.Required) != 0 {
		t.Errorf("AnyType+AnyColor left %d requirements, want the any-type fold", len(got.Required))
	}
	if got := spendAsThoughAny(&CastPermission{AnyColor: true}, cost); len(got.Required) != 1 {
		t.Errorf("AnyColor alone left %d requirements, want the {C}", len(got.Required))
	}
	if got := spendAsThoughAny(nil, cost); len(got.Required) != 3 {
		t.Errorf("no grant left %d requirements, want all three", len(got.Required))
	}
}

// anyTypeCastFixture exiles a {2}{C} creature off the opponent's
// library under `perm`, in the thief's main phase.
func anyTypeCastFixture(t *testing.T, perm CastPermission) (*Game, *Player, uuid.UUID) {
	t.Helper()
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	toMainPhase(t, g)
	opp.Library.Cards = nil
	seedTop(opp, "Stolen Seer", "Creature — Eldrazi", "{2}{C}")
	loot := impulseExile(t, g, opp, me, perm)
	return g, me, loot
}

// By hand: three black mana in the pool pays {2}{C} under any type,
// and does not under any colour.
func TestAnyTypePaysColorlessByHand(t *testing.T) {
	for _, tc := range []struct {
		name string
		perm CastPermission
		ok   bool
	}{
		{"any type", CastPermission{AnyType: true}, true},
		{"any color", CastPermission{AnyColor: true}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, loot := anyTypeCastFixture(t, tc.perm)
			me.ManaPool.AddMana(ManaToken{Color: "B"}, ManaToken{Color: "B"}, ManaToken{Color: "B"})
			err := g.CastSpell(me.ID, loot, CastSpellParams{FromZone: "exile", Strict: true})
			if tc.ok && err != nil {
				t.Fatalf("cast: %v", err)
			}
			if !tc.ok {
				var short *InsufficientManaError
				if !errors.As(err, &short) {
					t.Fatalf("cast: %v, want insufficient mana — {C} needs colorless", err)
				}
				return
			}
			if len(me.ManaPool) != 0 {
				t.Errorf("pool after the cast: %v, want empty", me.ManaPool)
			}
		})
	}
}

// Through the auto-tapper: three Swamps pay {2}{C} under any type,
// and not under any colour.
func TestAnyTypePaysColorlessThroughTheAutoTapper(t *testing.T) {
	for _, tc := range []struct {
		name string
		perm CastPermission
		ok   bool
	}{
		{"any type", CastPermission{AnyType: true}, true},
		{"any color", CastPermission{AnyColor: true}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, loot := anyTypeCastFixture(t, tc.perm)
			pushUntappedLands(t, g, me, 3, "Basic Land — Swamp")
			err := g.CastSpell(me.ID, loot, CastSpellParams{FromZone: "exile", AutoTap: true})
			if tc.ok && err != nil {
				t.Fatalf("cast: %v", err)
			}
			if !tc.ok && err == nil {
				t.Fatal("three Swamps paid {C} under an any-colour grant")
			}
			if tc.ok && !g.Stack.Contains(loot) {
				t.Error("the spell is not on the stack")
			}
		})
	}
}
