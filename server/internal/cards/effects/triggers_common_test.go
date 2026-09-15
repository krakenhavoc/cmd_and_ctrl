package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// triggers_common_test.go — the constructor vocabulary (#579). The 400+
// migrated card files are the broad guard; these pin the contract the
// constructors promise on their own: the item goes on the stack, Do
// defaults the controller, Optional asks first, and the predicates
// read the event the way their names say.

const triggerProbeOracle = "test-triggers-common-probe"

func init() {
	Register(Spec{
		OracleID: triggerProbeOracle,
		Name:     "Constructor Probe",
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Constructor Probe — you gain 2 life", Do(GainLife{Amount: 2})),
			Optional(WhenThisDies("Constructor Probe — draw a card", Do(DrawCards{N: 1})),
				"Constructor Probe — draw a card?"),
		},
	})
}

func TestWhenThisEntersUsesTheStackAndDoDefaultsTheController(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	before := me.Life

	id := enterFromHand(t, g, me.ID, "Constructor Probe", "Creature — Test", triggerProbeOracle)

	item := triggerOnStack(g, id)
	if item == nil {
		t.Fatal("WhenThisEnters did not put an item on the stack")
	}
	if item.Label != "Constructor Probe — you gain 2 life" {
		t.Errorf("label %q: the constructor must carry the label verbatim", item.Label)
	}
	if me.Life != before {
		t.Fatal("the effect ran before the trigger resolved")
	}
	passPriorityAroundTable(t, g)
	if me.Life != before+2 {
		t.Errorf("life %d, want %d — Do(GainLife{Amount: 2}) should default to the item's controller", me.Life, before+2)
	}
}

func TestOptionalWhenThisDiesAsksBeforeBuilding(t *testing.T) {
	for _, answer := range []bool{true, false} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		id := pushCatalogPermanent(g, me.ID, "Constructor Probe", "Creature — Test", triggerProbeOracle, false)
		before := me.Hand.Size()

		g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(id) })
		if triggerOnStack(g, id) != nil {
			t.Fatal("an Optional trigger reached the stack before its prompt was answered")
		}
		answerLatestTriggerPrompt(t, g, me.ID, answer)
		passPriorityAroundTable(t, g)

		want := before
		if answer {
			want++
		}
		if got := me.Hand.Size(); got != want {
			t.Errorf("answer=%v: hand %d, want %d", answer, got, want)
		}
	}
}

func TestCastPredicatesReadTheSpellAndTheCaster(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	creature := game.NewCard("Bear", me.ID)
	creature.TypeLine = "Creature — Bear"
	instant := game.NewCard("Shock", me.ID)
	instant.TypeLine = "Instant"
	me.Hand.PushTop(creature)
	me.Hand.PushTop(instant)
	src := &game.Card{InstanceID: uuid.New(), Controller: me.ID, Owner: me.ID}
	lki := game.Characteristic{}
	cast := func(actor, card uuid.UUID) game.Event {
		return game.Event{Kind: game.EventCast, Actor: actor, CardID: card}
	}

	if !YouCast(Noncreature())(cast(me.ID, instant.InstanceID), src, lki, g) {
		t.Error("YouCast(Noncreature()) missed my instant")
	}
	if YouCast(Noncreature())(cast(me.ID, creature.InstanceID), src, lki, g) {
		t.Error("YouCast(Noncreature()) matched my creature")
	}
	if YouCast(nil)(cast(opp.ID, instant.InstanceID), src, lki, g) {
		t.Error("YouCast matched an opponent's spell")
	}
	if !YouCast(nil)(cast(me.ID, creature.InstanceID), src, lki, g) {
		t.Error("YouCast(nil) should match any spell I cast")
	}
	if !AnOpponentCast(nil)(cast(opp.ID, instant.InstanceID), src, lki, g) {
		t.Error("AnOpponentCast missed an opponent's spell")
	}
	if AnOpponentCast(nil)(cast(me.ID, instant.InstanceID), src, lki, g) {
		t.Error("AnOpponentCast matched my own spell")
	}
	if YouCast(nil)(game.Event{Kind: game.EventDrawCard, Actor: me.ID, CardID: instant.InstanceID}, src, lki, g) {
		t.Error("YouCast matched a non-cast event")
	}
}

type failingStep struct{}

func (failingStep) Apply(*Context) error { return errors.New("boom") }

func TestDoStopsAtTheFirstError(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	before := me.Life
	item := &game.StackItem{Controller: me.ID}

	var err error
	g.WithWriteLock(func() { err = Do(failingStep{}, GainLife{Amount: 2})(g, item) })
	if err == nil || err.Error() != "boom" {
		t.Fatalf("err = %v, want boom", err)
	}
	if me.Life != before {
		t.Error("a step after the failing one still ran")
	}
	g.WithWriteLock(func() { err = Do(GainLife{Amount: 2}, failingStep{})(g, item) })
	if err == nil {
		t.Fatal("the error from the second step was swallowed")
	}
	if me.Life != before+2 {
		t.Error("the step before the failing one did not run")
	}
}

func TestConstructorsBuildOrdinaryAbilities(t *testing.T) {
	noop := func(*game.Game, *game.StackItem) error { return nil }
	t1 := WhenThisEnters("x", noop)
	if len(t1.Watches) != 1 || t1.Watches[0] != game.EventETB || t1.AppliesTo == nil || t1.Build == nil {
		t.Errorf("WhenThisEnters: %+v", t1)
	}
	if t1.OptionalPrompt != nil || t1.Targets != nil {
		t.Error("WhenThisEnters set a prompt or a target clause on its own")
	}
	t2 := Optional(t1, "really?")
	if t2.OptionalPrompt == nil || t2.OptionalPrompt.Question != "really?" || t1.OptionalPrompt != nil {
		t.Error("Optional must set the prompt on a copy and leave the original alone")
	}
	spec := TargetCreature("target creature")
	t3 := Targeting(t1, spec)
	if t3.Targets != spec || t1.Targets != nil {
		t.Error("Targeting must set the clause on a copy and leave the original alone")
	}
	t4 := WhenThisEntersOrAttacks("x", noop)
	if len(t4.Watches) != 2 || t4.Watches[0] != game.EventETB || t4.Watches[1] != game.EventAttack {
		t.Errorf("WhenThisEntersOrAttacks watches %v", t4.Watches)
	}
	src := &game.Card{InstanceID: uuid.New(), Controller: uuid.New()}
	if !Self(game.Event{CardID: src.InstanceID}, src, game.Characteristic{}, nil) || Self(game.Event{CardID: uuid.New()}, src, game.Characteristic{}, nil) {
		t.Error("Self reads ev.CardID against the source")
	}
	if !ByYou(game.Event{Actor: src.Controller}, src, game.Characteristic{}, nil) || ByAnOpponent(game.Event{Actor: src.Controller}, src, game.Characteristic{}, nil) {
		t.Error("ByYou / ByAnOpponent read ev.Actor against the controller")
	}
	if ByAnOpponent(game.Event{}, src, game.Characteristic{}, nil) {
		t.Error("ByAnOpponent must not match an actor-less event")
	}
	all := AllOf(Self, ByYou)
	if !all(game.Event{CardID: src.InstanceID, Actor: src.Controller}, src, game.Characteristic{}, nil) || all(game.Event{CardID: src.InstanceID}, src, game.Characteristic{}, nil) {
		t.Error("AllOf must require every condition")
	}
}
