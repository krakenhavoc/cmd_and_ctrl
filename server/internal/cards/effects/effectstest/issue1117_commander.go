package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// issue1117_commander_test.go — the #1117 Betor life-swap deck's
// commander and its three lifegain payoffs.

const (
	E2BetorOracle            = "990b5e12-6e04-4832-9d64-87278f12cbda"
	E2RodolfOracle           = "26a716b2-a2e2-405b-accd-df572416d8bf"
	E2EnduringTenacityOracle = "98e698ae-1a69-469c-9cfb-0e3fedeb71d4"
	E2KrrikOracle            = "cbe3a4e7-5dbe-4f58-8ee6-a1762b65acfd"
)

// e2ChangeLife is the drain / gain the tally counts.
func E2ChangeLife(g *game.Game, player uuid.UUID, delta int) {
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, player, delta) })
}

// e2CastSpellOfColor casts a non-catalog spell whose only interesting
// property is its colour — K'rrik's trigger reads the spell's COLOUR,
// not a pip in its cost, so the colour is set directly and the cost
// left empty.
func E2CastSpellOfColor(t *testing.T, g *game.Game, name, typeLine string, colors ...string) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, Colors: colors,
		Owner: active.ID, Controller: active.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}
