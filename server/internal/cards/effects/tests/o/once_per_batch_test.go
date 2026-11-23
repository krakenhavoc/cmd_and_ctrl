package testso

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestOncePerBatchFiresOnceForOneBatchOfEvents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := effectstest.PushCatalogPermanent(g, me.ID, "Batch Probe", "Enchantment", effectstest.BatchProbeOracle, false)
	before := me.Life

	g.WithWriteLock(func() {
		for i := 0; i < 3; i++ {
			g.EmitEvent(game.Event{Kind: game.EventAttack, Actor: me.ID, CardID: uuid.New(), Target: opp.ID})
		}
	})
	if n := len(g.PendingTriggers); n != 1 {
		t.Fatalf("%d pending triggers after three attack events, want 1", n)
	}
	// A different ability of the same source is a different key, so
	// it is not suppressed while the attack trigger is pending.
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventDrawCard, Actor: me.ID, CardID: uuid.New()}) })
	if n := len(g.PendingTriggers); n != 2 {
		t.Fatalf("%d pending triggers, want 2 — the draw trigger must not be suppressed by the attack trigger", n)
	}
	// Two differing triggers from one controller stop on the order
	// prompt, which is the S19 rule and not this test's subject: take
	// the draw trigger back out and resolve the attack one alone.
	g.WithWriteLock(func() { g.PendingTriggers = g.PendingTriggers[:1] })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before+1 {
		t.Errorf("life %d, want %d (one attack trigger for three attack events)", me.Life, before+1)
	}
	// The batch is over once the item resolved: the next attack event fires again.
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventAttack, Actor: me.ID, CardID: uuid.New(), Target: opp.ID})
	})
	if effectstest.TriggerOnStack(g, id) == nil && len(g.PendingTriggers) != 1 {
		t.Error("a new batch after resolution did not fire")
	}
}

func TestOncePerBatchCoversThePromptWindow(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Batch May Probe", "Enchantment", effectstest.BatchMayProbeOracle, false)
	before := me.Life

	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventDrawCard, Actor: me.ID, CardID: uuid.New()})
		g.EmitEvent(game.Event{Kind: game.EventDrawCard, Actor: me.ID, CardID: uuid.New()})
	})
	if prompts := effectstest.OwedMays(g, me.ID); prompts != 1 {
		t.Fatalf("%d triggers for two draws of one batch, want 1", prompts)
	}
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before+1 {
		t.Errorf("life %d, want %d", me.Life, before+1)
	}
}

func TestAdelineStillMakesOneBatchOfHumansPerAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Adeline, Resplendent Cathar", "Legendary Creature — Human Knight",
		"38515f89-348b-4cf3-b7bd-1f6fe4ce2fba", false)
	a := effectstest.PushCatalogPermanent(g, me.ID, "Bear A", "Creature — Bear", "", false)
	b := effectstest.PushCatalogPermanent(g, me.ID, "Bear B", "Creature — Bear", "", false)
	before := effectstest.CountBattlefieldNamed(g, me.ID, "Human")

	effectstest.DeclareAttack(t, g, opp.ID, a, b)
	effectstest.PassPriorityAroundTable(t, g)

	// One trigger for the whole declaration: one Human per opponent, once.
	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Human") - before; got != len(g.Seats)-1 {
		t.Errorf("Adeline made %d Humans for a two-creature attack, want %d (one per opponent, one batch)", got, len(g.Seats)-1)
	}
}

func TestDourPortMageDrawsAgainForASecondBatchWhileTheFirstDrawIsOnTheStack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mage := effectstest.B12Push(g, me.ID, "Dour Port-Mage", "Creature — Frog Wizard", effectstest.B829PortMageOracle, 1, 3)
	first := effectstest.B12Creature(g, me.ID, "First", "Creature — Bear", 2, 2)
	second := effectstest.B12Creature(g, me.ID, "Second", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	hand := me.Hand.Size()

	effectstest.CastCatalogSpell(t, g, "Unsummon", "Instant", effectstest.B829UnsummonOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: first}})
	effectstest.PassUntilSpellResolves(t, g)
	if effectstest.TriggerOnStack(g, mage) == nil {
		t.Fatal("the first bounce's draw trigger should be waiting on the stack")
	}

	// A second bounce, in response, while that draw is still on the
	// stack. Its own resolution is a new batch, so the Port-Mage
	// triggers again.
	effectstest.CastCatalogSpell(t, g, "Unsummon", "Instant", effectstest.B829UnsummonOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: second}})
	effectstest.PassPriorityAroundTable(t, g)

	if !me.Hand.Contains(first) || !me.Hand.Contains(second) {
		t.Fatal("both creatures should have been bounced to hand")
	}
	// Two bounced creatures and TWO cards drawn.
	if got := me.Hand.Size() - hand; got != 4 {
		t.Errorf("hand +%d, want +4 (two bounced creatures and two draws — one per batch)", got)
	}
}

func TestDourPortMageDrawsOnceForOneMassBounce(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Dour Port-Mage", "Creature — Frog Wizard", effectstest.B829PortMageOracle, 1, 3)
	for _, name := range []string{"First", "Second", "Third"} {
		effectstest.B12Creature(g, me.ID, name, "Creature — Bear", 2, 2)
	}
	effectstest.AdvanceToMain(t, g)
	hand := me.Hand.Size()

	effectstest.CastCatalogSpell(t, g, "Evacuation", "Instant", effectstest.B829EvacuationOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	// Three Bears and the Port-Mage itself return, and the three
	// OTHER creatures leaving are one batch: ONE card drawn.
	if got := me.Hand.Size() - hand; got != 5 {
		t.Errorf("hand +%d, want +5 (four bounced permanents and ONE draw for the batch)", got)
	}
}
