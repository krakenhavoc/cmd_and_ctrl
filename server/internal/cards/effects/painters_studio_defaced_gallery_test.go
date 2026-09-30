package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const paintersStudioOracle = "2af01e09-f233-4675-89cd-470816471c62"

func paintersStudioCard(owner uuid.UUID) game.Card {
	return roomsDCard(owner, paintersStudioOracle, "Painter's Studio", "{2}{R}", "Defaced Gallery", "{1}{R}", "R")
}

// Painter's Studio's unlock exiles the top two cards with a live "play"
// permission (Reckless Impulse's primitive).
func TestPaintersStudioExilesTheTopTwoToPlay(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	ids := seedSearchLibrary(me,
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Shock", TypeLine: "Instant"},
		game.Card{Name: "Deep Card", TypeLine: "Sorcery"},
	)
	forest, shock, deep := ids[0], ids[1], ids[2]

	roomsDCast(t, g, me, paintersStudioCard(me.ID), 0)

	if !g.Exile.Contains(forest) || !g.Exile.Contains(shock) || !me.Library.Contains(deep) {
		t.Fatal("the top two cards were not exiled")
	}
	for _, id := range []uuid.UUID{forest, shock} {
		perm := exiledPermission(g, id)
		if perm.Player != me.ID || perm.CastOnly || !permissionLive(g, perm, me.ID) {
			t.Errorf("%+v: the controller may play it now", perm)
		}
	}
}

// attackWithBears casts the Room half, then attacks with two creatures
// and returns their current powers after the trigger resolved.
func attackWithBears(t *testing.T, g *game.Game, me, opp *game.Player, face int) (int, int) {
	t.Helper()
	advanceToMain(t, g)
	a := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	b := b16Creature(g, me.ID, "Ox", "Creature — Ox", 3, 3, "G")
	idle := b16Creature(g, me.ID, "Idle", "Creature — Wall", 1, 1, "G")
	roomsDCast(t, g, me, paintersStudioCard(me.ID), face)
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{a, b} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatal(err)
		}
	}
	lockInAttacks(t, g)
	roomsDSettle(t, g, me.ID)
	if got := b31CurrentPowerOf(g, idle); got != 1 {
		t.Errorf("a creature that did not attack has power %d, want 1", got)
	}
	return b31CurrentPowerOf(g, a), b31CurrentPowerOf(g, b)
}

// Defaced Gallery: one trigger for the whole attack, +1/+0 to each
// attacking creature you control.
func TestDefacedGalleryPumpsEachAttacker(t *testing.T) {
	g := newCatalogGame(t)
	bear, ox := attackWithBears(t, g, g.Seats[0], g.Seats[1], 1)
	if bear != 3 || ox != 4 {
		t.Fatalf("attackers' power = %d and %d, want 3 and 4 (one +1/+0 each)", bear, ox)
	}
}

// With only Painter's Studio unlocked the Gallery door is locked and
// nothing happens on an attack.
func TestDefacedGalleryLockedDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	bear, ox := attackWithBears(t, g, g.Seats[0], g.Seats[1], 0)
	if bear != 2 || ox != 3 {
		t.Fatalf("attackers' power = %d and %d, want 2 and 3 (door locked)", bear, ox)
	}
}
