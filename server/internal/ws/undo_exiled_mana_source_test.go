package ws

import (
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// undo_exiled_mana_source_test.go — #1228. The auto-tapper can now
// spend a card OUT OF HAND to pay for a cast, which makes the auto-tap
// payment the first one that takes a card off a player's hand. A
// take-back has to put it back: the undo stack is a pre-mutation clone
// of the whole game, so this is a claim about the clone covering a
// HAND exit the tapper never used to make — the same claim #1215 made
// about a battlefield one.

// spiritGuidesAndASpell seeds a room at precombat main with `n` Spirit
// Guides in the active seat's hand and one `cost`-priced spell beside
// them. Nothing is on the battlefield, so the Guides are the only way
// to pay.
func spiritGuidesAndASpell(t *testing.T, n int, cost string) (*Room, *game.Game, uuid.UUID, []uuid.UUID, uuid.UUID) {
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
	guides, err := g.SpawnCards(seat, seat, game.ZoneHand, game.Card{
		Name:     "Simian Spirit Guide",
		TypeLine: "Creature — Ape Spirit",
		ManaCost: "{2}{R}",
		Power:    2, Toughness: 2,
		ManaAbilities: []game.ManaAbilityShape{{
			Zones:     []game.ZoneKind{game.ZoneHand},
			ExileSelf: true,
			Produced:  "{R}",
			Label:     "Exile this card from your hand: Add {R}",
		}},
	}, n)
	if err != nil {
		t.Fatalf("spawn Spirit Guides: %v", err)
	}
	spells, err := g.SpawnCards(seat, seat, game.ZoneHand, game.Card{
		Name: "Icy Manipulator", TypeLine: "Artifact", ManaCost: cost,
	}, 1)
	if err != nil {
		t.Fatalf("spawn spell: %v", err)
	}
	room := NewRoom(g, slog.New(slog.NewTextHandler(io.Discard, nil)), t.TempDir())
	return room, g, seat, guides, spells[0]
}

func inHand(g *game.Game, seat, id uuid.UUID) bool {
	var found bool
	g.ReadSnapshot(func() {
		if p := g.PlayerByIDForEffect(seat); p != nil {
			found = p.Hand.Contains(id)
		}
	})
	return found
}

// The whole claim in one test: the cast exiles two Spirit Guides, the
// undo hands them back to the hand, and the spell is off the stack.
func TestUndoRestoresASpiritGuideTheAutoTapperExiled(t *testing.T) {
	room, g, seat, guides, spell := spiritGuidesAndASpell(t, 2, "{2}")

	if _, _, err := room.Apply(seat, func() error {
		return g.CastSpell(seat, spell, game.CastSpellParams{Strict: true, AutoTap: true, FromZone: "hand"})
	}); err != nil {
		t.Fatalf("auto-tap cast off two Spirit Guides: %v", err)
	}
	for _, id := range guides {
		if inHand(g, seat, id) {
			t.Fatalf("Spirit Guide %v stayed in hand through the cast it paid for", id)
		}
	}

	if _, _, err := room.Undo(uuid.Nil); err != nil {
		t.Fatalf("Undo: %v", err)
	}
	for _, id := range guides {
		if !inHand(g, seat, id) {
			t.Errorf("Spirit Guide %v was not restored to hand by the undo", id)
		}
	}
	g.ReadSnapshot(func() {
		if g.Exile.Contains(guides[0]) {
			t.Error("the restored Spirit Guide is still in exile as well")
		}
		if g.Stack.Contains(spell) {
			t.Error("the undone spell is still on the stack")
		}
		if p := g.PlayerByIDForEffect(seat); p != nil && !p.Hand.Contains(spell) {
			t.Error("the undone spell did not go back to hand")
		}
	})
}
