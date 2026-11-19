package game

import (
	"testing"

	"github.com/google/uuid"
)

// player_shroud_test.go — #5842, CR 702.18a on a PLAYER: "You can't be
// the target of spells or abilities." Gilded Light grants it for a turn.
// The card half is in cards/effects/gilded_light_test.go.
//
// Written against hexproof's asymmetry, because that is the difference
// between the two: hexproof (CR 702.11c) asks who is casting, shroud
// asks nothing.

// grantPlayerKeyword is Gilded Light's clause (or Blossoming Calm's,
// with "hexproof"), applied directly, until end of turn.
func grantPlayerKeyword(g *Game, p *Player, keyword string) {
	g.WithWriteLock(func() {
		g.GrantPlayerStaticForEffect(p.ID, keyword, "Test — you gain "+keyword, uuid.Nil, Duration{})
	})
}

// TestShroudPlayerRefusesEveryoneThemselvesIncluded is the rule: an
// opponent's spell is refused, and so is the player's own.
func TestShroudPlayerRefusesEveryoneThemselvesIncluded(t *testing.T) {
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
	grantPlayerKeyword(g, me, KeywordShroud)

	if err := castAtPlayer(t, g, me, bolt, []string{"R"}, me.ID); err != ErrIllegalTarget {
		t.Fatalf("your own spell at yourself under shroud: got %v, want ErrIllegalTarget", err)
	}
	if err := castAtPlayer(t, g, opp, bolt, []string{"R"}, me.ID); err != ErrIllegalTarget {
		t.Fatalf("an opponent's spell at a shrouded player: got %v, want ErrIllegalTarget", err)
	}
	// Shroud covers only the player who has it.
	if err := castAtPlayer(t, g, me, bolt, []string{"R"}, opp.ID); err != nil {
		t.Fatalf("a spell at the unshrouded seat: %v", err)
	}
}

// TestShroudPlayerIsOfferedToNobody: the bot's move list and the
// client's legal_targets read the same enumeration, so a shrouded seat
// is missing from every picker, its own included. A hexproof seat, for
// contrast, is still in its own.
func TestShroudPlayerIsOfferedToNobody(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	grantPlayerKeyword(g, me, KeywordShroud)
	grantPlayerKeyword(g, opp, KeywordHexproof)

	var mine, theirs LegalTargets
	g.ReadSnapshot(func() {
		mine = g.legalTargetsLocked(SourceChooser(me.ID), anyPlayerSpec())
		theirs = g.legalTargetsLocked(SourceChooser(opp.ID), anyPlayerSpec())
	})
	if containsPlayer(mine.Players, me.ID) {
		t.Error("a shrouded player is offered to their own picker")
	}
	if containsPlayer(theirs.Players, me.ID) {
		t.Error("a shrouded player is offered to an opponent's picker")
	}
	if !containsPlayer(theirs.Players, opp.ID) {
		t.Error("a hexproof player is missing from their own picker")
	}
	if containsPlayer(mine.Players, opp.ID) {
		t.Error("a hexproof player is offered to an opponent's picker")
	}
}

// TestShroudPlayerIsNotWaivedByAHexproofBypass: "as though they didn't
// have hexproof" (Kaya, Bane of the Dead) names hexproof, so it waives a
// hexproof player and leaves a shrouded one alone.
func TestShroudPlayerIsNotWaivedByAHexproofBypass(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	// Unrestricted (not YoursOnly), so it waives for either seat's
	// spells: the two players below are each other's opponent.
	withHexproofBypasses(t, HexproofBypass{
		Player: func(*Game, *Player, *Card) bool { return true },
	})
	pushWaiver(g, opp)
	grantPlayerKeyword(g, me, KeywordShroud)
	grantPlayerKeyword(g, opp, KeywordHexproof)

	var shrouded, hexproof bool
	g.ReadSnapshot(func() {
		shrouded = g.CanPlayerBeTargetedByForEffect(me.ID, SourceChooser(opp.ID))
		hexproof = g.CanPlayerBeTargetedByForEffect(opp.ID, SourceChooser(me.ID))
	})
	if !hexproof {
		t.Fatal("the waiver did not reach a hexproof player, so this test proves nothing")
	}
	if shrouded {
		t.Error("a hexproof waiver let an opponent target a shrouded player")
	}
}

// TestShroudPlayerEndsAtEndOfTurn: the grant's zero Duration is "until
// end of turn", and the cleanup sweep takes it away.
func TestShroudPlayerEndsAtEndOfTurn(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	grantPlayerKeyword(g, me, KeywordShroud)

	var before bool
	g.ReadSnapshot(func() { before = g.PlayerHasKeywordLocked(me, KeywordShroud) })
	if !before {
		t.Fatal("the grant gave no shroud")
	}
	g.WithWriteLock(func() { g.sweepPlayerStaticsLocked(true) })

	var after, targetable bool
	g.ReadSnapshot(func() {
		after = g.PlayerHasKeywordLocked(me, KeywordShroud)
		targetable = g.CanPlayerBeTargetedByForEffect(me.ID, SourceChooser(opp.ID))
	})
	if after || !targetable {
		t.Errorf("after cleanup: shroud %v, targetable %v; want the shroud gone", after, targetable)
	}
}
