package testscr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestADayNightStaticFollowsTheDesignation(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := uuid.New()
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: id, Name: "Night Static Probe", TypeLine: "Creature — Bear",
		OracleID: effectstest.NightStaticProbeOracle, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	if got := effectstest.EffectivePower(t, g, id); got != 1 {
		t.Fatalf("neither day nor night: power %d, want 1", got)
	}
	g.WithWriteLock(func() { g.BecomeNightForEffect() })
	if got := effectstest.EffectivePower(t, g, id); got != 3 {
		t.Errorf("night: power %d, want 3 — the layer cache was not invalidated by the flip", got)
	}
	g.WithWriteLock(func() { g.BecomeDayForEffect() })
	if got := effectstest.EffectivePower(t, g, id); got != 1 {
		t.Errorf("day: power %d, want 1", got)
	}
}
