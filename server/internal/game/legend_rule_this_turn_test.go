package game

import (
	"testing"

	"github.com/google/uuid"
)

// legend_rule_this_turn_test.go — #3515: "the legend rule doesn't apply to
// permanents you control this turn" is a stored ScopedEffect record that
// legendRuleChoicesLocked reads, swept at cleanup, not a static.

func legendPair(g *Game, owner uuid.UUID, name string) (uuid.UUID, uuid.UUID) {
	return pushLegendForTest(g, owner, name, "Legendary Planeswalker — Test"),
		pushLegendForTest(g, owner, name, "Legendary Planeswalker — Test")
}

func TestLegendRuleThisTurnRecordSuspendsTheRuleForItsPlayerOnly(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	a, b := legendPair(g, me, "Mine")
	c, d := legendPair(g, opp, "Theirs")

	var wrote bool
	g.WithWriteLock(func() {
		wrote = g.LegendRuleDoesntApplyThisTurnForEffect(uuid.Nil, me, "test")
		g.runStateChecksLocked()
	})
	if !wrote {
		t.Fatal("the exemption was not registered")
	}
	if !onBattlefieldByID(g, a) || !onBattlefieldByID(g, b) || legendPromptFor(g, me) != nil {
		t.Error("the exempted player's duplicate legends were not left alone")
	}
	if legendPromptFor(g, opp) == nil || !onBattlefieldByID(g, c) || !onBattlefieldByID(g, d) {
		t.Error("another player's duplicate legends were exempted by a record naming someone else")
	}
}

func TestLegendRuleThisTurnRecordIsSweptAtCleanup(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0].ID
	a, b := legendPair(g, me, "Mine")

	g.WithWriteLock(func() {
		g.LegendRuleDoesntApplyThisTurnForEffect(uuid.Nil, me, "test")
		g.runStateChecksLocked()
	})
	if legendPromptFor(g, me) != nil {
		t.Fatal("the record did not suspend the rule while it stood")
	}
	g.WithWriteLock(func() {
		g.ClearEndOfTurnScopedStaticsLocked()
		g.runStateChecksLocked()
	})
	ch := legendPromptFor(g, me)
	if ch == nil {
		t.Fatal("the legend rule did not apply again after the cleanup sweep")
	}
	if err := g.ResolveLegendRule(ch.ID, me, b); err != nil {
		t.Fatalf("ResolveLegendRule: %v", err)
	}
	if onBattlefieldByID(g, a) || !onBattlefieldByID(g, b) {
		t.Error("the controller's choice was not honoured after the sweep")
	}
}

func TestLegendRuleThisTurnRecordNeedsASeatedPlayer(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	g.WithWriteLock(func() {
		if g.LegendRuleDoesntApplyThisTurnForEffect(uuid.Nil, uuid.Nil, "test") ||
			g.LegendRuleDoesntApplyThisTurnForEffect(uuid.Nil, uuid.New(), "test") {
			t.Error("a record was written for a player who is not seated")
		}
	})
}

// The record survives a restore point, so a table mid-turn keeps its
// exemption (the corpus board pins the file shape).
func TestLegendRuleThisTurnRecordSurvivesASnapshot(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0].ID
	a, b := legendPair(g, me, "Mine")
	g.WithWriteLock(func() { g.LegendRuleDoesntApplyThisTurnForEffect(uuid.Nil, me, "test") })

	restored, err := g.CaptureSnapshot().Restore()
	if err != nil {
		t.Fatalf("Restore: %v", err)
	}
	restored.WithWriteLock(func() { restored.runStateChecksLocked() })
	if legendPromptFor(restored, me) != nil || !onBattlefieldByID(restored, a) || !onBattlefieldByID(restored, b) {
		t.Error("the restored game lost the exemption")
	}
}
