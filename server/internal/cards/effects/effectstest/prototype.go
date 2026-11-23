package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// prototype_test.go — CR 718 / CR 702.160, prototype (#5656). The rules
// side is game/prototype.go; these drive it through real casts of the
// eleven Brothers' War cards.

// prototypeCard builds a prototype card the way deck.toGameCard does:
// the printed card, the inset as a second face, then face 0.
func PrototypeCard(t *testing.T, owner uuid.UUID, oracleID, name, typeLine, cost string, power, toughness int, text string) game.Card {
	t.Helper()
	faces := game.PrototypeFaces(game.Face{
		Name:       name,
		TypeLine:   typeLine,
		ManaCost:   cost,
		Power:      power,
		Toughness:  toughness,
		OracleText: text,
	})
	if len(faces) != 2 {
		t.Fatalf("%s: oracle text has no prototype inset", name)
	}
	c := game.Card{
		InstanceID: uuid.New(),
		OracleID:   oracleID,
		Layout:     game.LayoutPrototype,
		Owner:      owner,
		Controller: owner,
		Faces:      faces,
	}
	c.SettleImported()
	return c
}
