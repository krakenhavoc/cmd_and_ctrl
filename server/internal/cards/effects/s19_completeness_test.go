package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s19_completeness_test.go — S19 sub-PR 8: the cross-cutting cases
// the per-card files don't cover. Same-controller ordering via a
// real catalog wrath, CR 603.10 last-known information under a
// layer effect, and the manual-announce fallback for non-catalog
// cards.

// TestSameControllerDiesTriggersPromptForOrder: Damnation kills a
// Doomed Traveler and a Wurmcoil Engine under one controller. Two
// differing triggers → a trigger_order prompt; the chosen order is
// the resolution order; nothing reaches the stack until answered.
func TestSameControllerDiesTriggersPromptForOrder(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	travelerID := pushDiesCreatureForTest(g, me.ID, "Doomed Traveler",
		"a30907c0-fbde-4fd3-a8c7-f304305fcea7", "Creature — Human Soldier", 1, 1)
	wurmID := pushDiesCreatureForTest(g, me.ID, "Wurmcoil Engine",
		"d1a60f44-7696-49ee-91fb-cab5b3102962", "Artifact Creature — Phyrexian Wurm", 6, 6)

	castCatalogSpell(t, g, "Damnation", "Sorcery",
		"d57a8f0b-7989-4db5-8756-6f2690097252", nil)
	for i := 0; i < 8 && g.Stack.Size() > 0; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}

	var prompt *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceTriggerOrder && c.Chooser == me.ID {
			prompt = c
		}
	}
	if prompt == nil {
		t.Fatalf("no trigger_order prompt for the controller of two differing dies-triggers")
	}
	if len(g.StackMeta) != 0 {
		t.Fatalf("triggers reached the stack before the order was chosen")
	}
	var travelerTrig, wurmTrig uuid.UUID
	for _, tr := range g.PendingTriggers {
		switch tr.SourceCardID {
		case travelerID:
			travelerTrig = tr.ID
		case wurmID:
			wurmTrig = tr.ID
		}
	}
	if travelerTrig == uuid.Nil || wurmTrig == uuid.Nil {
		t.Fatalf("pending triggers missing: traveler=%v wurm=%v", travelerTrig, wurmTrig)
	}

	// Wurms first, then the Spirit.
	if err := g.ResolveTriggerOrder(prompt.ID, me.ID, []uuid.UUID{wurmTrig, travelerTrig}); err != nil {
		t.Fatalf("ResolveTriggerOrder: %v", err)
	}
	if len(g.StackMeta) != 2 {
		t.Fatalf("stack items after ordering = %d, want 2", len(g.StackMeta))
	}
	wurmItem, travItem := g.StackMeta[wurmTrig], g.StackMeta[travelerTrig]
	if wurmItem == nil || travItem == nil || !(wurmItem.Seq > travItem.Seq) {
		t.Errorf("Wurmcoil trigger must sit above Doomed Traveler's (resolves first)")
	}
	// One pass around resolves the top item only.
	for i := 0; i < 8 && len(g.StackMeta) == 2; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if n := countOnBattlefieldByName(g, "Phyrexian Wurm", me.ID); n != 2 {
		t.Errorf("after the first resolution: %d Wurms, want 2 (Wurmcoil chosen to resolve first)", n)
	}
	if n := countOnBattlefieldByName(g, "Spirit", me.ID); n != 0 {
		t.Errorf("Spirit appeared before its trigger resolved")
	}
	passPriorityAroundTable(t, g)
	if n := countOnBattlefieldByName(g, "Spirit", me.ID); n != 1 {
		t.Errorf("Spirit tokens: %d, want 1", n)
	}
}

func countOnBattlefieldByName(g *game.Game, name string, controller uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == name && c.Controller == controller {
			n++
		}
	}
	return n
}

// TestDiesTriggerLKISeesAnthemPower pins CR 603.10 through the real
// layer engine: a creature dying under Glorious Anthem hands its
// dies-trigger the pumped power, not the printed one. Uses a stub
// TriggeredAbility registered against a throwaway oracle ID so the
// assertion is on the harvester's LKI argument itself.
func TestDiesTriggerLKISeesAnthemPower(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Glorious Anthem", TypeLine: "Enchantment",
		OracleID: "e3886fe8-9b76-4613-8891-4ec74657c087", Owner: me.ID, Controller: me.ID,
	})
	const oracle = "test-lki-anthem-oracle"
	var seenPower int
	Register(Spec{
		OracleID: oracle,
		Name:     "LKI Probe",
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return cardDied(ev, source)
			},
			Build: func(_ game.Event, source *game.Card, lki game.Characteristic, _ *game.Game) *game.StackItem {
				seenPower = lki.Power
				return nil
			},
		}},
	})
	t.Cleanup(func() { delete(registry, oracle) })

	probeID := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "LKI Probe", TypeLine: "Creature — Test",
		OracleID: oracle, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	if got := effectivePower(t, g, probeID); got != 3 {
		t.Fatalf("anthem not applied before death: effective power %d, want 3", got)
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(probeID); err != nil {
			t.Fatalf("DestroyPermanentForEffect: %v", err)
		}
	})
	if seenPower != 3 {
		t.Errorf("dies-trigger LKI power = %d, want 3 (printed 2 + anthem)", seenPower)
	}
}

// TestNonCatalogPermanentFallsBackToManualTrigger is the S19 opt-in
// canary: a permanent with no catalog entry never auto-fires, and
// the S13.1 announce_trigger path still puts a (effect-less) item on
// the stack for the players to resolve by hand.
func TestNonCatalogPermanentFallsBackToManualTrigger(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	homebrewID := pushDiesCreatureForTest(g, me.ID, "Homebrew Creature",
		"not-in-catalog-oracle", "Creature — Test", 2, 2)
	handBefore := me.Hand.Size()

	// Its "ETB" and "dies" produce no auto-trigger.
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventETB, CardID: homebrewID, Actor: me.ID})
	})
	if len(g.PendingTriggers) != 0 || len(g.StackMeta) != 0 {
		t.Fatalf("non-catalog card auto-fired a trigger")
	}

	// Manual announce still works and lands on the stack — since #974
	// as the announce returns, rather than at whatever boundary came
	// next.
	if err := g.AnnounceTrigger(me.ID, homebrewID, game.AbilityParams{Label: "Homebrew ETB — draw (manual)"}); err != nil {
		t.Fatalf("AnnounceTrigger: %v", err)
	}
	item := triggerOnStack(g, homebrewID)
	if item == nil {
		t.Fatalf("manual trigger not on the stack")
	}
	if item.Effect != nil {
		t.Errorf("manual trigger must have no engine-side effect")
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != handBefore {
		t.Errorf("manual trigger resolution changed the hand — it should be effect-less")
	}
}
