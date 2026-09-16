package deck

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// CR 118.6 end to end, from the Scryfall record: Ancestral Vision has
// no mana cost, so the importer leaves ManaCost empty, and the engine
// must not read that as a free {0}. Ornithopter prints {0} and still
// casts. Lives in the deck package for the printed_keywords_test.go
// reason: a hand-built game.Card would not prove what the importer
// stamps.
func TestNoManaCostImportedCardIsRefusedFromHand(t *testing.T) {
	vision := cards.Card{
		ID:       uuid.New(),
		OracleID: uuid.New(),
		Name:     "Ancestral Vision",
		Layout:   "normal",
		TypeLine: "Sorcery",
		// Scryfall's mana_cost is "" for a card with no mana cost.
		OracleText: "Suspend 4—{U}\nTarget player draws three cards.",
	}
	thopter := cards.Card{
		ID:       uuid.New(),
		OracleID: uuid.New(),
		Name:     "Ornithopter",
		Layout:   "normal",
		TypeLine: "Artifact Creature — Thopter",
		ManaCost: "{0}",
	}
	list := &List{Mainboard: []cards.Card{vision, thopter}}
	deck := list.ToGameCards()
	visionID, thopterID := deck[0].InstanceID, deck[1].InstanceID
	if !game.HasNoManaCost(deck[0]) {
		t.Fatalf("imported Ancestral Vision does not read as having no mana cost: %+v", deck[0])
	}
	if game.HasNoManaCost(deck[1]) {
		t.Fatalf("imported Ornithopter reads as having no mana cost")
	}

	g, p := startedGame(t, deck)
	moveTo(t, g, p, visionID, game.ZoneRef{Kind: game.ZoneHand, Owner: p.ID})
	moveTo(t, g, p, thopterID, game.ZoneRef{Kind: game.ZoneHand, Owner: p.ID})
	g.Turn = game.Turn{
		Number:         3,
		ActiveSeat:     0,
		PriorityHolder: 0,
		Phase:          game.PhasePrecombatMain,
		Step:           game.StepPrecombatMain,
	}

	for _, strict := range []bool{true, false} {
		err := g.CastSpell(p.ID, visionID, game.CastSpellParams{FromZone: "hand", Strict: strict})
		if !errors.Is(err, game.ErrNoManaCost) {
			t.Errorf("Ancestral Vision from hand (strict=%v): got %v, want ErrNoManaCost", strict, err)
		}
	}
	if err := g.CastSpell(p.ID, thopterID, game.CastSpellParams{FromZone: "hand", Strict: true}); err != nil {
		t.Errorf("Ornithopter from hand: %v", err)
	}
}
