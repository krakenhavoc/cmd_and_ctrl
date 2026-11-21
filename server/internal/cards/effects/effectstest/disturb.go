package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	cardsb "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/b"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// disturb_test.go — ADR 0107 §4 (#1855), disturb end to end on real
// cards: the cast out of the graveyard (game/disturb_test.go pins the
// engine half on a fixture), the back face's "If [this] would be put
// into a graveyard from anywhere, exile it instead" (DisturbedExile),
// and the cards that need more than the cast.
//
// Every card here goes down the import road (deck.ToGameCard), for the
// reason siege_transform_test.go gives: a disturb card's whole point is
// its back face, which only the importer puts on game.Card.Faces.

// disturbFace is one printed face of a disturb row.
type DisturbFace struct {
	Name, TypeLine, Cost, Power, Toughness string
}

// disturbRow is the Scryfall record of a disturb card in the shape the
// bulk dump ships every `transform` card: a null top-level cost and the
// real data on the faces.
func DisturbRow(oracleID string, colors []string, front, back DisturbFace) cards.Card {
	face := func(f DisturbFace) cards.CardFace {
		return cards.CardFace{Name: f.Name, TypeLine: f.TypeLine, ManaCost: f.Cost, Power: f.Power, Toughness: f.Toughness, Colors: colors}
	}
	return cards.Card{
		ID:        uuid.New(),
		Name:      front.Name + " // " + back.Name,
		Layout:    "transform",
		TypeLine:  front.TypeLine + " // " + back.TypeLine,
		OracleID:  uuid.MustParse(oracleID),
		CardFaces: []cards.CardFace{face(front), face(back)},
	}
}

func BaithookAnglerRow() cards.Card {
	return DisturbRow(cardsb.BaithookAnglerOracleID, []string{"U"},
		DisturbFace{"Baithook Angler", "Creature — Human Peasant", "{1}{U}", "2", "1"},
		DisturbFace{"Hook-Haunt Drifter", "Creature — Spirit", "", "1", "2"})
}

// importToGraveyard runs a row down the import road into a player's
// graveyard, front face up (CR 712.8a), and walks to a main phase.
func ImportToGraveyard(t *testing.T, g *game.Game, row cards.Card, p *game.Player) uuid.UUID {
	t.Helper()
	c := deck.ToGameCard(row, false)
	c.Owner, c.Controller = p.ID, p.ID
	p.Graveyard.PushTop(c)
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	return c.InstanceID
}

// cardWhere returns the zone a card is in and the card itself.
func CardWhere(g *game.Game, p *game.Player, id uuid.UUID) (string, game.Card) {
	for name, z := range map[string]*game.Zone{
		"battlefield": g.Battlefield, "stack": g.Stack, "exile": g.Exile,
		"graveyard": p.Graveyard, "hand": p.Hand, "library": p.Library,
	} {
		for _, c := range z.Cards {
			if c.InstanceID == id {
				return name, c
			}
		}
	}
	return "", game.Card{}
}
