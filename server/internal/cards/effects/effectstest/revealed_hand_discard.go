package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ADR 0116 (#2078): the revealed-hand pick's card filter, through the
// three cards that ship on it. One test per rule interaction the ADR
// lists under Tests.

const (
	ThoughtseizeOracle     = "edd8d1e8-be43-4c38-bb3a-83081fbaf0b5"
	UnmaskOracle           = "1a3a274d-41da-47db-ad41-5720f36a7963"
	PelakkaPredationOracle = "b0fd6889-20b4-439b-aa97-2e90aca1675a"
)

// revealHand replaces p's hand with cards and returns their IDs in
// hand order.
func RevealHand(p *game.Player, cards ...game.Card) []uuid.UUID {
	p.Hand.Cards = nil
	ids := make([]uuid.UUID, 0, len(cards))
	for _, c := range cards {
		if c.InstanceID == uuid.Nil {
			c.InstanceID = uuid.New()
		}
		c.Owner, c.Controller = p.ID, p.ID
		p.Hand.PushTop(c)
		ids = append(ids, c.InstanceID)
	}
	return ids
}

func RhForest() game.Card {
	return game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"}
}

func RhBolt() game.Card {
	return game.Card{Name: "Lightning Bolt", TypeLine: "Instant", ManaCost: "{R}", Colors: []string{"R"}}
}

// rhPelakka is Pelakka Predation in a hand, built the way the importer
// builds it: two faces, face 0 materialised (CR 712.8a).
func RhPelakka() game.Card {
	return MdfcCard(uuid.Nil, PelakkaPredationOracle, game.LayoutModalDFC,
		game.Face{Name: "Pelakka Predation", TypeLine: "Sorcery", ManaCost: "{2}{B}", Colors: []string{"B"}},
		game.Face{Name: "Pelakka Caverns", TypeLine: "Land", Colors: nil},
	)
}

func RhFireball() game.Card {
	return game.Card{Name: "Fireball", TypeLine: "Sorcery", ManaCost: "{X}{R}", Colors: []string{"R"}}
}

// openPick returns the one revealed-hand prompt, or fails.
func OpenPick(t *testing.T, g *game.Game) *game.PendingChoice {
	t.Helper()
	var found *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceDiscardFromHand {
			if found != nil {
				t.Fatal("more than one revealed-hand prompt is open")
			}
			found = c
		}
	}
	if found == nil {
		t.Fatal("no revealed-hand prompt is open")
	}
	return found
}
