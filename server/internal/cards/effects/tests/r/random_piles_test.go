package testsr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// random_piles_test.go — #6248, Whims of the Fates: each player
// separates their permanents into three piles, one of each player's
// piles is picked at random, and those permanents are sacrificed.
func TestEachPlayerSacrificesARandomPile(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	push := func(p *game.Player, name string) uuid.UUID {
		return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: name, TypeLine: "Artifact", Owner: p.ID, Controller: p.ID,
		})
	}
	mine := []uuid.UUID{push(me, "Mine A"), push(me, "Mine B"), push(me, "Mine C")}
	theirs := []uuid.UUID{push(opp, "Theirs A"), push(opp, "Theirs B"), push(opp, "Theirs C")}
	g.WithWriteLock(func() {
		if err := effects.EachPlayerSacrificesARandomPile(3, "Whims of the Fates")(g, &game.StackItem{Controller: me.ID}); err != nil {
			t.Fatal(err)
		}
	})
	answer := func(pick uuid.UUID) {
		t.Helper()
		var c *game.PendingChoice
		g.WithWriteLock(func() {
			for _, p := range g.PendingChoices {
				if p.Kind == game.PendingChoiceChooseCards {
					c = p
				}
			}
		})
		if c == nil {
			t.Fatal("no pile prompt is open")
		}
		if err := g.ResolveChooseCards(c.ID, c.Chooser, []uuid.UUID{pick}); err != nil {
			t.Fatal(err)
		}
	}
	answer(mine[0])
	answer(mine[1])
	answer(theirs[0])
	answer(theirs[1])
	left := func(ids []uuid.UUID) int {
		n := 0
		g.WithWriteLock(func() {
			for _, id := range ids {
				for i := range g.Battlefield.Cards {
					if g.Battlefield.Cards[i].InstanceID == id {
						n++
					}
				}
			}
		})
		return n
	}
	if got := left(mine); got != 2 {
		t.Errorf("I have %d of my 3 permanents, want 2 (one pile of one sacrificed)", got)
	}
	if got := left(theirs); got != 2 {
		t.Errorf("they have %d of their 3 permanents, want 2", got)
	}
	if n := len(me.Graveyard.Cards) + len(opp.Graveyard.Cards); n != 2 {
		t.Errorf("%d cards in graveyards, want the two sacrificed", n)
	}
}
