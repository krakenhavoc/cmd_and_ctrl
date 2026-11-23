package testsu

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestZackFairMovesTheEquipmentThatWasAttachedToHim(t *testing.T) {
	g, zack, bear, me := effectstest.ZackFairBoard(t)
	sword := effectstest.SeedEquipment(g, me, "Test Blade", "")
	if err := effectstest.CloudDeckAttach(g, sword, game.TargetRef{Kind: game.TargetCard, ID: zack}); err != nil {
		t.Fatalf("attach: %v", err)
	}

	effectstest.ZackFairActivate(t, g, zack, bear)
	effectstest.PassPriorityAroundTable(t, g)

	if effects.OnBattlefield(g, zack) {
		t.Fatal("Zack was not sacrificed")
	}
	if host := effectstest.AttachmentHostOf(t, g, sword); host.Kind != game.TargetCard || host.ID != bear {
		t.Errorf("Equipment host = %+v, want the Bear %s", host, bear)
	}
}

// Only an Equipment is moved: the Aura on Zack goes to the graveyard
// before the ability resolves (CR 704.5m) and is not "an Equipment".
func TestZackFairLeavesAnAuraAlone(t *testing.T) {
	g, zack, bear, me := effectstest.ZackFairBoard(t)
	aura := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Test Aura", TypeLine: "Enchantment — Aura",
		Owner: me, Controller: me,
	})
	if err := effectstest.CloudDeckAttach(g, aura, game.TargetRef{Kind: game.TargetCard, ID: zack}); err != nil {
		t.Fatalf("attach: %v", err)
	}

	effectstest.ZackFairActivate(t, g, zack, bear)
	effectstest.PassPriorityAroundTable(t, g)

	if effects.OnBattlefield(g, aura) {
		if host := effectstest.AttachmentHostOf(t, g, aura); host.ID == bear {
			t.Error("the Aura was moved to the Bear; only an Equipment moves")
		}
	}
	if got := effectstest.CloudDeckCounters(t, g, bear); got != 1 {
		t.Errorf("Bear counters = %d, want 1 (the counter still moves)", got)
	}
}

// "An Equipment": with two attached, the controller picks the one that
// moves, and the other stays unattached.
func TestZackFairWithTwoEquipmentAsksWhichMoves(t *testing.T) {
	g, zack, bear, me := effectstest.ZackFairBoard(t)
	first := effectstest.SeedEquipment(g, me, "First Blade", "")
	second := effectstest.SeedEquipment(g, me, "Second Blade", "")
	for _, eq := range []uuid.UUID{first, second} {
		if err := effectstest.CloudDeckAttach(g, eq, game.TargetRef{Kind: game.TargetCard, ID: zack}); err != nil {
			t.Fatalf("attach: %v", err)
		}
	}

	effectstest.ZackFairActivate(t, g, zack, bear)
	effectstest.PassPriorityAroundTable(t, g)

	if pick := effectstest.LatestChooseCardsFor(g, me); pick == nil {
		t.Fatal("no prompt asked which Equipment moves")
	} else if len(pick.ChooseCards) != 2 {
		t.Errorf("prompt offered %d Equipment, want 2", len(pick.ChooseCards))
	}
	effectstest.AnswerChooseCards(t, g, me, second)

	if host := effectstest.AttachmentHostOf(t, g, second); host.ID != bear {
		t.Errorf("chosen Equipment host = %+v, want the Bear", host)
	}
	if host := effectstest.AttachmentHostOf(t, g, first); host.Kind != "" {
		t.Errorf("the other Equipment is attached to %+v, want unattached", host)
	}
}

// An Equipment that left the battlefield before the ability resolved is
// not the object that was attached; nothing is moved and nothing fails.
func TestZackFairSkipsAnEquipmentThatHasGone(t *testing.T) {
	g, zack, bear, me := effectstest.ZackFairBoard(t)
	sword := effectstest.SeedEquipment(g, me, "Test Blade", "")
	if err := effectstest.CloudDeckAttach(g, sword, game.TargetRef{Kind: game.TargetCard, ID: zack}); err != nil {
		t.Fatalf("attach: %v", err)
	}

	effectstest.ZackFairActivate(t, g, zack, bear)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(sword); err != nil {
			t.Fatalf("destroy Equipment: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)

	if effects.OnBattlefield(g, sword) {
		t.Fatal("setup: the Equipment should have been destroyed")
	}
	if got := effectstest.CloudDeckCounters(t, g, bear); got != 1 {
		t.Errorf("Bear counters = %d, want 1", got)
	}
}
