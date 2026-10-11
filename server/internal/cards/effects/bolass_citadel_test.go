package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// bolass_citadel_test.go — #2955: "{T}, Sacrifice ten nonland
// permanents: Each opponent loses 10 life." The top-of-library
// permission and the life payment are pinned in
// granted_permissions_test.go.

func citadelBoard(g *game.Game, me uuid.UUID, nonland, lands int) (citadel uuid.UUID, others []uuid.UUID) {
	citadel = pushCatalogPermanent(g, me, "Bolas's Citadel", "Legendary Artifact", citadelOracle, false)
	for i := 0; i < nonland; i++ {
		others = append(others, pushCatalogPermanent(g, me, "Rock", "Artifact", "", false))
	}
	for i := 0; i < lands; i++ {
		others = append(others, pushCatalogPermanent(g, me, "Forest", "Basic Land — Forest", "", false))
	}
	return citadel, others
}

func TestCitadelNeedsTenEligiblePermanents(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	citadel, others := citadelBoard(g, me.ID, 8, 5) // 9 nonland with the Citadel, plus 5 lands

	nine := append([]uuid.UUID{citadel}, others[:8]...)
	if err := g.ActivateCatalogAbility(me.ID, citadel, 0, game.ActivateAbilityParams{SacrificeIDs: nine}); err == nil {
		t.Fatal("nine permanents paid a ten-permanent cost")
	}
	// Lands are not eligible, so nine nonland plus a land is refused too.
	withLand := append(append([]uuid.UUID{}, nine...), others[8])
	if err := g.ActivateCatalogAbility(me.ID, citadel, 0, game.ActivateAbilityParams{SacrificeIDs: withLand}); err == nil {
		t.Fatal("a land paid part of the cost")
	}
	for _, id := range append(others, citadel) {
		if !g.Battlefield.Contains(id) {
			t.Errorf("%s left on a refused activation", cardNameOf(g, id))
		}
	}
	if b16Tapped(t, g, citadel) {
		t.Error("a refused activation tapped the Citadel")
	}
	if len(g.StackMeta) != 0 {
		t.Error("a refused activation used the stack")
	}
}

func TestCitadelPaysTenAndTapBeforeResolving(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	citadel, others := citadelBoard(g, me.ID, 12, 2)
	ten := others[:10]
	spare := others[10:] // two Rocks and two Forests stay

	lives := map[uuid.UUID]int{}
	for _, s := range g.Seats {
		lives[s.ID] = s.Life
	}
	if err := g.ActivateCatalogAbility(me.ID, citadel, 0, game.ActivateAbilityParams{SacrificeIDs: ten}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	for _, id := range ten {
		if g.Battlefield.Contains(id) {
			t.Errorf("%s was not paid", cardNameOf(g, id))
		}
	}
	for _, id := range append(spare, citadel) {
		if !g.Battlefield.Contains(id) {
			t.Errorf("%s was sacrificed but not named", cardNameOf(g, id))
		}
	}
	if !b16Tapped(t, g, citadel) {
		t.Error("the Citadel is not tapped")
	}
	if got := countEventsOf(g, game.EventSacrifice, ten...); got != 10 {
		t.Errorf("%d sacrifice events, want 10", got)
	}
	for _, s := range g.Seats {
		if s.Life != lives[s.ID] {
			t.Errorf("%s's life moved before the ability resolved", s.ID)
		}
	}
	passPriorityAroundTable(t, g)
	for _, s := range g.Seats {
		want := lives[s.ID]
		if s.ID != me.ID {
			want -= 10
		}
		if s.Life != want {
			t.Errorf("seat %s life = %d, want %d", s.ID, s.Life, want)
		}
	}
}

// Nothing says "other": the Citadel may be one of its own ten, and the
// ability still resolves with the source gone.
func TestCitadelMayBeAmongTheTen(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	citadel, others := citadelBoard(g, me.ID, 9, 0)
	ten := append([]uuid.UUID{citadel}, others...)
	opp := g.Seats[1]
	before := opp.Life

	if err := g.ActivateCatalogAbility(me.ID, citadel, 0, game.ActivateAbilityParams{SacrificeIDs: ten}); err != nil {
		t.Fatalf("activate with the source among the ten: %v", err)
	}
	if g.Battlefield.Contains(citadel) {
		t.Error("the Citadel stayed")
	}
	passPriorityAroundTable(t, g)
	if opp.Life != before-10 {
		t.Errorf("life %d -> %d, want -10", before, opp.Life)
	}
}

// The cost is paid, so answering by removing the Citadel does not stop it.
func TestCitadelAbilityResolvesAfterTheCitadelIsRemoved(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	citadel, others := citadelBoard(g, me.ID, 10, 0)
	before := g.Seats[1].Life

	if err := g.ActivateCatalogAbility(me.ID, citadel, 0, game.ActivateAbilityParams{SacrificeIDs: others}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	g.WithWriteLock(func() { _ = g.ExileCardsForEffect([]uuid.UUID{citadel}) })
	passPriorityAroundTable(t, g)
	for _, s := range g.Seats {
		if s.ID != me.ID && s.Life != before-10 {
			t.Errorf("seat %s life %d, want %d", s.ID, s.Life, before-10)
		}
	}
}

func TestCitadelCannotBeActivatedTwiceInATurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	citadel, others := citadelBoard(g, me.ID, 20, 0)
	if err := g.ActivateCatalogAbility(me.ID, citadel, 0, game.ActivateAbilityParams{SacrificeIDs: others[:10]}); err != nil {
		t.Fatalf("first: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, citadel, 0, game.ActivateAbilityParams{SacrificeIDs: others[10:]}); err == nil {
		t.Error("a tapped Citadel activated again")
	}
}
