package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// zack_fair_test.go — Zack Fair's "attach an Equipment that was attached
// to Zack Fair to that creature" (#3780): the sacrificed permanent's
// last-known attachments, read after the state-based action has already
// unattached them (CR 704.5n).

func zackFairActivate(t *testing.T, g *game.Game, zack, target uuid.UUID) {
	t.Helper()
	me := g.Seats[0]
	me.ManaPool.AddMana(game.ManaToken{Color: "W"})
	if err := g.ActivateCatalogAbility(me.ID, zack, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
}

func zackFairBoard(t *testing.T) (g *game.Game, zack, bear, me uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	seat := g.Seats[0]
	advanceToMain(t, g)
	zack = castAndResolveCreature(t, g, "Zack Fair", "Legendary Creature — Human Soldier", zackFairOracle)
	passPriorityAroundTable(t, g)
	return g, zack, seedBear(g, seat.ID), seat.ID
}

func TestZackFairMovesTheEquipmentThatWasAttachedToHim(t *testing.T) {
	g, zack, bear, me := zackFairBoard(t)
	sword := seedEquipment(g, me, "Test Blade", "")
	if err := cloudDeckAttach(g, sword, game.TargetRef{Kind: game.TargetCard, ID: zack}); err != nil {
		t.Fatalf("attach: %v", err)
	}

	zackFairActivate(t, g, zack, bear)
	passPriorityAroundTable(t, g)

	if onBattlefield(g, zack) {
		t.Fatal("Zack was not sacrificed")
	}
	if host := attachmentHostOf(t, g, sword); host.Kind != game.TargetCard || host.ID != bear {
		t.Errorf("Equipment host = %+v, want the Bear %s", host, bear)
	}
}

// Only an Equipment is moved: the Aura on Zack goes to the graveyard
// before the ability resolves (CR 704.5m) and is not "an Equipment".
func TestZackFairLeavesAnAuraAlone(t *testing.T) {
	g, zack, bear, me := zackFairBoard(t)
	aura := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Test Aura", TypeLine: "Enchantment — Aura",
		Owner: me, Controller: me,
	})
	if err := cloudDeckAttach(g, aura, game.TargetRef{Kind: game.TargetCard, ID: zack}); err != nil {
		t.Fatalf("attach: %v", err)
	}

	zackFairActivate(t, g, zack, bear)
	passPriorityAroundTable(t, g)

	if onBattlefield(g, aura) {
		if host := attachmentHostOf(t, g, aura); host.ID == bear {
			t.Error("the Aura was moved to the Bear; only an Equipment moves")
		}
	}
	if got := cloudDeckCounters(t, g, bear); got != 1 {
		t.Errorf("Bear counters = %d, want 1 (the counter still moves)", got)
	}
}

// "An Equipment": with two attached, the controller picks the one that
// moves, and the other stays unattached.
func TestZackFairWithTwoEquipmentAsksWhichMoves(t *testing.T) {
	g, zack, bear, me := zackFairBoard(t)
	first := seedEquipment(g, me, "First Blade", "")
	second := seedEquipment(g, me, "Second Blade", "")
	for _, eq := range []uuid.UUID{first, second} {
		if err := cloudDeckAttach(g, eq, game.TargetRef{Kind: game.TargetCard, ID: zack}); err != nil {
			t.Fatalf("attach: %v", err)
		}
	}

	zackFairActivate(t, g, zack, bear)
	passPriorityAroundTable(t, g)

	if pick := latestChooseCardsFor(g, me); pick == nil {
		t.Fatal("no prompt asked which Equipment moves")
	} else if len(pick.ChooseCards) != 2 {
		t.Errorf("prompt offered %d Equipment, want 2", len(pick.ChooseCards))
	}
	answerChooseCards(t, g, me, second)

	if host := attachmentHostOf(t, g, second); host.ID != bear {
		t.Errorf("chosen Equipment host = %+v, want the Bear", host)
	}
	if host := attachmentHostOf(t, g, first); host.Kind != "" {
		t.Errorf("the other Equipment is attached to %+v, want unattached", host)
	}
}

// An Equipment that left the battlefield before the ability resolved is
// not the object that was attached; nothing is moved and nothing fails.
func TestZackFairSkipsAnEquipmentThatHasGone(t *testing.T) {
	g, zack, bear, me := zackFairBoard(t)
	sword := seedEquipment(g, me, "Test Blade", "")
	if err := cloudDeckAttach(g, sword, game.TargetRef{Kind: game.TargetCard, ID: zack}); err != nil {
		t.Fatalf("attach: %v", err)
	}

	zackFairActivate(t, g, zack, bear)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(sword); err != nil {
			t.Fatalf("destroy Equipment: %v", err)
		}
	})
	passPriorityAroundTable(t, g)

	if onBattlefield(g, sword) {
		t.Fatal("setup: the Equipment should have been destroyed")
	}
	if got := cloudDeckCounters(t, g, bear); got != 1 {
		t.Errorf("Bear counters = %d, want 1", got)
	}
}
