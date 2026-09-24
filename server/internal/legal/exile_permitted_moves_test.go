package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// exile_permitted_moves_test.go — the enumerator half of #1573. A card
// Gonti, Night Minister or Outrageous Robbery exiles face down is a
// cast surface for its permission holder alone, and "mana of any TYPE"
// makes a {C} in its cost payable off Swamps. The enumerator prices
// through the same pricer the cast path pays with, so an any-colour
// grant over the same card offers nothing on the same board, and
// dispatchAll proves the offered cast is one the engine accepts.
func TestEnumeratorOffersAFaceDownPermittedCardPaidWithAnyType(t *testing.T) {
	for _, tc := range []struct {
		name  string
		perm  game.CastPermission
		moves int
	}{
		{"any type", game.CastPermission{AnyType: true, Duration: game.WhileInZoneDuration()}, 1},
		{"any color", game.CastPermission{AnyColor: true, Duration: game.WhileInZoneDuration()}, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newTable(t)
			seat := g.Seats[g.Turn.ActiveSeat]
			victim := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
			clearHand(seat)
			advanceTo(t, g, game.StepPrecombatMain)
			for i := 0; i < 3; i++ {
				g.Battlefield.PushTop(game.Card{
					InstanceID: uuid.New(), Name: "Swamp", TypeLine: "Basic Land — Swamp",
					Owner: seat.ID, Controller: seat.ID,
				})
			}
			loot := game.NewCard("Enum Stolen Seer", victim.ID)
			loot.TypeLine = "Creature — Eldrazi"
			loot.ManaCost = "{2}{C}"
			victim.Library.PushTop(loot)
			g.WithWriteLock(func() {
				if err := g.ExileTopFaceDownWithPermissionForEffect(victim.ID, seat.ID, 1, tc.perm); err != nil {
					t.Fatalf("ExileTopFaceDownWithPermissionForEffect: %v", err)
				}
			})

			moves := legal.EnumerateFor(g, seat.ID)
			if got := castMovesFor(moves, loot.InstanceID); len(got) != tc.moves {
				t.Fatalf("holder offered %d casts, want %d: %v", len(got), tc.moves, labels(moves))
			}
			for _, p := range g.Seats {
				if p.ID == seat.ID {
					continue
				}
				if got := castMovesFor(legal.EnumerateFor(g, p.ID), loot.InstanceID); len(got) != 0 {
					t.Errorf("seat %s was offered the face-down card", p.Name)
				}
			}
			dispatchAll(t, g, seat.ID, castMovesFor(moves, loot.InstanceID))
		})
	}
}
