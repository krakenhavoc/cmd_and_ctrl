package game

import (
	"testing"

	"github.com/google/uuid"
)

// exile_until_cast_test.go — #3213. The caller's hit test decides where
// the walk stops, the grant has no mana value cap, and the chooser is
// whoever the caller names. Since #4250 the cast is made inside the
// resolution that walks (CR 608.2g), through a resolution_cast prompt.

func nonlandNotNamed(name string) func(Card) bool {
	return func(c Card) bool { return !c.IsLand() && c.Name != name }
}

// exileUntilResolving resolves an ability whose effect is the walk for
// `player`, so the cast is offered inside an open resolution. keep picks
// the keeping form (#4205).
func exileUntilResolving(t *testing.T, g *Game, player, source uuid.UUID, hit func(Card) bool, keep bool) {
	t.Helper()
	pushAbilityItem(g, StackItemActivated, g.Seats[0].ID, uuid.Nil, "Fixture exile-until", func(g *Game, _ *StackItem) error {
		if keep {
			return g.ExileUntilCastFreeKeepingForEffect(player, source, hit)
		}
		return g.ExileUntilCastFreeForEffect(player, source, hit)
	})
	resolveTop(t, g)
}

// The walk skips lands and a card with the excluded name, stops on the
// first card the test accepts, and leaves the cards below it alone. The
// offer is the walked player's, inside the resolution.
func TestExileUntilCastStopsOnTheCallersHit(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[1]
	hit := libCard(me.ID, "Divination", "Sorcery", "{2}{U}")
	same := libCard(me.ID, "Same Name", "Instant", "{U}")
	land := libCard(me.ID, "Mountain", "Basic Land — Mountain", "")
	deep := libCard(me.ID, "Never Reached", "Instant", "{U}")
	libraryOf(t, me, deep, hit, same, land)

	exileUntilResolving(t, g, me.ID, g.Seats[0].ID, nonlandNotNamed("Same Name"), false)
	c := resolutionCastPrompt(g, me.ID)
	if c == nil || len(c.DiscardOptions) != 1 || c.DiscardOptions[0] != hit.InstanceID {
		t.Fatalf("prompt = %+v, want the Divination for the walked player", c)
	}
	if c.MayCastKeyword != MayCastKeywordExileUntil {
		t.Errorf("keyword = %q", c.MayCastKeyword)
	}
	if len(c.CastPile) != 2 || !g.ResolutionPaused() {
		t.Errorf("pile %d, paused %v: want the two passed-over cards on a prompt holding the resolution", len(c.CastPile), g.ResolutionPaused())
	}
	if !me.Library.Contains(deep.InstanceID) {
		t.Errorf("the walk went past its hit")
	}
}

// The hit carries an uncapped, free, flash-timed grant, and is cast
// during the resolution: on the stack before it finishes, with the pile
// on the bottom.
func TestExileUntilCastCastsTheHitBeforeTheResolutionFinishes(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	hit := libCard(me.ID, "Emrakul", "Legendary Creature — Eldrazi", "{15}")
	land := libCard(me.ID, "Mountain", "Basic Land — Mountain", "")
	libraryOf(t, me, hit, land)

	exileUntilResolving(t, g, me.ID, hit.InstanceID, nonlandNotNamed("x"), false)
	perm := g.CastPermissionOnCardByIDForEffect(hit.InstanceID)
	if perm == nil || perm.Cost != "{0}" || perm.Timing != TimingFlash || !perm.CastOnly ||
		perm.MaxSpellManaValue != nil || perm.LapseOnPass != LapseToLibraryBottom {
		t.Fatalf("grant = %+v", perm)
	}
	if me.Library.Contains(land.InstanceID) {
		t.Error("the pile went to the bottom before the answer")
	}
	answerResolutionCast(t, g, me.ID, true)
	if !g.Stack.Contains(hit.InstanceID) || g.ResolutionPaused() {
		t.Fatal("the hit is not on the stack, or the resolution is still held")
	}
	if !me.Library.Contains(land.InstanceID) || len(g.Exile.Cards) != 0 {
		t.Errorf("the passed-over land should be on the bottom and nothing left in exile")
	}
}

