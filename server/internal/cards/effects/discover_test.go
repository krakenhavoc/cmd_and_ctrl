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
		if isKeywordCastPrompt(c) && c.Chooser == chooser {
			answerKeywordCast(t, g, c, apply)
			return
		}
	}
	t.Fatalf("no may_cast prompt outstanding")
}

func discoverLibraryCard(owner uuid.UUID, name, typeLine, cost string) game.Card {
	return game.Card{InstanceID: uuid.New(), Name: name, TypeLine: typeLine, ManaCost: cost, Owner: owner, Controller: owner}
}

// Outside an open resolution there is no prompt to hold (the catalog's
// cards discover from inside one, and their tests answer the prompt): the
// primitive defaults to the controller, files what it found as a step
// record, and leaves the card in exile under the free-cast grant.
func TestDiscoverPrimitiveDefaultsToTheControllerAndFilesTheResult(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	me.Library.Cards = nil
	hit := discoverLibraryCard(me.ID, "Opt", "Instant", "{U}")
	pushLibraryCardForTest(me, hit)

	source := uuid.New()
	item := &game.StackItem{Controller: me.ID, SourceCardID: source}
	g.WithWriteLock(func() {
		if err := (Discover{N: 2}).Apply(NewContext(g, item)); err != nil {
			t.Fatalf("Discover: %v", err)
		}
	})
	rec, ok := g.StepRecordForEffect(item, DiscoverStepKey)
	if !ok || len(rec.IDs) != 1 || rec.IDs[0] != hit.InstanceID || rec.N() != 1 {
		t.Errorf("step record = %+v (filed %v), want the hit and its mana value 1", rec, ok)
	}
	perm := g.CastPermissionOnCardByIDForEffect(hit.InstanceID)
	if perm == nil || perm.Player != me.ID || perm.Cost != "{0}" || perm.Discover == nil || perm.Discover.N != 2 {
		t.Errorf("the controller's free-cast grant on the hit = %+v", perm)
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
	perm := g.CastPermissionOnCardByIDForEffect(hit.InstanceID)
	if perm == nil || perm.Player != opp.ID {
		t.Fatalf("the grant on the named player's card = %+v, want it theirs", perm)
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
