package testsa

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// any_player_pays_for_each_test.go — ADR 0187 §2 (#3572): the card side
// of the per-object any-player offer. No card uses it yet (Cleansing is
// slice 6), so the primitive is applied straight from the effect.
// testEachUnsaved is the per-object offer's consequence as a registered
// key (#4253 retired the OnUnsaved closure): it records what it was told
// in the test's package-level probe.
var (
	testEachUnsavedRuns       int
	testEachUnsavedIDs        []uuid.UUID
	testEachUnsavedController uuid.UUID
	testEachUnsaved           = effects.AnyPlayerEachConsequence("test/4253-any-player-each-unsaved", func(ctx *effects.Context, r game.AnyPlayerEachAnswer) error {
		testEachUnsavedRuns++
		testEachUnsavedIDs = r.Unsaved
		testEachUnsavedController = ctx.Controller()
		return nil
	})
)

func TestAnyPlayerPaysForEachBindsObjectsAndOnUnsavedToTheItem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	controller := g.Seats[2]
	source := uuid.New()
	g.Battlefield.PushTop(game.Card{InstanceID: source, Name: "Test Cleansing", TypeLine: "Sorcery", Owner: controller.ID, Controller: controller.ID})
	var lands []uuid.UUID
	for i := 0; i < 2; i++ {
		id := uuid.New()
		g.Battlefield.PushTop(game.Card{InstanceID: id, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: g.Seats[i].ID, Controller: g.Seats[i].ID})
		lands = append(lands, id)
	}
	item := &game.StackItem{Controller: controller.ID, SourceCardID: source}
	testEachUnsavedRuns, testEachUnsavedIDs, testEachUnsavedController = 0, nil, uuid.Nil
	g.WithWriteLock(func() {
		err := effects.AnyPlayerPaysForEach{
			Objects:  func(*effects.Context) []uuid.UUID { return lands },
			EachLife: 1,
			Noun:     "land",
			Then:     testEachUnsaved,
		}.Apply(effects.NewContext(g, item))
		if err != nil {
			t.Fatalf("Apply: %v", err)
		}
	})
	link := effectstest.EachOpenLeg(g)
	if link == nil || link.Kind != game.PendingChoiceChooseCards || link.Chooser != g.Seats[0].ID {
		t.Fatalf("open prompt %+v, want the active player's choose_cards leg", link)
	}
	if o := link.AnyPlayerOffer; o == nil || o.EachLife != 1 || o.Controller != controller.ID {
		t.Fatalf("offer %+v, want one life each, the item's controller", o)
	}
	// Seat 0 saves its own land; seat 1 and the rest pay for nothing.
	if err := g.ResolveChooseCards(link.ID, link.Chooser, []uuid.UUID{lands[0]}); err != nil {
		t.Fatalf("seat 0 answer: %v", err)
	}
	for testEachUnsavedRuns == 0 {
		next := effectstest.EachOpenLeg(g)
		if next == nil {
			break
		}
		if err := g.ResolveChooseCards(next.ID, next.Chooser, nil); err != nil {
			t.Fatalf("answer: %v", err)
		}
	}
	if runs, unsaved := testEachUnsavedRuns, testEachUnsavedIDs; runs != 1 || len(unsaved) != 1 || unsaved[0] != lands[1] {
		t.Fatalf("the unsaved consequence ran %d times on %v, want once on seat 1's land", runs, unsaved)
	}
	if testEachUnsavedController != controller.ID {
		t.Errorf("the consequence's context controller = %v, want the item's controller %v", testEachUnsavedController, controller.ID)
	}
	if g.Seats[0].Life != game.StartingLife-1 {
		t.Errorf("seat 0 life = %d, want one life paid", g.Seats[0].Life)
	}
}
