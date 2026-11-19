package game

import (
	"testing"

	"github.com/google/uuid"
)

// player_shroud_test.go — #5842, CR 702.18a on a PLAYER: "This
// permanent or player can't be the target of spells or abilities."
// Gilded Light is the card; these are the engine half, written from
// both sides of the line hexproof draws (CR 702.11c) and shroud does
// not: the opponent's spell is refused, and so is the player's own.

// grantShroud is Gilded Light's clause, applied directly: shroud until
// end of turn.
func grantShroud(g *Game, p *Player) {
	g.WithWriteLock(func() {
		g.GrantPlayerStaticForEffect(p.ID, KeywordShroud, "Test — shroud", uuid.Nil, Duration{})
	})
}

// TestShroudedPlayerRefusesEveryoneThemselvesIncluded is the headline:
// unlike hexproof, shroud asks neither who nor what.
func TestShroudedPlayerRefusesEveryoneThemselvesIncluded(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)

	const bolt = "test-player-shroud-bolt"
	withCatalogTargetSpec(t, func(id string) *TargetSpec {
		if id == bolt {
			return anyPlayerSpec()
		}
		return nil
	})

	grantShroud(g, me)

	if err := castAtPlayer(t, g, me, bolt, []string{"R"}, me.ID); err != ErrIllegalTarget {
		t.Fatalf("your own spell at yourself under shroud: got %v, want ErrIllegalTarget", err)
	}
	if err := castAtPlayer(t, g, opp, bolt, []string{"R"}, me.ID); err != ErrIllegalTarget {
		t.Fatalf("an opponent's spell at a shrouded player: got %v, want ErrIllegalTarget", err)
	}
	// Shroud is the player's own: the other seat is still a target.
	if err := castAtPlayer(t, g, me, bolt, []string{"R"}, opp.ID); err != nil {
		t.Fatalf("a spell at the seat without shroud: %v", err)
	}
}

// TestShroudedPlayerIsOfferedToNobody: the enumerator the bot and the
// client's legal_targets read drops the seat for every chooser, its
// own included, so the refusal never comes as a surprise.
func TestShroudedPlayerIsOfferedToNobody(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	grantShroud(g, me)

	var mine, theirs LegalTargets
	g.ReadSnapshot(func() {
		mine = g.legalTargetsLocked(SourceChooser(me.ID), anyPlayerSpec())
		theirs = g.legalTargetsLocked(SourceChooser(opp.ID), anyPlayerSpec())
	})
	if containsPlayer(mine.Players, me.ID) {
		t.Error("a shrouded player is offered to their OWN picker")
	}
	if containsPlayer(theirs.Players, me.ID) {
		t.Error("a shrouded player is offered to an opponent's picker")
	}
	if !containsPlayer(theirs.Players, opp.ID) {
		t.Error("the enumeration lost the unshrouded seat")
	}
}

// TestHexproofWaiverDoesNotReachAPlayersShroud: Kaya's "as though they
// didn't have hexproof" waives hexproof and nothing else, as it does
// for a permanent's shroud.
func TestHexproofWaiverDoesNotReachAPlayersShroud(t *testing.T) {
	withHexproofBypasses(t, HexproofBypass{
		Player: func(_ *Game, target *Player, source *Card) bool { return target.ID != source.Controller },
	})
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushWaiver(g, opp)
	g.WithWriteLock(func() {
		g.GrantPlayerStaticForEffect(me.ID, KeywordHexproof, "Test — hexproof", uuid.Nil, Duration{})
	})

	var ok bool
	g.ReadSnapshot(func() { ok = g.CanPlayerBeTargetedByForEffect(me.ID, SourceChooser(opp.ID)) })
	if !ok {
		t.Fatal("control: the waiver should let its controller target a hexproof player")
	}

	grantShroud(g, me)
	g.ReadSnapshot(func() { ok = g.CanPlayerBeTargetedByForEffect(me.ID, SourceChooser(opp.ID)) })
	if ok {
		t.Error("a hexproof waiver let an opponent target a shrouded player")
	}
}

// TestPlayerShroudEndsAtCleanup: Gilded Light's duration is the zero
// Duration, until end of turn, so the next turn the player is a target
// again.
func TestPlayerShroudEndsAtCleanup(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	grantShroud(g, me)

	var before bool
	g.ReadSnapshot(func() { before = g.PlayerHasKeywordLocked(me, KeywordShroud) })
	if !before {
		t.Fatal("the grant did not land")
	}
	start := g.Turn.ActiveSeat
	for g.Turn.ActiveSeat == start {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	var after bool
	g.ReadSnapshot(func() { after = g.CanPlayerBeTargetedByForEffect(me.ID, SourceChooser(opp.ID)) })
	if !after {
		t.Error("shroud until end of turn outlived the turn")
	}
}
