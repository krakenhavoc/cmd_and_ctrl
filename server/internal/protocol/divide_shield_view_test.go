package protocol

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// divide_shield_view_test.go — ADR 0108 §7 (#1904): the divide_shield
// prompt reaches the wire with its charge and one entry per damage
// event, so the client can answer it, and the live source shields reach
// the banner.
func TestDivideShieldIsOnTheWire(t *testing.T) {
	g := newTwoSeatGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := game.NewCard("Hill Giant", opp.ID)
	src.TypeLine = "Creature — Giant"
	src.Power, src.Toughness = 3, 3
	g.Battlefield.PushTop(src)
	bear := game.NewCard("Bear", me.ID)
	bear.TypeLine = "Creature — Bear"
	bear.Power, bear.Toughness = 2, 2
	g.Battlefield.PushTop(bear)
	g.WithWriteLock(func() {
		ref, zone, _ := g.DamageSourceRefLocked(src.InstanceID)
		g.PreventDamageFromSourceThisTurnForEffect(game.DamageShield{
			Controller: me.ID, Source: ref, SourceZone: zone, ProtectPlayer: me.ID,
			ProtectTypes: []string{"creature"}, Amount: 2, Label: "Refraction Trap",
		})
		_ = g.DamageInstanceForEffect(func() error {
			_ = g.DealDamageToCreatureForEffect(src.InstanceID, bear.InstanceID, 2)
			return g.DealDamageToPlayerForEffect(src.InstanceID, me.ID, 3)
		})
	})
	v := ViewOfGame(g)
	if len(v.DamageShields) != 1 || v.DamageShields[0] != "Refraction Trap (Hill Giant) — 2 left" {
		t.Errorf("damage_shields = %v", v.DamageShields)
	}
	var ds *DivideShieldView
	for _, c := range v.PendingChoices {
		if c.Kind == string(game.PendingChoiceDivideShield) {
			ds = c.DivideShield
		}
	}
	if ds == nil {
		t.Fatal("no divide_shield on the wire")
	}
	if ds.Charge != 2 || len(ds.Entries) != 2 {
		t.Fatalf("divide_shield = %+v, want 2 to divide between two events", ds)
	}
	if e := ds.Entries[0]; e.SourceName != "Hill Giant" || e.TargetName != "Bear" || e.Amount != 2 || e.TargetIsPlayer {
		t.Errorf("first entry = %+v", e)
	}
	if e := ds.Entries[1]; e.TargetID != me.ID.String() || !e.TargetIsPlayer || e.Amount != 3 {
		t.Errorf("second entry = %+v", e)
	}
}
