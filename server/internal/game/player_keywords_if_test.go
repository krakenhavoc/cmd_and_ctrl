package game

import "testing"

// player_keywords_if_test.go — #2013, CR 702.11c: a permanent's
// printed "During your turn, you … have hexproof". The condition is
// asked by the same walk every player-ability consumer reads, on every
// query, so the answer follows the turn with nothing stored.

// withCatalogPlayerKeywordsIf stubs the conditional half of the reader.
func withCatalogPlayerKeywordsIf(t *testing.T, fn func(oracleID string) []PlayerKeywordIf) {
	t.Helper()
	prev := CatalogPlayerKeywordsIf
	CatalogPlayerKeywordsIf = fn
	t.Cleanup(func() { CatalogPlayerKeywordsIf = prev })
}

// TestConditionalPlayerHexproofFollowsTheCondition: the opponent's
// spell is refused while the condition holds and accepted the moment it
// stops, through the one targeting predicate the announce gate, the
// resolution re-check and the enumerator share.
func TestConditionalPlayerHexproofFollowsTheCondition(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	const breaker = "test-player-hexproof-your-turn"
	withCatalogPlayerKeywordsIf(t, func(id string) []PlayerKeywordIf {
		if id != breaker {
			return nil
		}
		return []PlayerKeywordIf{{
			Keyword: KeywordHexproof,
			Label:   "During your turn, you have hexproof.",
			While: func(g *Game, source Card) bool {
				return g.Seats[g.Turn.ActiveSeat].ID == source.Controller
			},
		}}
	})
	pushLeyline(t, g, me, breaker)

	var onMyTurn, mineOnMyTurn bool
	var abilities []string
	g.ReadSnapshot(func() {
		onMyTurn = g.CanPlayerBeTargetedByForEffect(me.ID, SourceChooser(opp.ID))
		mineOnMyTurn = g.CanPlayerBeTargetedByForEffect(me.ID, SourceChooser(me.ID))
		abilities = g.PlayerAbilitiesForEffect(me)
	})
	if g.Turn.ActiveSeat != 0 {
		t.Fatalf("expected seat 0's turn, got %d", g.Turn.ActiveSeat)
	}
	if onMyTurn {
		t.Error("an opponent may target the player while the condition holds")
	}
	if !mineOnMyTurn {
		t.Error("the player can't target themselves under their own hexproof")
	}
	if len(abilities) != 1 || abilities[0] != KeywordHexproof {
		t.Errorf("the player's abilities are %v, want [hexproof]", abilities)
	}

	g.WithWriteLock(func() { g.Turn.ActiveSeat = 1 })
	var onTheirTurn bool
	g.ReadSnapshot(func() {
		onTheirTurn = g.CanPlayerBeTargetedByForEffect(me.ID, SourceChooser(opp.ID))
		abilities = g.PlayerAbilitiesForEffect(me)
	})
	if !onTheirTurn {
		t.Error("the hexproof outlived its condition")
	}
	if len(abilities) != 0 {
		t.Errorf("with the condition false the player has %v", abilities)
	}
}

// TestConditionalPlayerKeywordWithNoConditionGrantsNothing: a nil While
// is the weaker direction, not "always".
func TestConditionalPlayerKeywordWithNoConditionGrantsNothing(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	const oracle = "test-player-hexproof-no-condition"
	withCatalogPlayerKeywordsIf(t, func(id string) []PlayerKeywordIf {
		if id == oracle {
			return []PlayerKeywordIf{{Keyword: KeywordHexproof, Label: "x"}}
		}
		return nil
	})
	pushLeyline(t, g, me, oracle)
	var has bool
	g.ReadSnapshot(func() { has = g.PlayerHasKeywordLocked(me, KeywordHexproof) })
	if has {
		t.Error("an entry with no condition granted hexproof")
	}
}
