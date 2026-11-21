package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// vowB15Row is a Scryfall record for a transform card: front and back faces.
func VowB15Row(oracle, front, frontType, frontCost, fp, ft, back, backType, bp, bt string, colors []string) cards.Card {
	return cards.Card{
		ID:       uuid.New(),
		Name:     front + " // " + back,
		Layout:   "transform",
		TypeLine: frontType + " // " + backType,
		OracleID: uuid.MustParse(oracle),
		CardFaces: []cards.CardFace{
			{Name: front, TypeLine: frontType, ManaCost: frontCost, Colors: colors, Power: fp, Toughness: ft},
			{Name: back, TypeLine: backType, Colors: colors, Power: bp, Toughness: bt},
		},
	}
}

// vowB15Put puts the row's card onto p's battlefield, front face up, not
// summoning sick.
func VowB15Put(t *testing.T, g *game.Game, row cards.Card, p *game.Player) uuid.UUID {
	t.Helper()
	id := ImportToHand(row, p)
	c, ok := CardInZone(p.Hand, id)
	if !ok {
		t.Fatalf("%s did not reach the hand", row.Name)
	}
	p.Hand.Remove(id)
	PushBattlefieldCardWithTimestamp(g, c)
	return id
}
