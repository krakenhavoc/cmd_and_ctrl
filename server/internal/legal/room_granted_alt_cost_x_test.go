package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// #3724: an {X} spell and a spell with no mana cost are Access Maze
// moves too.
func TestEnumeratorOffersAccessMazeForXAndManalessSpells(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	battlefieldCard(g, seat, accessMazeRoom(game.DoorRightUnlocked))
	key := effects.GrantedAltCostLifeFromHandOnceYourTurn
	xSpell := handCard(seat, game.Card{Name: "X Sorcery", TypeLine: "Sorcery", ManaCost: "{X}{B}{B}"})
	manaless := handCard(seat, game.Card{Name: "Ancestral Vision", Layout: "normal", TypeLine: "Sorcery"})

	moves := legal.EnumerateFor(g, seat.ID)
	for _, id := range []uuid.UUID{xSpell, manaless} {
		casts := castPayloadsOf(t, moves, id)
		found := false
		for _, c := range casts {
			found = found || c.AlternativeCost == key
		}
		if !found {
			t.Fatalf("no Access Maze cast offered: %+v (%v)", casts, labels(moves))
		}
		dispatchAll(t, g, seat.ID, castMovesFor(moves, id))
	}
}
