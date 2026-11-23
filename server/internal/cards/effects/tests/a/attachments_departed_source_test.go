package testsa

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The reported bug: equip on the stack, Equipment sacrificed in
// response. The ability resolves, nothing attaches, the creature is
// untouched, and no effect error is logged.
func TestEquipResolvingAfterTheEquipmentWasSacrificedDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: effectstest.TestCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	hammer := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Loxodon Warhammer", TypeLine: effectstest.EquipTypeLine,
		OracleID: effectstest.WarhammerOracle, Owner: me.ID, Controller: me.ID,
	})

	effectstest.ActivateEquipWithoutResolving(t, g, me.ID, hammer, bear)
	before := len(g.Events)
	g.WithWriteLock(func() {
		if err := g.SacrificePermanentForEffect(hammer); err != nil {
			t.Fatalf("sacrifice in response: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)

	if n := effectstest.CountCatalogEvents(g, game.EventEffectError, before); n != 0 {
		t.Errorf("EventEffectError count = %d, want 0 — the equip resolved and did nothing", n)
	}
	if n := effectstest.CountCatalogEvents(g, game.EventAttachSkipped, before); n != 1 {
		t.Errorf("EventAttachSkipped count = %d, want 1", n)
	}
	if n := effectstest.CountCatalogEvents(g, game.EventAttach, before); n != 0 {
		t.Errorf("EventAttach count = %d, want 0", n)
	}
	var attachments []uuid.UUID
	g.ReadSnapshot(func() { attachments = g.AttachmentsOf(bear) })
	if len(attachments) != 0 {
		t.Errorf("the Bear picked up %v", attachments)
	}
	if got := effectstest.EffectivePower(t, g, bear); got != 2 {
		t.Errorf("Bear power %d, want 2 — the Warhammer's +3/+0 must not apply", got)
	}
}

// CR 400.7: bounced and replayed before equip resolves. A card with
// the Warhammer's instance ID is back on the battlefield, and it is a
// different object, so the old ability attaches nothing.
func TestEquipResolvingAfterTheEquipmentWasBouncedAndReplayedDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: effectstest.TestCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	hammer := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Loxodon Warhammer", TypeLine: effectstest.EquipTypeLine,
		OracleID: effectstest.WarhammerOracle, Owner: me.ID, Controller: me.ID,
	})

	effectstest.ActivateEquipWithoutResolving(t, g, me.ID, hammer, bear)
	before := len(g.Events)
	g.WithWriteLock(func() {
		if err := g.BounceToHandForEffect(hammer); err != nil {
			t.Fatalf("bounce in response: %v", err)
		}
		if _, err := g.PutFromHandOntoBattlefieldForEffect(hammer, game.HandEntryOptions{Controller: me.ID}); err != nil {
			t.Fatalf("replay: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)

	if n := effectstest.CountCatalogEvents(g, game.EventEffectError, before); n != 0 {
		t.Errorf("EventEffectError count = %d, want 0", n)
	}
	if host := effectstest.AttachmentHostOf(t, g, hammer); host.Kind != "" {
		t.Errorf("the NEW Warhammer attached to %+v — CR 400.7 says it is a new object", host)
	}
	if got := effectstest.EffectivePower(t, g, bear); got != 2 {
		t.Errorf("Bear power %d, want 2", got)
	}
}

// The control, on the same card and the same harness: nothing happens
// in the response window, the equip resolves, the Warhammer attaches
// and both statics reach the host.
func TestEquipResolvingWithItsEquipmentStillThereAttaches(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: effectstest.TestCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	hammer := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Loxodon Warhammer", TypeLine: effectstest.EquipTypeLine,
		OracleID: effectstest.WarhammerOracle, Owner: me.ID, Controller: me.ID,
	})
	before := len(g.Events)

	effectstest.EquipTo(t, g, me.ID, hammer, bear)

	if host := effectstest.AttachmentHostOf(t, g, hammer); host.Kind != game.TargetCard || host.ID != bear {
		t.Fatalf("AttachedTo = %+v, want card %s", host, bear)
	}
	if got := effectstest.EffectivePower(t, g, bear); got != 5 {
		t.Errorf("Bear power %d, want 5", got)
	}
	if n := effectstest.CountCatalogEvents(g, game.EventAttachSkipped, before); n != 0 {
		t.Errorf("EventAttachSkipped count = %d, want 0", n)
	}
}
