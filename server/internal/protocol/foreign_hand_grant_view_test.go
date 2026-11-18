package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// foreign_hand_grant_view_test.go — #5922. A standing permission over
// ANOTHER seat's hand (Sen Triplets: "you may play lands and cast spells
// from that player's hand this turn") reaches its holder's frame as
// `castable_here` on the cards in that hand, and nobody else's: not the
// hand's owner, whose own cast is the ordinary one, and not a bystander
// who merely sees the revealed hand.

func handCardNamed(t *testing.T, p *game.Player, name string) uuid.UUID {
	t.Helper()
	for _, c := range p.Hand.Cards {
		if c.Name == name {
			return c.InstanceID
		}
	}
	t.Fatalf("no %q in %s's hand", name, p.Name)
	return uuid.Nil
}

func TestAForeignHandRuleStampsCastableHereForItsHolderOnly(t *testing.T) {
	g := busyTable(t, 0)
	n := len(g.Seats)
	holder := g.Seats[g.Turn.ActiveSeat]
	owner := g.Seats[(g.Turn.ActiveSeat+1)%n]
	bystander := g.Seats[(g.Turn.ActiveSeat+2)%n]
	bears := handCardNamed(t, owner, "Grizzly Bears")
	forest := handCardNamed(t, owner, "Forest")
	g.WithWriteLock(func() {
		eot := g.UntilEndOfTurnDuration()
		g.GrantHandRevealedForEffect(owner.ID, "Sen Triplets", uuid.Nil, eot)
		g.GrantCastPermissionForEffect(game.CastPermission{
			Player: holder.ID, Zone: game.ZoneHand, ZoneOwner: owner.ID,
			Scope: game.ScopeStanding, Duration: eot, SourceName: "Sen Triplets",
		})
	})

	handOf := func(viewer uuid.UUID) ZoneView {
		return seatViewOf(t, ViewOfGameFor(g, viewer.String()), owner.ID).Hand
	}
	for _, id := range []uuid.UUID{bears, forest} {
		if c := cardInSeatZone(t, handOf(holder.ID), id); !c.CastableHere {
			t.Errorf("%s: castable_here is clear for the permission's holder", c.Name)
		}
		if c := cardInSeatZone(t, handOf(bystander.ID), id); c.CastableHere || c.Name == "" {
			t.Errorf("bystander's copy: name %q, castable_here %v; want the revealed card and no bit", c.Name, c.CastableHere)
		}
		if c := cardInSeatZone(t, handOf(owner.ID), id); c.CastableHere {
			t.Errorf("%s: the hand's owner carries castable_here on their own hand card", c.Name)
		}
	}
	// The holder's own hand is untouched by the rule over another's.
	mine := handCardNamed(t, holder, "Grizzly Bears")
	if c := cardInSeatZone(t, seatViewOf(t, ViewOfGameFor(g, holder.ID.String()), holder.ID).Hand, mine); c.CastableHere {
		t.Error("the holder's own hand card carries castable_here")
	}
}
