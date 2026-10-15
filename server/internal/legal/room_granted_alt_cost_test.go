package legal_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// room_granted_alt_cost_test.go — #3174, the enumerator half of Access
// Maze: a granted offer behind a Room door, once during each of its
// controller's turns, for life equal to the spell's mana value. The
// moves walk game.CastOffersForLocked, so the life cast is a move
// exactly while the door is unlocked and the claim is unspent, and
// every one is a move CastSpell accepts.

const oracleCrampedVentsAccessMaze = "1e847d69-e527-4bb9-a624-17c25ac4fed4"

func accessMazeRoom(unlocked game.DoorMask) game.Card {
	c := game.Card{
		OracleID: oracleCrampedVentsAccessMaze,
		Layout:   game.LayoutSplit,
		Faces: []game.Face{
			{Name: "Cramped Vents", TypeLine: "Enchantment — Room", ManaCost: "{3}{B}"},
			{Name: "Access Maze", TypeLine: "Enchantment — Room", ManaCost: "{5}{B}{B}"},
		},
	}
	c.SettleImported()
	c.Unlocked = unlocked
	return c
}

func TestEnumeratorOffersAccessMazeOnlyWhileUnlockedAndOncePerTurn(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	room := battlefieldCard(g, seat, accessMazeRoom(game.DoorLeftUnlocked))
	spell := handCard(seat, game.Card{Name: "Test Sorcery", TypeLine: "Sorcery", ManaCost: "{2}{B}{B}"})
	key := effects.GrantedAltCostLifeFromHandOnceYourTurn

	// No mana, the door locked: no cast of any kind.
	if casts := castPayloadsOf(t, legal.EnumerateFor(g, seat.ID), spell); len(casts) != 0 {
		t.Fatalf("a cast was offered with no mana behind a locked door: %+v", casts)
	}

	g.WithWriteLock(func() {
		if err := g.UnlockDoorForEffect(room, game.DoorRight, seat.ID); err != nil {
			t.Fatalf("unlock: %v", err)
		}
	})
	moves := legal.EnumerateFor(g, seat.ID)
	casts := castPayloadsOf(t, moves, spell)
	if len(casts) != 1 || casts[0].AlternativeCost != key {
		t.Fatalf("casts behind an unlocked door = %+v, want one claiming Access Maze: %v", casts, labels(moves))
	}
	dispatchAll(t, g, seat.ID, castMovesFor(moves, spell))

	// Not when the life would kill the seat (the enumerator's policy).
	seat.Life = 4
	if casts := castPayloadsOf(t, legal.EnumerateFor(g, seat.ID), spell); len(casts) != 0 {
		t.Fatalf("a cast for 4 life was offered at 4 life: %+v", casts)
	}
	seat.Life = 40

	// Spend the claim: the next spell has no move.
	if err := g.CastSpell(seat.ID, spell, game.CastSpellParams{Strict: true, AlternativeCost: key}); err != nil {
		t.Fatalf("life cast: %v", err)
	}
	other := handCard(seat, game.Card{Name: "Test Sorcery Two", TypeLine: "Sorcery", ManaCost: "{1}{B}"})
	if casts := castPayloadsOf(t, legal.EnumerateFor(g, seat.ID), other); len(casts) != 0 {
		t.Fatalf("a second life cast was offered this turn: %+v", casts)
	}
}
