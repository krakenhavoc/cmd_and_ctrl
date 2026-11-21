package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// duration_copy_test.go — #1593's cards end to end: a permanent already
// on the battlefield becomes a copy of something until end of turn (or
// for good), takes on the copied card's catalog behaviour while it is
// the copy, and gives it back when the copy ends. The engine shape is
// pinned in server/internal/game/duration_copy_test.go.

const (
	OracleMirageMirror         = "6a84eda1-7e72-4b5a-849f-b6e177565aeb"
	OracleCytoshape            = "14221c18-7801-49c9-a2c8-53b8c5181d63"
	OracleMirrorweave          = "026b7221-0caf-4c8b-8c1b-e7de836797b7"
	OracleUnstableShapeshifter = "6a71deb3-6659-47ea-8687-136cf75fa381"
	OracleLazavDimirMastermind = "8027a610-613e-4640-9840-c8778694f312"
	OracleZulaportCutthroat    = "76b003e0-15af-4f22-bdf2-1ade5430964a"
)

// dcSeed puts a permanent on the battlefield through the zone-move
// event, so it carries an entry stamp like a real one.
func DcSeed(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, p, tough int) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       name,
		TypeLine:   typeLine,
		OracleID:   oracle,
		ManaCost:   "{2}",
		Power:      p,
		Toughness:  tough,
		Owner:      owner,
		Controller: owner,
	})
}

// dcCard reads a battlefield permanent after settling the layer pass.
func DcCard(t *testing.T, g *game.Game, id uuid.UUID) game.Card {
	t.Helper()
	var out game.Card
	found := false
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				out, found = c, true
			}
		}
	})
	if !found {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return out
}

// dcCleanup runs the CR 514.2 sweep the cleanup step runs.
func DcCleanup(g *game.Game) {
	g.WithWriteLock(func() { g.ClearEndOfTurnScopedStaticsLocked() })
}

func DcMana(t *testing.T, g *game.Game, player uuid.UUID, mana string) {
	t.Helper()
	var err error
	g.WithWriteLock(func() { err = g.AddManaForEffect(player, uuid.Nil, mana) })
	if err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
}

// --- Shifting Woodland ----------------------------------------------

func DcGraveyardCard(g *game.Game, owner *game.Player, name, typeLine, oracle string, p, tough int) uuid.UUID {
	id := uuid.New()
	owner.Graveyard.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: p, Toughness: tough, Owner: owner.ID, Controller: owner.ID,
	})
	return id
}

// --- Cytoshape -------------------------------------------------------

// dcAnswerChooseCards answers the open ChooseCards prompt with `pick`.
func DcAnswerChooseCards(t *testing.T, g *game.Game, pick uuid.UUID) *game.PendingChoice {
	t.Helper()
	for i := 0; i < 16; i++ {
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceChooseCards {
				if err := g.ResolveChooseCards(c.ID, c.Chooser, []uuid.UUID{pick}); err != nil {
					t.Fatalf("ResolveChooseCards: %v", err)
				}
				return c
			}
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("no ChooseCards prompt was queued")
	return nil
}

// --- Unstable Shapeshifter -------------------------------------------

// dcEnter puts a creature onto the battlefield with the zone-move AND
// ETB events a real entry emits.
func DcEnter(t *testing.T, g *game.Game, owner uuid.UUID, name, typeLine string, p, tough int) uuid.UUID {
	t.Helper()
	id := SeedCopyableCreature(g, owner, name, typeLine, p, tough)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventETB, CardID: id, Actor: owner})
	})
	return id
}

// --- Lazav, Familiar Stranger -----------------------------------------

const OracleLazavFamiliarStranger = "ca60623a-1a98-4153-a0ee-2f34981de5cb"
