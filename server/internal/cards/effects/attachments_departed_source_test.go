package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// attachments_departed_source_test.go — #812's card-level proof.
//
// The soak found it on Loxodon Warhammer: its equip was activated and
// the Warhammer was sacrificed to Krark-Clan Ironworks in response.
// CR 608.2 resolves the ability anyway — its target is still a
// creature its controller controls — and CR 301.5c means the attach
// does nothing, because only a permanent on the battlefield can be
// attached. The engine logged an EventEffectError instead, which the
// catalog soak treats as a bug and fails the nightly run on.
//
// The rule lives at one choke point (game.AttachSourceForEffect); no
// card file checks anything. These two tests are the printed card
// going through it.

// countCatalogEvents is the event-log read these tests share.
func countCatalogEvents(g *game.Game, kind game.EventKind, from int) int {
	n := 0
	g.ReadSnapshot(func() {
		for _, ev := range g.Events[from:] {
			if ev.Kind == kind {
				n++
			}
		}
	})
	return n
}

// activateEquipWithoutResolving activates the equipment's equip
// ability (always its last declared ability) and leaves it on the
// stack, so the test can act in the response window equipTo closes.
func activateEquipWithoutResolving(t *testing.T, g *game.Game, controller, equipment, creature uuid.UUID) {
	t.Helper()
	abilities := game.ActivatedAbilitiesForCard(game.Card{OracleID: oracleOfBattlefieldCard(t, g, equipment)})
	idx := len(abilities) - 1
	if idx < 0 {
		t.Fatal("equipment has no activated ability")
	}
	if err := g.ActivateCatalogAbility(controller, equipment, idx, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: creature}},
	}); err != nil {
		t.Fatalf("equip: %v", err)
	}
}

// The reported bug: equip on the stack, Equipment sacrificed in
// response. The ability resolves, nothing attaches, the creature is
// untouched, and no effect error is logged.
func TestEquipResolvingAfterTheEquipmentWasSacrificedDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: testCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	hammer := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Loxodon Warhammer", TypeLine: equipTypeLine,
		OracleID: warhammerOracle, Owner: me.ID, Controller: me.ID,
	})

	activateEquipWithoutResolving(t, g, me.ID, hammer, bear)
	before := len(g.Events)
	g.WithWriteLock(func() {
		if err := g.SacrificePermanentForEffect(hammer); err != nil {
			t.Fatalf("sacrifice in response: %v", err)
		}
	})
	passPriorityAroundTable(t, g)

	if n := countCatalogEvents(g, game.EventEffectError, before); n != 0 {
		t.Errorf("EventEffectError count = %d, want 0 — the equip resolved and did nothing", n)
	}
	if n := countCatalogEvents(g, game.EventAttachSkipped, before); n != 1 {
		t.Errorf("EventAttachSkipped count = %d, want 1", n)
	}
	if n := countCatalogEvents(g, game.EventAttach, before); n != 0 {
		t.Errorf("EventAttach count = %d, want 0", n)
	}
	var attachments []uuid.UUID
	g.ReadSnapshot(func() { attachments = g.AttachmentsOf(bear) })
	if len(attachments) != 0 {
		t.Errorf("the Bear picked up %v", attachments)
	}
	if got := effectivePower(t, g, bear); got != 2 {
		t.Errorf("Bear power %d, want 2 — the Warhammer's +3/+0 must not apply", got)
	}
}

// CR 400.7: bounced and replayed before equip resolves. A card with
// the Warhammer's instance ID is back on the battlefield, and it is a
// different object, so the old ability attaches nothing.
func TestEquipResolvingAfterTheEquipmentWasBouncedAndReplayedDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: testCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	hammer := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Loxodon Warhammer", TypeLine: equipTypeLine,
		OracleID: warhammerOracle, Owner: me.ID, Controller: me.ID,
	})

	activateEquipWithoutResolving(t, g, me.ID, hammer, bear)
	before := len(g.Events)
	g.WithWriteLock(func() {
		if err := g.BounceToHandForEffect(hammer); err != nil {
			t.Fatalf("bounce in response: %v", err)
		}
		if _, err := g.PutFromHandOntoBattlefieldForEffect(hammer, game.HandEntryOptions{Controller: me.ID}); err != nil {
			t.Fatalf("replay: %v", err)
		}
	})
	passPriorityAroundTable(t, g)

	if n := countCatalogEvents(g, game.EventEffectError, before); n != 0 {
		t.Errorf("EventEffectError count = %d, want 0", n)
	}
	if host := attachmentHostOf(t, g, hammer); host.Kind != "" {
		t.Errorf("the NEW Warhammer attached to %+v — CR 400.7 says it is a new object", host)
	}
	if got := effectivePower(t, g, bear); got != 2 {
		t.Errorf("Bear power %d, want 2", got)
	}
}

// The control, on the same card and the same harness: nothing happens
// in the response window, the equip resolves, the Warhammer attaches
// and both statics reach the host.
func TestEquipResolvingWithItsEquipmentStillThereAttaches(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: testCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	hammer := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Loxodon Warhammer", TypeLine: equipTypeLine,
		OracleID: warhammerOracle, Owner: me.ID, Controller: me.ID,
	})
	before := len(g.Events)

	equipTo(t, g, me.ID, hammer, bear)

	if host := attachmentHostOf(t, g, hammer); host.Kind != game.TargetCard || host.ID != bear {
		t.Fatalf("AttachedTo = %+v, want card %s", host, bear)
	}
	if got := effectivePower(t, g, bear); got != 5 {
		t.Errorf("Bear power %d, want 5", got)
	}
	if n := countCatalogEvents(g, game.EventAttachSkipped, before); n != 0 {
		t.Errorf("EventAttachSkipped count = %d, want 0", n)
	}
}
