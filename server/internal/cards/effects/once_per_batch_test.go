package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// once_per_batch_test.go — #587: "whenever one or more …" is an engine
// flag. The engine emits one event per creature; an OncePerBatch
// ability fires for the first event of an event batch and declines
// every later event of that same batch.
//
// #829 replaced the "while its item is pending, on the stack, or
// waiting on its prompt" reading of "the batch" with the real batch
// identity on Event.Batch — see the tests at the bottom of this file
// and server/internal/game/event_batch.go.

const (
	batchProbeOracle      = "test-once-per-batch-probe"
	batchMayProbeOracle   = "test-once-per-batch-optional-probe"
	batchProbeAttackLabel = "Batch Probe — you gain 1 life"
)

func init() {
	Register(Spec{
		OracleID: batchProbeOracle,
		Name:     "Batch Probe",
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attackDeclaredByYou(ev, source.Controller)
			}, batchProbeAttackLabel, Do(GainLife{Amount: 1}))),
			// A second, unrelated ability on the same source: its own
			// key, so the first one's item must not suppress it.
			WheneverYouDraw("Batch Probe — you gain 5 life", Do(GainLife{Amount: 5})),
		},
	})
	Register(Spec{
		OracleID: batchMayProbeOracle,
		Name:     "Batch May Probe",
		Triggered: []game.TriggeredAbility{
			OncePerBatch(Optional(WheneverYouDraw("Batch May Probe — you may gain 1 life", Do(GainLife{Amount: 1})),
				"Batch May Probe — gain 1 life?")),
		},
	})
}

func TestOncePerBatchFiresOnceForOneBatchOfEvents(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := pushCatalogPermanent(g, me.ID, "Batch Probe", "Enchantment", batchProbeOracle, false)
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
	passPriorityAroundTable(t, g)
	if me.Life != before+1 {
		t.Errorf("life %d, want %d (one attack trigger for three attack events)", me.Life, before+1)
	}
	// The batch is over once the item resolved: the next attack event fires again.
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventAttack, Actor: me.ID, CardID: uuid.New(), Target: opp.ID})
	})
	if triggerOnStack(g, id) == nil && len(g.PendingTriggers) != 1 {
		t.Error("a new batch after resolution did not fire")
	}
}

func TestOncePerBatchCoversThePromptWindow(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Batch May Probe", "Enchantment", batchMayProbeOracle, false)
	before := me.Life

	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventDrawCard, Actor: me.ID, CardID: uuid.New()})
		g.EmitEvent(game.Event{Kind: game.EventDrawCard, Actor: me.ID, CardID: uuid.New()})
	})
	prompts := 0
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceTriggerPrompt && c.Chooser == me.ID {
			prompts++
		}
	}
	if prompts != 1 {
		t.Fatalf("%d trigger prompts for two draws, want 1 — the second event arrived while the first was waiting on its prompt", prompts)
	}
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if me.Life != before+1 {
		t.Errorf("life %d, want %d", me.Life, before+1)
	}
}

func TestAdelineStillMakesOneBatchOfHumansPerAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Adeline, Resplendent Cathar", "Legendary Creature — Human Knight",
		"38515f89-348b-4cf3-b7bd-1f6fe4ce2fba", false)
	a := pushCatalogPermanent(g, me.ID, "Bear A", "Creature — Bear", "", false)
	b := pushCatalogPermanent(g, me.ID, "Bear B", "Creature — Bear", "", false)
	before := countBattlefieldNamed(g, me.ID, "Human")

	declareAttack(t, g, opp.ID, a, b)
	passPriorityAroundTable(t, g)

	// One trigger for the whole declaration: one Human per opponent, once.
	if got := countBattlefieldNamed(g, me.ID, "Human") - before; got != len(g.Seats)-1 {
		t.Errorf("Adeline made %d Humans for a two-creature attack, want %d (one per opponent, one batch)", got, len(g.Seats)-1)
	}
}

// --- #829: a LATER batch is a second trigger -----------------------
//
// "One or more" collapses one batch, not everything the ability has
// in flight. Dour Port-Mage is the reported card: two bounces at two
// different times are two occurrences (CR 603.2c), even when the
// first one's draw is still sitting on the stack.

const (
	b829UnsummonOracle   = "837182db-1bf3-4a2c-bd01-1af9d9873561"
	b829EvacuationOracle = "fdd94383-b573-439a-8e1c-925af887c5a6"
	b829PortMageOracle   = "cf58e309-00e8-438e-813e-2e1c1002db23"
)

func TestDourPortMageDrawsAgainForASecondBatchWhileTheFirstDrawIsOnTheStack(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mage := b12Push(g, me.ID, "Dour Port-Mage", "Creature — Frog Wizard", b829PortMageOracle, 1, 3)
	first := b12Creature(g, me.ID, "First", "Creature — Bear", 2, 2)
	second := b12Creature(g, me.ID, "Second", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	hand := me.Hand.Size()

	castCatalogSpell(t, g, "Unsummon", "Instant", b829UnsummonOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: first}})
	passUntilSpellResolves(t, g)
	if triggerOnStack(g, mage) == nil {
		t.Fatal("the first bounce's draw trigger should be waiting on the stack")
	}

	// A second bounce, in response, while that draw is still on the
	// stack. Its own resolution is a new batch, so the Port-Mage
	// triggers again.
	castCatalogSpell(t, g, "Unsummon", "Instant", b829UnsummonOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: second}})
	passPriorityAroundTable(t, g)

	if !me.Hand.Contains(first) || !me.Hand.Contains(second) {
		t.Fatal("both creatures should have been bounced to hand")
	}
	// Two bounced creatures and TWO cards drawn.
	if got := me.Hand.Size() - hand; got != 4 {
		t.Errorf("hand +%d, want +4 (two bounced creatures and two draws — one per batch)", got)
	}
}

func TestDourPortMageDrawsOnceForOneMassBounce(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Dour Port-Mage", "Creature — Frog Wizard", b829PortMageOracle, 1, 3)
	for _, name := range []string{"First", "Second", "Third"} {
		b12Creature(g, me.ID, name, "Creature — Bear", 2, 2)
	}
	advanceToMain(t, g)
	hand := me.Hand.Size()

	castCatalogSpell(t, g, "Evacuation", "Instant", b829EvacuationOracle, nil)
	passPriorityAroundTable(t, g)

	// Three Bears and the Port-Mage itself return, and the three
	// OTHER creatures leaving are one batch: ONE card drawn.
	if got := me.Hand.Size() - hand; got != 5 {
		t.Errorf("hand +%d, want +5 (four bounced permanents and ONE draw for the batch)", got)
	}
}
