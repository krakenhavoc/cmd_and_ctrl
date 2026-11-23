package teststo

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestWhenThisEntersUsesTheStackAndDoDefaultsTheController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	before := me.Life

	id := effectstest.EnterFromHand(t, g, me.ID, "Constructor Probe", "Creature — Test", effectstest.TriggerProbeOracle)

	item := effectstest.TriggerOnStack(g, id)
	if item == nil {
		t.Fatal("WhenThisEnters did not put an item on the stack")
	}
	if item.Label != "Constructor Probe — you gain 2 life" {
		t.Errorf("label %q: the constructor must carry the label verbatim", item.Label)
	}
	if me.Life != before {
		t.Fatal("the effect ran before the trigger resolved")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before+2 {
		t.Errorf("life %d, want %d — Do(GainLife{Amount: 2}) should default to the item's controller", me.Life, before+2)
	}
}

func TestOptionalWhenThisDiesAsksBeforeBuilding(t *testing.T) {
	for _, answer := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		id := effectstest.PushCatalogPermanent(g, me.ID, "Constructor Probe", "Creature — Test", effectstest.TriggerProbeOracle, false)
		before := me.Hand.Size()

		g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(id) })
		if effectstest.TriggerOnStack(g, id) != nil {
			t.Fatal("an Optional trigger reached the stack before its prompt was answered")
		}
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, answer)
		effectstest.PassPriorityAroundTable(t, g)

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
	g := effectstest.NewCatalogGame(t)
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

	if !effects.YouCast(effects.Noncreature())(cast(me.ID, instant.InstanceID), src, lki, g) {
		t.Error("YouCast(Noncreature()) missed my instant")
	}
	if effects.YouCast(effects.Noncreature())(cast(me.ID, creature.InstanceID), src, lki, g) {
		t.Error("YouCast(Noncreature()) matched my creature")
	}
	if effects.YouCast(nil)(cast(opp.ID, instant.InstanceID), src, lki, g) {
		t.Error("YouCast matched an opponent's spell")
	}
	if !effects.YouCast(nil)(cast(me.ID, creature.InstanceID), src, lki, g) {
		t.Error("YouCast(nil) should match any spell I cast")
	}
	if !effects.AnOpponentCast(nil)(cast(opp.ID, instant.InstanceID), src, lki, g) {
		t.Error("AnOpponentCast missed an opponent's spell")
	}
	if effects.AnOpponentCast(nil)(cast(me.ID, instant.InstanceID), src, lki, g) {
		t.Error("AnOpponentCast matched my own spell")
	}
	if effects.YouCast(nil)(game.Event{Kind: game.EventDrawCard, Actor: me.ID, CardID: instant.InstanceID}, src, lki, g) {
		t.Error("YouCast matched a non-cast event")
	}
}

type failingStep struct{}

func (failingStep) Apply(*effects.Context) error { return errors.New("boom") }

func TestDoStopsAtTheFirstError(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	before := me.Life
	item := &game.StackItem{Controller: me.ID}

	var err error
	g.WithWriteLock(func() { err = effects.Do(failingStep{}, effects.GainLife{Amount: 2})(g, item) })
	if err == nil || err.Error() != "boom" {
		t.Fatalf("err = %v, want boom", err)
	}
	if me.Life != before {
		t.Error("a step after the failing one still ran")
	}
	g.WithWriteLock(func() { err = effects.Do(effects.GainLife{Amount: 2}, failingStep{})(g, item) })
	if err == nil {
		t.Fatal("the error from the second step was swallowed")
	}
	if me.Life != before+2 {
		t.Error("the step before the failing one did not run")
	}
}

func TestConstructorsBuildOrdinaryAbilities(t *testing.T) {
	noop := func(*game.Game, *game.StackItem) error { return nil }
	t1 := effects.WhenThisEnters("x", noop)
	// ADR 0041 P9 (tier 4-2): the effect is declared on the row and
	// the engine builds the item, so a constructor carries no Build.
	if len(t1.Watches) != 1 || t1.Watches[0] != game.EventETB || t1.AppliesTo == nil ||
		t1.Effect == nil || t1.Build != nil || t1.Key != "x" {
		t.Errorf("WhenThisEnters: %+v", t1)
	}
	if t1.OptionalPrompt != nil || t1.Targets != nil {
		t.Error("WhenThisEnters set a prompt or a target clause on its own")
	}
	t2 := effects.Optional(t1, "really?")
	if t2.OptionalPrompt == nil || t2.OptionalPrompt.Question != "really?" || t1.OptionalPrompt != nil {
		t.Error("Optional must set the prompt on a copy and leave the original alone")
	}
	spec := effects.TargetCreature("target creature")
	t3 := effects.Targeting(t1, spec)
	if t3.Targets != spec || t1.Targets != nil {
		t.Error("Targeting must set the clause on a copy and leave the original alone")
	}
	t4 := effects.WhenThisEntersOrAttacks("x", noop)
	if len(t4.Watches) != 2 || t4.Watches[0] != game.EventETB || t4.Watches[1] != game.EventAttack {
		t.Errorf("WhenThisEntersOrAttacks watches %v", t4.Watches)
	}
	src := &game.Card{InstanceID: uuid.New(), Controller: uuid.New()}
	if !effects.Self(game.Event{CardID: src.InstanceID}, src, game.Characteristic{}, nil) || effects.Self(game.Event{CardID: uuid.New()}, src, game.Characteristic{}, nil) {
		t.Error("Self reads ev.CardID against the source")
	}
	if !effects.ByYou(game.Event{Actor: src.Controller}, src, game.Characteristic{}, nil) || effects.ByAnOpponent(game.Event{Actor: src.Controller}, src, game.Characteristic{}, nil) {
		t.Error("ByYou / ByAnOpponent read ev.Actor against the controller")
	}
	if effects.ByAnOpponent(game.Event{}, src, game.Characteristic{}, nil) {
		t.Error("ByAnOpponent must not match an actor-less event")
	}
	all := effects.AllOf(effects.Self, effects.ByYou)
	if !all(game.Event{CardID: src.InstanceID, Actor: src.Controller}, src, game.Characteristic{}, nil) || all(game.Event{CardID: src.InstanceID}, src, game.Characteristic{}, nil) {
		t.Error("AllOf must require every condition")
	}
}
