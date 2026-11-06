package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// omen_moves_test.go — the enumerator half of CR 720 (#5175). The
// enumerator walks the same face gate as the engine and the view, so it
// offers an omen card's Omen half from hand (CR 720.3), and once that
// Omen has resolved there is nothing left to offer: the card is in its
// owner's library (CR 720.3d), not in exile under a grant.

// omenInHand seats a two-faced omen card, laid out as Scryfall does
// (`adventure`, an Omen sorcery half), in a player's hand. {0} on both
// halves so affordability is not under test.
func omenInHand(p *game.Player) uuid.UUID {
	c := game.Card{
		InstanceID: uuid.New(),
		OracleID:   "omen-enum-oracle",
		Layout:     game.LayoutAdventure,
		Owner:      p.ID,
		Controller: p.ID,
		Faces: []game.Face{
			{Name: "Enum Dragon", TypeLine: "Creature — Dragon", ManaCost: "{0}", Power: 4, Toughness: 4},
			{Name: "Enum Portent", TypeLine: "Sorcery — Omen", ManaCost: "{0}"},
		},
	}
	c.SetFace(0)
	p.Hand.PushTop(c)
	return c.InstanceID
}

func TestEnumeratorOffersBothOmenHalvesFromHandAndNothingAfterward(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)

	id := omenInHand(seat)
	moves := legal.EnumerateFor(g, seat.ID)
	got := castMovesFor(moves, id)
	seen := map[int]bool{}
	for _, m := range got {
		seen[faceOf(t, m)] = true
	}
	if len(got) != 2 || !seen[0] || !seen[1] {
		t.Fatalf("omen card offered faces %v (%d casts), want the Dragon (0) and the Omen (1)", seen, len(got))
	}

	if err := g.CastSpell(seat.ID, id, game.CastSpellParams{Face: 1}); err != nil {
		t.Fatalf("cast the Omen half: %v", err)
	}
	for i := 0; i < 16 && g.Stack.Contains(id); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !seat.Library.Contains(id) {
		t.Fatal("the resolved Omen is not in its owner's library")
	}
	advanceTo(t, g, game.StepPostcombatMain)
	if got := castMovesFor(legal.EnumerateFor(g, seat.ID), id); len(got) != 0 {
		t.Errorf("a shuffled-away Omen card is still offered %d casts", len(got))
	}
}
