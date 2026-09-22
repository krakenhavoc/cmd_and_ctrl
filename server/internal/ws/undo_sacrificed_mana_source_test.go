package ws

import (
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// undo_sacrificed_mana_source_test.go — #1215. The auto-tapper can now
// crack a Treasure to pay for a cast, which makes the auto-tap
// payment the first one that DESTROYS a permanent. A take-back has to
// put it back: the undo stack is a pre-mutation clone of the whole
// game, so this is a claim about the clone covering a battlefield exit
// the tapper never used to make, not about a new undo path.

// treasuresAndASpell seeds a room at precombat main with `n` Treasures
// on the battlefield and one `cost`-priced spell in the active seat's
// hand. Returns the room, the game, the seat, the Treasure IDs and
// the spell.
func treasuresAndASpell(t *testing.T, n int, cost string) (*Room, *game.Game, uuid.UUID, []uuid.UUID, uuid.UUID) {
	t.Helper()
	g := seedTestGame(t)
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	seat := g.Seats[0].ID
	treasures, err := g.SpawnCards(seat, seat, game.ZoneBattlefield, effects.TreasureToken(), n)
	if err != nil {
		t.Fatalf("spawn Treasures: %v", err)
	}
	spells, err := g.SpawnCards(seat, seat, game.ZoneHand, game.Card{
		Name: "Icy Manipulator", TypeLine: "Artifact", ManaCost: cost,
	}, 1)
	if err != nil {
		t.Fatalf("spawn spell: %v", err)
	}
	room := NewRoom(g, slog.New(slog.NewTextHandler(io.Discard, nil)), t.TempDir())
	return room, g, seat, treasures, spells[0]
}

func onBattlefield(g *game.Game, id uuid.UUID) bool {
	var found bool
	g.ReadSnapshot(func() { found = g.Battlefield.Contains(id) })
	return found
}

// The whole claim in one test: the cast eats two Treasures, the undo
// hands them back untapped, and the spell is off the stack.
func TestUndoRestoresATreasureTheAutoTapperCracked(t *testing.T) {
	room, g, seat, treasures, spell := treasuresAndASpell(t, 2, "{2}")

	if _, _, err := room.Apply(seat, func() error {
		return g.CastSpell(seat, spell, game.CastSpellParams{Strict: true, AutoTap: true, FromZone: "hand"})
	}); err != nil {
		t.Fatalf("auto-tap cast off two Treasures: %v", err)
	}
	for _, id := range treasures {
		if onBattlefield(g, id) {
			t.Fatalf("Treasure %v survived the cast it paid for", id)
		}
	}

	if _, _, err := room.Undo(uuid.Nil); err != nil {
		t.Fatalf("Undo: %v", err)
	}
	for _, id := range treasures {
		if !onBattlefield(g, id) {
			t.Errorf("Treasure %v was not restored by the undo", id)
		}
	}
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == treasures[0] && c.Tapped {
				t.Error("the restored Treasure came back tapped")
			}
		}
		if g.Stack.Contains(spell) {
			t.Error("the undone spell is still on the stack")
		}
		if p := g.PlayerByIDForEffect(seat); p != nil && !p.Hand.Contains(spell) {
			t.Error("the undone spell did not go back to hand")
		}
	})
}
