package game

import (
	"testing"

	"github.com/google/uuid"
)

// face_down_lifecycle_test.go — #697, and ADR 0069 decision 5.
//
// Two bugs and one rule. The rule is that "face down" belongs to an
// OBJECT in a zone, so a card that changes zones loses it (CR 400.7);
// the bugs are the two paths that reached MoveCard without the reset
// its callers were each doing for themselves.
//
// Every assertion here fails if the one line is backed out of
// MoveCard, including the row for the exit route that was already
// correct — which is the point of putting the reset in the primitive
// rather than in the callers.

// TestMoveCardClearsTheFaceDownFlagOnEveryPath is #697's first bug.
// Each subtest is a path that reaches MoveCard, and the assertion is
// the same for all of them because the reset is one line in the
// primitive rather than one line per caller. Back that line out and
// the admin-move and cast rows fail.
func TestMoveCardClearsTheFaceDownFlagOnEveryPath(t *testing.T) {
	for _, tc := range []struct {
		name string
		dst  ZoneKind
		move func(g *Game, id uuid.UUID) error
	}{
		{
			// The live bug: a sandbox move_card on a Necropotence
			// exile landed the card face down in a hand.
			name: "admin move_card to hand", dst: ZoneHand,
			move: func(g *Game, id uuid.UUID) error {
				return g.MoveCardByID(
					ZoneRef{Kind: ZoneExile},
					ZoneRef{Kind: ZoneHand, Owner: g.Seats[0].ID}, id)
			},
		},
		{
			name: "admin move_card to graveyard", dst: ZoneGraveyard,
			move: func(g *Game, id uuid.UUID) error {
				return g.MoveCardByID(
					ZoneRef{Kind: ZoneExile},
					ZoneRef{Kind: ZoneGraveyard, Owner: g.Seats[0].ID}, id)
			},
		},
		{
			// The battlefield destination is the ENTRY branch of the
			// sandbox move, which never reached zone_route's reset.
			name: "admin move_card to battlefield", dst: ZoneBattlefield,
			move: func(g *Game, id uuid.UUID) error {
				return g.MoveCardByID(
					ZoneRef{Kind: ZoneExile},
					ZoneRef{Kind: ZoneBattlefield}, id)
			},
		},
		{
			// Already correct before #697 because it goes through the
			// shared exit route; here so the route keeps working after
			// its own clear was deleted.
			name: "the shared exit route (bounce)", dst: ZoneHand,
			move: func(g *Game, id uuid.UUID) error {
				var err error
				g.WithWriteLock(func() { err = g.BounceToHandForEffect(id) })
				return err
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, id := faceDownGame(t, "Griselbrand", "Legendary Creature — Demon", "oracle-grissy")
			me := g.Seats[0]
			g.WithWriteLock(func() {
				if _, err := g.ExileTopFaceDownForEffect(me.ID, 1); err != nil {
					t.Fatalf("exile face down: %v", err)
				}
			})
			if c, _, _ := cardAnywhere(g, id); !c.FaceDown {
				t.Fatal("fixture: the card is not face down, so the reset proves nothing")
			}
			if err := tc.move(g, id); err != nil {
				t.Fatalf("move: %v", err)
			}
			c, zone, ok := cardAnywhere(g, id)
			if !ok {
				t.Fatal("card vanished")
			}
			if zone != tc.dst {
				t.Fatalf("card landed in %q, want %q", zone, tc.dst)
			}
			if c.FaceDown || c.FaceDownKind != FaceDownNone {
				t.Errorf("CR 400.7: the card is in %q still marked face down (%q) — #697",
					zone, c.FaceDownKind)
			}
		})
	}
}

// TestCastFromExileTurnsTheCardFaceUp is #697's second, latent half and
// foretell's own path: CR 406.3a turns a face-down card face up just
// before it is cast. Nothing in the catalog can cast a face-down
// exiled card today, so the cast grant is built by hand — which is
// exactly the shape #658 will produce.
func TestCastFromExileTurnsTheCardFaceUp(t *testing.T) {
	g, id := faceDownGame(t, "Lightning Bolt", "Instant", "oracle-bolt")
	me := g.Seats[0]
	g.WithWriteLock(func() {
		if _, err := g.routeCardToZoneLocked(zoneRoute{
			CardID: id, Dst: ZoneExile, Actor: me.ID, FaceDown: FaceDownForetold,
		}); err != nil {
			t.Fatalf("route: %v", err)
		}
		g.GrantCastPermissionOverCardForEffect(id, CastPermission{Player: me.ID, Duration: WhileInZoneDuration()})
	})
	if err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: string(ZoneExile)}); err != nil {
		t.Fatalf("cast from exile: %v", err)
	}
	c, zone, ok := cardAnywhere(g, id)
	if !ok || zone != ZoneStack {
		t.Fatalf("card is in %q, want the stack", zone)
	}
	if c.FaceDown {
		t.Error("CR 406.3a: a card cast out of a face-down exile is still face down — #697")
	}
}
