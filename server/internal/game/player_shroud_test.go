package game

import (
	"testing"

	"github.com/google/uuid"
)

// player_shroud_test.go — #5842, CR 702.18a on a PLAYER: "Shroud" means
// "This permanent or player can't be the target of spells or
// abilities." Hexproof (CR 702.11c) asks who controls the source;
// shroud asks nothing, so the player's own spell is refused as well.
// Each test is written against its hexproof twin in
// player_protection_test.go, and the difference is the assertion.

// grantShroudUntilEOT is Gilded Light's clause, applied directly.
func grantShroudUntilEOT(g *Game, p *Player) {
	g.WithWriteLock(func() {
		g.GrantPlayerStaticForEffect(p.ID, KeywordShroud, "Test — shroud",
			uuid.Nil, g.UntilEndOfTurnDuration())
	})
}

// TestShroudedPlayerRefusesEveryoneIncludingThemselves is the headline,
// and the asymmetry with TestHexproofPlayerRefusesAnOpponentAndNotThemselves
// is the whole point: the player's OWN spell at themselves is refused.
func TestShroudedPlayerRefusesEveryoneIncludingThemselves(t *testing.T) {
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

	grantShroudUntilEOT(g, me)
	if !playerHasAbility(g, me, KeywordShroud) {
		t.Fatal("the grant did not take at all")
	}

	if err := castAtPlayer(t, g, me, bolt, []string{"R"}, me.ID); err != ErrIllegalTarget {
		t.Fatalf("your own spell at yourself under shroud: got %v, want ErrIllegalTarget", err)
	}
	if err := castAtPlayer(t, g, opp, bolt, []string{"R"}, me.ID); err != ErrIllegalTarget {
		t.Fatalf("an opponent's spell at a shrouded player: got %v, want ErrIllegalTarget", err)
	}
	// The shroud is the granted player's alone.
	if err := castAtPlayer(t, g, me, bolt, []string{"R"}, opp.ID); err != nil {
		t.Fatalf("a spell at the UNshrouded seat: %v", err)
	}
}

// TestShroudedPlayerIsNotOfferedByTheEnumerator: the bot's move list and
// the client's legal_targets read legalTargetsLocked, so the shrouded
// seat is missing from EVERY picker, its own included.
func TestShroudedPlayerIsNotOfferedByTheEnumerator(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	grantShroudUntilEOT(g, me)

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
	if !containsPlayer(mine.Players, opp.ID) || !containsPlayer(theirs.Players, opp.ID) {
		t.Error("the enumeration lost the unshrouded seat")
	}
}

// TestPlayerShroudEndsAtEndOfTurn: "until end of turn" (CR 514.2), so
// the next turn the player is an ordinary target again, and the sweep
// has dropped the entry.
func TestPlayerShroudEndsAtEndOfTurn(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	grantShroudUntilEOT(g, me)
	if !playerHasAbility(g, me, KeywordShroud) {
		t.Fatal("the grant did not take at all")
	}
	advanceOneTurn(t, g)
	if playerHasAbility(g, me, KeywordShroud) {
		t.Error("the shroud survived the turn it was granted in")
	}
	if n := len(me.Statics); n != 0 {
		t.Errorf("Player.Statics still holds %d expired entries; the sweep did not run", n)
	}
	var targets LegalTargets
	g.ReadSnapshot(func() {
		targets = g.legalTargetsLocked(SourceChooser(me.ID), anyPlayerSpec())
	})
	if !containsPlayer(targets.Players, me.ID) {
		t.Error("the player is still not targetable after the shroud ended")
	}
}
