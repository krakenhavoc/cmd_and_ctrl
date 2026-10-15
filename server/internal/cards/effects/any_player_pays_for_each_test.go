package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func eachOpenLeg(g *game.Game) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceChooseCards && c.AnyPlayerOffer != nil {
			return c
		}
	}
	return nil
}

// any_player_pays_for_each_test.go — ADR 0187 §2 (#3572): the card side
// of the per-object any-player offer. No card uses it yet (Cleansing is
// slice 6), so the primitive is applied straight from the effect.
func TestAnyPlayerPaysForEachBindsObjectsAndOnUnsavedToTheItem(t *testing.T) {
	g := newCatalogGame(t)
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
	runs := 0
	var unsaved []uuid.UUID
	var sawController uuid.UUID
	g.WithWriteLock(func() {
		err := AnyPlayerPaysForEach{
			Objects:  func(*Context) []uuid.UUID { return lands },
			EachLife: 1,
			Noun:     "land",
			OnUnsaved: func(ctx *Context, ids []uuid.UUID) error {
				runs++
				unsaved = ids
				sawController = ctx.Controller()
				return nil
			},
		}.Apply(NewContext(g, item))
		if err != nil {
			t.Fatalf("Apply: %v", err)
		}
	})
	link := eachOpenLeg(g)
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
	for runs == 0 {
		next := eachOpenLeg(g)
		if next == nil {
			break
		}
		if err := g.ResolveChooseCards(next.ID, next.Chooser, nil); err != nil {
			t.Fatalf("answer: %v", err)
		}
	}
	if runs != 1 || len(unsaved) != 1 || unsaved[0] != lands[1] {
		t.Fatalf("OnUnsaved ran %d times on %v, want once on seat 1's land", runs, unsaved)
	}
	if sawController != controller.ID {
		t.Errorf("OnUnsaved's context controller = %v, want the item's controller %v", sawController, controller.ID)
	}
	if g.Seats[0].Life != game.StartingLife-1 {
		t.Errorf("seat 0 life = %d, want one life paid", g.Seats[0].Life)
	}
}
