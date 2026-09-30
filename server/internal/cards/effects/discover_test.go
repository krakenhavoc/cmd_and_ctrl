package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// discover_test.go — the card-side words for CR 701.57 (ADR 0099). The
// engine rules are pinned in game/discover_test.go; these pin that the
// primitive discovers for the right player and hands its continuation
// a context, and that WheneverYouDiscover watches only its controller.

func answerMayCastPrompt(t *testing.T, g *game.Game, chooser uuid.UUID, apply bool) {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceMayCast {
			if err := g.ResolveMayCast(c.ID, chooser, apply); err != nil {
				t.Fatalf("ResolveMayCast: %v", err)
			}
			return
		}
	}
	t.Fatalf("no may_cast prompt outstanding")
}

func discoverLibraryCard(owner uuid.UUID, name, typeLine, cost string) game.Card {
	return game.Card{InstanceID: uuid.New(), Name: name, TypeLine: typeLine, ManaCost: cost, Owner: owner, Controller: owner}
}

func TestDiscoverPrimitiveDefaultsToTheControllerAndHandsThenAContext(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	me.Library.Cards = nil
	hit := discoverLibraryCard(me.ID, "Opt", "Instant", "{U}")
	pushLibraryCardForTest(me, hit)

	source := uuid.New()
	var got game.DiscoverResult
	var ctxController uuid.UUID
	g.WithWriteLock(func() {
		item := &game.StackItem{Controller: me.ID, SourceCardID: source}
		err := Discover{N: 2, Then: func(ctx *Context, r game.DiscoverResult) error {
			got = r
			ctxController = ctx.Controller()
			return nil
		}}.Apply(NewContext(g, item))
		if err != nil {
			t.Fatalf("Discover: %v", err)
		}
	})
	answerMayCastPrompt(t, g, me.ID, false)
	if !me.Hand.Contains(hit.InstanceID) {
		t.Fatalf("declined discover hit is not in the controller's hand")
	}
	if got.Discovered != hit.InstanceID || got.N != 2 || got.ManaValue != 1 || ctxController != me.ID {
		t.Errorf("Then saw %+v with controller %s", got, ctxController)
	}
}

func TestDiscoverPrimitiveNamesAnotherPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	opp.Library.Cards = nil
	hit := discoverLibraryCard(opp.ID, "Opt", "Instant", "{U}")
	pushLibraryCardForTest(opp, hit)
	g.WithWriteLock(func() {
		item := &game.StackItem{Controller: me.ID}
		_ = Discover{Player: opp.ID, N: 1}.Apply(NewContext(g, item))
	})
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceMayCast && c.Chooser != opp.ID {
			t.Fatalf("the discover prompt went to %s, not the named player", c.Chooser)
		}
	}
	answerMayCastPrompt(t, g, opp.ID, false)
	if !opp.Hand.Contains(hit.InstanceID) {
		t.Errorf("the named player's discovered card is not in their hand")
	}
}

func TestWheneverYouDiscoverWatchesOnlyItsController(t *testing.T) {
	ab := WheneverYouDiscover("test — whenever you discover", func(*game.Game, *game.StackItem) error { return nil })
	if len(ab.Watches) != 1 || ab.Watches[0] != game.EventDiscover {
		t.Fatalf("watches %v, want EventDiscover", ab.Watches)
	}
	me, other := uuid.New(), uuid.New()
	src := &game.Card{Controller: me}
	if !ab.AppliesTo(game.Event{Kind: game.EventDiscover, Actor: me}, src, game.Characteristic{}, nil) {
		t.Errorf("did not fire on its controller's discover")
	}
	if ab.AppliesTo(game.Event{Kind: game.EventDiscover, Actor: other}, src, game.Characteristic{}, nil) {
		t.Errorf("fired on another player's discover")
	}
}