// Declining puts the hit on the bottom with the rest. No hit at all
// puts everything back and asks nothing.
func TestExileUntilCastDeclineAndNoHitBottomEverything(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	hit := libCard(me.ID, "Divination", "Sorcery", "{2}{U}")
	land := libCard(me.ID, "Mountain", "Basic Land — Mountain", "")
	libraryOf(t, me, hit, land)
	exileUntilResolving(t, g, me.ID, hit.InstanceID, nonlandNotNamed("x"), false)
	answerResolutionCast(t, g, me.ID, false)
	if len(g.Exile.Cards) != 0 || me.Library.Size() != 2 {
		t.Errorf("declined: exile %d, library %d, want 0 and 2", len(g.Exile.Cards), me.Library.Size())
	}
	if g.CastPermissionOnCardByIDForEffect(hit.InstanceID) != nil || g.ResolutionPaused() {
		t.Error("the decline left the grant, or the resolution held")
	}

	libraryOf(t, me, land)
	exileUntilResolving(t, g, me.ID, hit.InstanceID, nonlandNotNamed("x"), false)
	if resolutionCastPrompt(g, me.ID) != nil {
		t.Fatalf("prompted with no hit")
	}
	if len(g.Exile.Cards) != 0 || me.Library.Size() != 1 {
		t.Errorf("no hit: exile %d, library %d, want 0 and 1", len(g.Exile.Cards), me.Library.Size())
	}
}

// #4205: the keeping form leaves a declined hit in exile and bottoms the
// rest of the pile; the base form (cascade's) still bottoms the hit.
func TestExileUntilCastKeepingLeavesADeclinedHitInExile(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	hit := libCard(me.ID, "Divination", "Sorcery", "{2}{U}")
	land := libCard(me.ID, "Mountain", "Basic Land — Mountain", "")
	libraryOf(t, me, hit, land)
	exileUntilResolving(t, g, me.ID, hit.InstanceID, nonlandNotNamed("x"), true)
	c := resolutionCastPrompt(g, me.ID)
	if c == nil || c.DeclineLabel != "Leave it in exile" {
		t.Fatalf("prompt = %+v", c)
	}
	answerResolutionCast(t, g, me.ID, false)
	if !g.Exile.Contains(hit.InstanceID) || me.Library.Contains(hit.InstanceID) {
		t.Errorf("the declined hit should stay in exile")
	}
	if !me.Library.Contains(land.InstanceID) || g.Exile.Contains(land.InstanceID) {
		t.Errorf("the passed-over land should be on the bottom of the library")
	}
	if g.CastPermissionOnCardByIDForEffect(hit.InstanceID) != nil {
		t.Errorf("the grant outlived the decline")
	}
}

// Outside a resolution there is nothing to hold open: the keeping form's
// grant alone is the window, and passing without casting leaves the hit
// in exile (LapseStaysInExile).
func TestExileUntilCastKeepingUncastHitStaysInExileOnThePass(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	hit := libCard(me.ID, "Counterspell", "Instant", "{U}{U}")
	land := libCard(me.ID, "Mountain", "Basic Land — Mountain", "")
	libraryOf(t, me, hit, land)
	g.WithWriteLock(func() {
		_ = g.ExileUntilCastFreeKeepingForEffect(me.ID, hit.InstanceID, nonlandNotNamed("x"))
	})
	if resolutionCastPrompt(g, me.ID) != nil {
		t.Fatal("no resolution is open, so nothing is asked")
	}
	perm := g.CastPermissionOnCardByIDForEffect(hit.InstanceID)
	if perm == nil || perm.LapseOnPass != LapseStaysInExile {
		t.Fatalf("grant = %+v", perm)
	}
	if !me.Library.Contains(land.InstanceID) {
		t.Errorf("the passed-over land should be on the bottom")
	}
	if g.Seats[g.Turn.PriorityHolder].ID != me.ID {
		t.Fatalf("test needs the caster to hold priority")
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	if !g.Exile.Contains(hit.InstanceID) || me.Library.Contains(hit.InstanceID) {
		t.Errorf("an uncast hit should stay in exile after the pass")
	}
	if g.CastPermissionOnCardByIDForEffect(hit.InstanceID) != nil {
		t.Errorf("the grant should have closed on the pass")
	}
}
