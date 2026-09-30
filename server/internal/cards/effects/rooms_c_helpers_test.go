package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// rooms_c_helpers_test.go — shared scaffolding for the ADR 0103 PR 3
// Room card tests of group C: a real Room card in hand, a real cast of
// either half, a real unlock.

// roomCardC is a Room as the importer would hand it over: one split
// card, two Enchantment — Room faces, keyed on the catalog's oracle ID.
func roomCardC(owner uuid.UUID, oracle, leftName, leftCost, rightName, rightCost string, colors ...string) game.Card {
	c := game.Card{
		InstanceID: uuid.New(),
		OracleID:   oracle,
		Owner:      owner,
		Controller: owner,
		Layout:     game.LayoutSplit,
		Faces: []game.Face{
			{Name: leftName, TypeLine: "Enchantment — Room", ManaCost: leftCost, Colors: colors},
			{Name: rightName, TypeLine: "Enchantment — Room", ManaCost: rightCost, Colors: colors},
		},
	}
	c.SettleImported()
	return c
}

// castRoomC casts one half (`face` 0 or 1) of `c` from `me`'s hand; the
// caller settles the stack (a targeted door trigger asks for its pick first).
func castRoomC(t *testing.T, g *game.Game, me uuid.UUID, c game.Card, face int) {
	t.Helper()
	g.WithWriteLock(func() { g.PlayerByIDForEffect(me).Hand.PushTop(c) })
	if err := g.CastSpell(me, c.InstanceID, game.CastSpellParams{Face: face}); err != nil {
		t.Fatalf("cast face %d of %s: %v", face, c.Faces[face].Name, err)
	}
}

// unlockDoorC takes the CR 709.5e unlock special action on `door`; the
// caller settles the stack.
func unlockDoorC(t *testing.T, g *game.Game, me, room uuid.UUID, door game.DoorSide) {
	t.Helper()
	if err := g.PerformSpecialAction(me, room, game.SpecialActionUnlock, game.SpecialActionParams{Door: door}); err != nil {
		t.Fatalf("unlock door %v: %v", door, err)
	}
}

// sizeOfC is a battlefield permanent's current power and toughness.
func sizeOfC(t *testing.T, g *game.Game, id uuid.UUID) (int, int) {
	t.Helper()
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		t.Fatalf("%s is not in play", id)
	}
	return c.CurrentPower(), c.CurrentToughness()
}
