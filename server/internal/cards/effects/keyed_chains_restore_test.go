package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// keyed_chains_restore_test.go — #4290: chains of keyed sacrifice runs,
// each continuation asking the next.

// Tomb of Annihilation's Oubliette asks for a creature, then an artifact,
// then a land, each its own keyed run whose continuation asks the next:
// every question is a restore point.
func TestTombOublietteSacrificeChainIsARestorePointAtEveryStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	source := pushPermanentOf(g, me.ID, "Tomb of Annihilation", "Dungeon")
	creatures := []uuid.UUID{pushVanillaCreature(g, me.ID, "Creature A", 1, 1), pushVanillaCreature(g, me.ID, "Creature B", 1, 1)}
	artifacts := []uuid.UUID{pushPermanentOf(g, me.ID, "Artifact A", "Artifact"), pushPermanentOf(g, me.ID, "Artifact B", "Artifact")}
	lands := []uuid.UUID{pushPermanentOf(g, me.ID, "Land A", "Basic Land — Forest"), pushPermanentOf(g, me.ID, "Land B", "Basic Land — Forest")}
	g.WithWriteLock(func() {
		item := &game.StackItem{Kind: game.StackItemTriggered, Controller: me.ID, Owner: me.ID, SourceCardID: source}
		if err := tombOubliette(g, item); err != nil {
			t.Fatalf("tombOubliette: %v", err)
		}
	})
	for i, step := range []struct {
		pick uuid.UUID
		key  string
	}{
		{creatures[0], oublietteSacrificed[0].Key()},
		{artifacts[0], oublietteSacrificed[1].Key()},
		{lands[0], oublietteSacrificed[2].Key()},
	} {
		g = restoreRoundTrip(t, g, true)
		me = g.Seats[0]
		c := sacrificeChoiceFor(g, me.ID)
		if c == nil || c.PickThen != step.key {
			t.Fatalf("step %d: the prompt is %+v, want a leg of %s", i, c, step.key)
		}
		answerSacrifice(t, g, me.ID, step.pick)
		if g.Battlefield.Contains(step.pick) {
			t.Errorf("step %d: the chosen permanent was not sacrificed on the restored table", i)
		}
	}
	for _, kept := range []uuid.UUID{creatures[1], artifacts[1], lands[1]} {
		if !g.Battlefield.Contains(kept) {
			t.Errorf("%s was not chosen and should have stayed", kept)
		}
	}
	if sacrificeChoiceFor(g, me.ID) != nil {
		t.Error("a fourth question was asked")
	}
}

// pushPermanentOf puts a typed permanent on the battlefield, taking the
// lock itself.
func pushPermanentOf(g *game.Game, owner uuid.UUID, name, typeLine string) uuid.UUID {
	var id uuid.UUID
	g.WithWriteLock(func() { id = pushNamedPermanent(g, owner, name, typeLine) })
	return id
}
